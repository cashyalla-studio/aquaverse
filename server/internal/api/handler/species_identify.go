package handler

import (
	"encoding/base64"
	"io"
	"net/http"
	"strings"

	"github.com/cashyalla/aquaverse/internal/api/middleware"
	"github.com/cashyalla/aquaverse/internal/service"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type SpeciesIdentifyHandler struct {
	svc *service.SpeciesIdentifyService
}

func NewSpeciesIdentifyHandler(svc *service.SpeciesIdentifyService) *SpeciesIdentifyHandler {
	return &SpeciesIdentifyHandler{svc: svc}
}

// Identify POST /api/v1/species/identify
// multipart/form-data: image 파일 또는 JSON: {image_base64, media_type}
func (h *SpeciesIdentifyHandler) Identify(c echo.Context) error {
	var base64Image, mediaType string

	contentType := c.Request().Header.Get("Content-Type")
	if strings.Contains(contentType, "multipart") {
		// 파일 업로드
		file, err := c.FormFile("image")
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "image file required"})
		}
		if file.Size > 5*1024*1024 { // 5MB 제한
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "image too large (max 5MB)"})
		}

		src, err := file.Open()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "file open error"})
		}
		defer src.Close()

		data, err := io.ReadAll(src)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "file read error"})
		}

		base64Image = base64.StdEncoding.EncodeToString(data)
		mediaType = file.Header.Get("Content-Type")
		if mediaType == "" {
			mediaType = "image/jpeg"
		}
	} else {
		// JSON body
		var req struct {
			ImageBase64 string `json:"image_base64"`
			MediaType   string `json:"media_type"`
		}
		if err := c.Bind(&req); err != nil || req.ImageBase64 == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "image_base64 required"})
		}
		base64Image = req.ImageBase64
		mediaType = req.MediaType
		if mediaType == "" {
			mediaType = "image/jpeg"
		}
	}

	// 선택적 인증 — 로그인한 경우 user_id 기록
	var userID *uuid.UUID
	if id, err := middleware.GetUserID(c); err == nil {
		userID = &id
	}

	result, err := h.svc.IdentifyFromBase64(c.Request().Context(), userID, base64Image, mediaType)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, result)
}
