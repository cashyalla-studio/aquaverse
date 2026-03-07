package handler

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/cashyalla/aquaverse/internal/api/middleware"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	miniogo "github.com/minio/minio-go/v7"
)

// UploadHandler MinIO presigned URL 생성 핸들러
type UploadHandler struct {
	minioClient *miniogo.Client
	publicBase  string // 외부 접근 가능한 base URL (예: https://cdn.aquaverse.app)
}

// NewUploadHandler UploadHandler 생성자
// publicBase: presigned URL이 아닌 공개 접근 URL 접두사 (예: "https://cdn.aquaverse.app")
func NewUploadHandler(minioClient *miniogo.Client, publicBase string) *UploadHandler {
	return &UploadHandler{
		minioClient: minioClient,
		publicBase:  publicBase,
	}
}

// PresignRequest presigned URL 요청 파라미터
type PresignRequest struct {
	// bucket: "listings" 또는 "fish" (쿼리 파라미터)
	Bucket    string `query:"bucket"`
	Filename  string `query:"filename"`
}

// PresignResponse presigned URL 응답
type PresignResponse struct {
	UploadURL string `json:"upload_url"`
	PublicURL string `json:"public_url"`
	Key       string `json:"key"`
}

// allowedBuckets 허용된 버킷 목록
var allowedBuckets = map[string]bool{
	"listings": true,
	"fish":     true,
}

// PresignUpload POST /api/v1/upload/presign
// JWT 미들웨어 통과 후 호출됨
func (h *UploadHandler) PresignUpload(c echo.Context) error {
	// 인증된 사용자 확인
	userID, ok := c.Get(middleware.ContextKeyUserID).(string)
	if !ok || userID == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}

	var req PresignRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request parameters")
	}

	// 버킷 검증
	bucket := strings.ToLower(strings.TrimSpace(req.Bucket))
	if bucket == "" {
		bucket = "listings"
	}
	if !allowedBuckets[bucket] {
		return echo.NewHTTPError(http.StatusBadRequest, "bucket must be 'listings' or 'fish'")
	}

	// 파일 확장자 추출 및 검증
	ext := ""
	if req.Filename != "" {
		ext = strings.ToLower(filepath.Ext(req.Filename))
	}
	if ext == "" {
		ext = ".jpg"
	}

	// 허용된 이미지 확장자만 허용
	allowedExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
		".gif":  true,
	}
	if !allowedExts[ext] {
		return echo.NewHTTPError(http.StatusBadRequest, "allowed extensions: jpg, jpeg, png, webp, gif")
	}

	// 키 생성: {bucket}/{uuid}.{ext}
	objectKey := fmt.Sprintf("%s/%s%s", bucket, uuid.New().String(), ext)

	// presigned PUT URL 생성 (유효시간 15분)
	presignedURL, err := h.minioClient.PresignedPutObject(
		c.Request().Context(),
		bucket,
		objectKey,
		15*time.Minute,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate presigned URL")
	}

	// 공개 URL 구성
	publicURL := fmt.Sprintf("%s/%s", strings.TrimRight(h.publicBase, "/"), objectKey)

	return c.JSON(http.StatusOK, PresignResponse{
		UploadURL: presignedURL.String(),
		PublicURL: publicURL,
		Key:       objectKey,
	})
}
