package middleware

import (
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

// NewAccountTradeLimit 30일 미만 & 전화 미인증 계정의 거래 제한
// 마켓플레이스 거래 생성 엔드포인트에 적용
func NewAccountTradeLimit(db *sqlx.DB) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID := MustGetUserID(c)

			var info struct {
				AgeDays       int  `db:"account_age_days"`
				PhoneVerified bool `db:"phone_verified"`
			}
			err := db.GetContext(c.Request().Context(), &info,
				`SELECT COALESCE(account_age_days, 0) as account_age_days, phone_verified FROM users WHERE id=$1`,
				userID,
			)
			if err != nil {
				return next(c) // DB 에러 시 통과 (fail-open)
			}

			// 30일 이상 또는 전화 인증 완료 시 제한 없음
			if info.AgeDays >= 30 || info.PhoneVerified {
				return next(c)
			}

			// 신규계정: 안전결제(에스크로) 전용으로 제한
			c.Set("new_account_limited", true)
			c.Set("daily_limit_krw", float64(30000))

			return next(c)
		}
	}
}
