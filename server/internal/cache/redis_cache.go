package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/cashyalla/aquaverse/internal/service"
	"github.com/redis/go-redis/v9"
)

const (
	ttlFishDetail = 30 * time.Minute
	ttlFishList   = 5 * time.Minute
	ttlBoards     = 1 * time.Hour
)

type RedisCache struct {
	rdb *redis.Client
}

func NewRedisCache(rdb *redis.Client) *RedisCache {
	return &RedisCache{rdb: rdb}
}

// ── FishCache ──────────────────────────────────────────

func (c *RedisCache) GetFish(ctx context.Context, key string) (*domain.FishData, error) {
	val, err := c.rdb.Get(ctx, "av:"+key).Bytes()
	if err != nil {
		return nil, err
	}
	var fish domain.FishData
	if err := json.Unmarshal(val, &fish); err != nil {
		return nil, err
	}
	return &fish, nil
}

func (c *RedisCache) SetFish(ctx context.Context, key string, fish *domain.FishData) error {
	b, err := json.Marshal(fish)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, "av:"+key, b, ttlFishDetail).Err()
}

func (c *RedisCache) GetFishList(ctx context.Context, key string) (*domain.FishListResult, error) {
	val, err := c.rdb.Get(ctx, "av:"+key).Bytes()
	if err != nil {
		return nil, err
	}
	var result domain.FishListResult
	if err := json.Unmarshal(val, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *RedisCache) SetFishList(ctx context.Context, key string, result *domain.FishListResult) error {
	b, err := json.Marshal(result)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, "av:"+key, b, ttlFishList).Err()
}

func (c *RedisCache) InvalidateFish(ctx context.Context, id int64) error {
	pattern := fmt.Sprintf("av:fish:%d:*", id)
	return c.deleteByPattern(ctx, pattern)
}

// ── 커뮤니티 캐시 ──────────────────────────────────────

func (c *RedisCache) GetBoards(ctx context.Context, locale string) ([]domain.Board, error) {
	val, err := c.rdb.Get(ctx, fmt.Sprintf("av:boards:%s", locale)).Bytes()
	if err != nil {
		return nil, err
	}
	var boards []domain.Board
	return boards, json.Unmarshal(val, &boards)
}

func (c *RedisCache) SetBoards(ctx context.Context, locale string, boards []domain.Board) error {
	b, _ := json.Marshal(boards)
	return c.rdb.Set(ctx, fmt.Sprintf("av:boards:%s", locale), b, ttlBoards).Err()
}

// ── 마켓플레이스 캐시 ───────────────────────────────────

func (c *RedisCache) GetListings(ctx context.Context, key string) (*domain.ListingListResult, error) {
	val, err := c.rdb.Get(ctx, "av:listings:"+key).Bytes()
	if err != nil {
		return nil, err
	}
	var result domain.ListingListResult
	return &result, json.Unmarshal(val, &result)
}

func (c *RedisCache) SetListings(ctx context.Context, key string, result *domain.ListingListResult) error {
	b, _ := json.Marshal(result)
	return c.rdb.Set(ctx, "av:listings:"+key, b, 2*time.Minute).Err()
}

// ── 세션/Rate Limit ────────────────────────────────────

// IncrRateLimit 분당 요청 횟수 증가 (인증 rate limit)
func (c *RedisCache) IncrRateLimit(ctx context.Context, key string, window time.Duration) (int64, error) {
	pipe := c.rdb.Pipeline()
	incr := pipe.Incr(ctx, "av:rl:"+key)
	pipe.Expire(ctx, "av:rl:"+key, window)
	_, err := pipe.Exec(ctx)
	return incr.Val(), err
}

// SetEmailVerifyCode 이메일 인증 코드 저장 (5분 TTL)
func (c *RedisCache) SetEmailVerifyCode(ctx context.Context, email, code string) error {
	return c.rdb.Set(ctx, fmt.Sprintf("av:email_verify:%s", email), code, 5*time.Minute).Err()
}

func (c *RedisCache) GetEmailVerifyCode(ctx context.Context, email string) (string, error) {
	return c.rdb.Get(ctx, fmt.Sprintf("av:email_verify:%s", email)).Result()
}

func (c *RedisCache) DeleteEmailVerifyCode(ctx context.Context, email string) error {
	return c.rdb.Del(ctx, fmt.Sprintf("av:email_verify:%s", email)).Err()
}

// ── 알림 큐 ───────────────────────────────────────────

func (c *RedisCache) PushNotifyJob(ctx context.Context, job []byte) error {
	return c.rdb.LPush(ctx, "av:notify:queue", job).Err()
}

func (c *RedisCache) PopNotifyJob(ctx context.Context, timeout time.Duration) ([]byte, error) {
	result, err := c.rdb.BRPop(ctx, timeout, "av:notify:queue").Result()
	if err != nil {
		return nil, err
	}
	if len(result) < 2 {
		return nil, nil
	}
	return []byte(result[1]), nil
}

// ── 사기 방지 캐시 ─────────────────────────────────────

// GetRecentListingCount 24시간 내 분양글 수 조회
func (c *RedisCache) GetRecentListingCount(ctx context.Context, sellerID string) (int64, error) {
	key := fmt.Sprintf("av:listing_count:%s", sellerID)
	count, err := c.rdb.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return count, err
}

// IncrRecentListingCount 분양글 수 증가 (24시간 TTL)
func (c *RedisCache) IncrRecentListingCount(ctx context.Context, sellerID string) error {
	key := fmt.Sprintf("av:listing_count:%s", sellerID)
	pipe := c.rdb.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, 24*time.Hour)
	_, err := pipe.Exec(ctx)
	return err
}

// ── 내부 유틸 ──────────────────────────────────────────

func (c *RedisCache) deleteByPattern(ctx context.Context, pattern string) error {
	var cursor uint64
	for {
		keys, nextCursor, err := c.rdb.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			c.rdb.Del(ctx, keys...)
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}
