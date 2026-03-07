package handler

import (
	"net/http"

	"github.com/cashyalla/aquaverse/internal/api/middleware"
	"github.com/cashyalla/aquaverse/internal/service"
	"github.com/labstack/echo/v4"
)

type TOTPHandler struct {
	svc *service.TOTPService
}

func NewTOTPHandler(svc *service.TOTPService) *TOTPHandler {
	return &TOTPHandler{svc: svc}
}

// GetStatus GET /api/v1/auth/totp
func (h *TOTPHandler) GetStatus(c echo.Context) error {
	userID := middleware.MustGetUserID(c)
	status, err := h.svc.GetTOTPStatus(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, status)
}

// Enable POST /api/v1/auth/totp/enable — 비밀키 생성 및 QR URL 반환
func (h *TOTPHandler) Enable(c echo.Context) error {
	userID := middleware.MustGetUserID(c)
	secret, qrURL, err := h.svc.EnableTOTP(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{
		"secret":  secret,
		"qr_url":  qrURL,
		"message": "QR 코드를 인증 앱으로 스캔한 후 /verify 엔드포인트로 코드를 확인하세요.",
	})
}

// Verify POST /api/v1/auth/totp/verify — 코드 확인 후 TOTP 활성화
func (h *TOTPHandler) Verify(c echo.Context) error {
	userID := middleware.MustGetUserID(c)
	var req struct {
		Code string `json:"code"`
	}
	if err := c.Bind(&req); err != nil || req.Code == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "code required"})
	}

	backupCodes, err := h.svc.VerifyAndActivateTOTP(c.Request().Context(), userID, req.Code)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":       "enabled",
		"backup_codes": backupCodes,
		"message":      "2단계 인증이 활성화되었습니다. 백업 코드를 안전한 곳에 저장하세요.",
	})
}

// Disable DELETE /api/v1/auth/totp — TOTP 비활성화
func (h *TOTPHandler) Disable(c echo.Context) error {
	userID := middleware.MustGetUserID(c)
	var req struct {
		Code string `json:"code"`
	}
	if err := c.Bind(&req); err != nil || req.Code == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "code required"})
	}

	if err := h.svc.DisableTOTP(c.Request().Context(), userID, req.Code); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"status": "disabled"})
}
