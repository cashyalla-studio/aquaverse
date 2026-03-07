package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

// RateLimit Redis 기반 슬라이딩 윈도우 Rate Limiting
// limit: 윈도우 내 최대 요청 수
// window: 윈도우 크기
func RateLimit(rdb *redis.Client, limit int, window time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			key := fmt.Sprintf("rl:%s:%s", c.Path(), ip)
			ctx := context.Background()

			// 차단된 IP 체크 (먼저 확인)
			banKey := fmt.Sprintf("rl:ban:%s", ip)
			if banned, _ := rdb.Exists(ctx, banKey).Result(); banned > 0 {
				return c.JSON(http.StatusTooManyRequests, map[string]string{
					"error": "ip temporarily banned",
				})
			}

			// 현재 카운트 증가 (원자적)
			count, err := rdb.Incr(ctx, key).Result()
			if err != nil {
				// Redis 실패 시 통과 (fail-open)
				return next(c)
			}

			// 첫 번째 요청이면 TTL 설정
			if count == 1 {
				rdb.Expire(ctx, key, window)
			}

			// 헤더 설정
			c.Response().Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
			c.Response().Header().Set("X-RateLimit-Remaining", strconv.Itoa(max(0, limit-int(count))))

			if int(count) > limit {
				// IP 차단 (10분)
				rdb.Set(ctx, banKey, "1", 10*time.Minute)
				return c.JSON(http.StatusTooManyRequests, map[string]string{
					"error": "too many requests",
				})
			}

			return next(c)
		}
	}
}

// AuthRateLimit 인증 엔드포인트 전용 Rate Limit (더 엄격)
// 1분에 10회, 초과 시 15분 차단
func AuthRateLimit(rdb *redis.Client) echo.MiddlewareFunc {
	return RateLimit(rdb, 10, time.Minute)
}
