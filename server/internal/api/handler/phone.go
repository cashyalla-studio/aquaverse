package handler

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/cashyalla/aquaverse/internal/api/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
)

type PhoneHandler struct {
	db *sqlx.DB
}

func NewPhoneHandler(db *sqlx.DB) *PhoneHandler {
	return &PhoneHandler{db: db}
}

// SendCode POST /api/v1/phone/send - 인증 코드 발송
func (h *PhoneHandler) SendCode(c echo.Context) error {
	var req struct {
		PhoneNumber string `json:"phone_number"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}
	if len(req.PhoneNumber) < 10 {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid phone number")
	}

	userID := middleware.MustGetUserID(c)
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	expiresAt := time.Now().Add(10 * time.Minute)

	_, err := h.db.ExecContext(c.Request().Context(),
		`INSERT INTO user_phone_verifications (user_id, phone_number, code, expires_at)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT DO NOTHING`,
		userID, req.PhoneNumber, code, expiresAt,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create verification")
	}

	// TODO: 실제 SMS 발송 (CoolSMS / Twilio)
	// 개발환경에서는 코드를 응답에 포함
	resp := map[string]interface{}{
		"message":    "verification code sent",
		"expires_at": expiresAt,
	}
	// 개발환경 전용: 코드 노출
	if c.Echo().Debug {
		resp["dev_code"] = code
	}
	return c.JSON(http.StatusOK, resp)
}

// VerifyCode POST /api/v1/phone/verify - 인증 코드 확인
func (h *PhoneHandler) VerifyCode(c echo.Context) error {
	var req struct {
		PhoneNumber string `json:"phone_number"`
		Code        string `json:"code"`
	}
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	userID := middleware.MustGetUserID(c)

	var stored struct {
		ID         int64     `db:"id"`
		Code       string    `db:"code"`
		IsVerified bool      `db:"is_verified"`
		ExpiresAt  time.Time `db:"expires_at"`
	}
	err := h.db.GetContext(c.Request().Context(), &stored,
		`SELECT id, code, is_verified, expires_at FROM user_phone_verifications
		 WHERE user_id=$1 AND phone_number=$2 ORDER BY created_at DESC LIMIT 1`,
		userID, req.PhoneNumber,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "verification not found")
	}
	if stored.IsVerified {
		return echo.NewHTTPError(http.StatusConflict, "already verified")
	}
	if time.Now().After(stored.ExpiresAt) {
		return echo.NewHTTPError(http.StatusGone, "code expired")
	}
	if stored.Code != req.Code {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid code")
	}

	tx, _ := h.db.BeginTxx(c.Request().Context(), nil)
	tx.ExecContext(c.Request().Context(),
		`UPDATE user_phone_verifications SET is_verified=TRUE, verified_at=NOW() WHERE id=$1`, stored.ID)
	tx.ExecContext(c.Request().Context(),
		`UPDATE users SET phone_number=$1, phone_verified=TRUE WHERE id=$2`, req.PhoneNumber, userID)
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "commit failed")
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "verified"})
}
