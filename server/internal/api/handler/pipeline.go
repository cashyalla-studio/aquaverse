package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/cashyalla/aquaverse/internal/repository"
	"github.com/labstack/echo/v4"
)

// PipelineRepo pipeline 핸들러가 필요로 하는 repository 인터페이스
type PipelineRepo interface {
	ListCrawlJobs(ctx context.Context, limit int) ([]domain.CrawlJob, error)
	GetPipelineStats(ctx context.Context) (*repository.PipelineStats, error)
}

// PipelineProcessor pipeline 핸들러가 필요로 하는 프로세서 인터페이스
type PipelineProcessor interface {
	ProcessPending(ctx context.Context) error
	EnrichLowQuality(ctx context.Context) error
	TranslatePending(ctx context.Context) error
}

// PipelineHandler 파이프라인 관리 핸들러
type PipelineHandler struct {
	repo      PipelineRepo
	processor PipelineProcessor
	logger    *slog.Logger
}

// NewPipelineHandler PipelineHandler 생성자
func NewPipelineHandler(repo PipelineRepo, processor PipelineProcessor, logger *slog.Logger) *PipelineHandler {
	return &PipelineHandler{
		repo:      repo,
		processor: processor,
		logger:    logger,
	}
}

// GetStats GET /api/v1/admin/pipeline/stats
func (h *PipelineHandler) GetStats(c echo.Context) error {
	stats, err := h.repo.GetPipelineStats(c.Request().Context())
	if err != nil {
		h.logger.Error("get pipeline stats failed", "err", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, stats)
}

// TriggerProcess POST /api/v1/admin/pipeline/process
// PENDING 원시 크롤 데이터를 처리한다 (비동기).
func (h *PipelineHandler) TriggerProcess(c echo.Context) error {
	go func() {
		if err := h.processor.ProcessPending(context.Background()); err != nil {
			h.logger.Error("pipeline process pending failed", "err", err)
		}
	}()
	return c.JSON(http.StatusOK, map[string]string{"message": "started"})
}

// TriggerEnrich POST /api/v1/admin/pipeline/enrich
// 품질이 낮은 어종 데이터를 AI로 보강한다 (비동기).
func (h *PipelineHandler) TriggerEnrich(c echo.Context) error {
	go func() {
		if err := h.processor.EnrichLowQuality(context.Background()); err != nil {
			h.logger.Error("pipeline enrich low quality failed", "err", err)
		}
	}()
	return c.JSON(http.StatusOK, map[string]string{"message": "started"})
}

// TriggerTranslate POST /api/v1/admin/pipeline/translate
// 번역이 필요한 어종을 AI로 번역한다 (비동기).
func (h *PipelineHandler) TriggerTranslate(c echo.Context) error {
	go func() {
		if err := h.processor.TranslatePending(context.Background()); err != nil {
			h.logger.Error("pipeline translate pending failed", "err", err)
		}
	}()
	return c.JSON(http.StatusOK, map[string]string{"message": "started"})
}

// ListJobs GET /api/v1/admin/pipeline/jobs
// 최근 크롤 잡 20개를 반환한다.
func (h *PipelineHandler) ListJobs(c echo.Context) error {
	jobs, err := h.repo.ListCrawlJobs(c.Request().Context(), 20)
	if err != nil {
		h.logger.Error("list crawl jobs failed", "err", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"jobs": jobs})
}
