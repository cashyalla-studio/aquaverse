package crawler

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"time"

	"github.com/cashyalla/aquaverse/internal/domain"
	"github.com/robfig/cron/v3"
)

// DataProcessor 데이터 가공 파이프라인 인터페이스
type DataProcessor interface {
	ProcessPending(ctx context.Context) error
	EnrichLowQuality(ctx context.Context) error
	TranslatePending(ctx context.Context) error
}

// CrawlerRepository 스케줄러가 필요로 하는 repository 인터페이스
type CrawlerRepository interface {
	SaveRawCrawlData(ctx context.Context, data *domain.RawCrawlData) error
	CreateCrawlJob(ctx context.Context, job *domain.CrawlJob) error
	UpdateCrawlJob(ctx context.Context, job *domain.CrawlJob) error
	ListCrawlJobs(ctx context.Context, limit int) ([]domain.CrawlJob, error)
}

// Scheduler 크롤링 스케줄러 (CrawlJob DB 레코드 관리 + 어댑터 실행)
type Scheduler struct {
	cron      *cron.Cron
	repo      CrawlerRepository
	processor DataProcessor
	adapters  []SourceAdapter
	logger    *slog.Logger
}

// NewScheduler Scheduler 생성자.
// repo가 nil이면 CrawlJob DB 기록 없이 어댑터만 실행한다 (하위 호환).
func NewScheduler(
	repo CrawlerRepository,
	processor DataProcessor,
	adapters []SourceAdapter,
	logger *slog.Logger,
) *Scheduler {
	return &Scheduler{
		cron:      cron.New(),
		repo:      repo,
		processor: processor,
		adapters:  adapters,
		logger:    logger,
	}
}

// NewSchedulerLegacy cmd/server에서 기존 FishBaseClient 기반 스케줄러를 그대로 사용할 수 있도록
// 하위 호환 생성자를 제공한다.
func NewSchedulerLegacy(
	fishbase *FishBaseClient,
	processor DataProcessor,
	logger *slog.Logger,
) *Scheduler {
	return &Scheduler{
		cron:      cron.New(),
		repo:      nil,
		processor: processor,
		adapters:  nil,
		logger:    logger,
	}
}

// Start 크론 스케줄 등록 및 시작
func (s *Scheduler) Start() {
	// 어댑터 크롤: 매주 월요일 02:00
	s.cron.AddFunc("0 2 * * 1", func() {
		if err := s.RunAll(context.Background()); err != nil {
			s.logger.Error("scheduled RunAll failed", "err", err)
		}
	})

	// 데이터 가공: 매일 07:00
	s.cron.AddFunc("0 7 * * *", func() {
		if err := s.processor.ProcessPending(context.Background()); err != nil {
			s.logger.Error("scheduled ProcessPending failed", "err", err)
		}
	})

	// AI 재가공: 매일 08:00
	s.cron.AddFunc("0 8 * * *", func() {
		if err := s.processor.EnrichLowQuality(context.Background()); err != nil {
			s.logger.Error("scheduled EnrichLowQuality failed", "err", err)
		}
	})

	// 번역: 매일 09:00
	s.cron.AddFunc("0 9 * * *", func() {
		if err := s.processor.TranslatePending(context.Background()); err != nil {
			s.logger.Error("scheduled TranslatePending failed", "err", err)
		}
	})

	s.cron.Start()
	s.logger.Info("crawler scheduler started")
}

// Stop 크론 스케줄러 정지
func (s *Scheduler) Stop() {
	s.cron.Stop()
}

// RunAdapter 특정 이름의 어댑터를 찾아 실행한다.
// CrawlJob DB 레코드를 생성하고 완료/실패 상태로 업데이트한다.
func (s *Scheduler) RunAdapter(ctx context.Context, adapterName string) error {
	var target SourceAdapter
	for _, a := range s.adapters {
		if a.Name() == adapterName {
			target = a
			break
		}
	}
	if target == nil {
		return fmt.Errorf("adapter %q not found", adapterName)
	}

	return s.runSingleAdapter(ctx, target)
}

// RunAll 모든 어댑터를 순차적으로 실행한 뒤 ProcessPending을 호출한다.
func (s *Scheduler) RunAll(ctx context.Context) error {
	s.logger.Info("RunAll started", "adapter_count", len(s.adapters))

	for _, adapter := range s.adapters {
		if err := s.runSingleAdapter(ctx, adapter); err != nil {
			// 한 어댑터가 실패해도 나머지는 계속 실행
			s.logger.Error("adapter run failed",
				"adapter", adapter.Name(),
				"err", err)
		}
	}

	s.logger.Info("all adapters completed, starting ProcessPending")
	if err := s.processor.ProcessPending(ctx); err != nil {
		return fmt.Errorf("ProcessPending after RunAll: %w", err)
	}

	return nil
}

// Status 현재 크롤 작업 상태를 반환한다.
func (s *Scheduler) Status(ctx context.Context) (map[string]interface{}, error) {
	status := map[string]interface{}{
		"adapters": make([]string, 0, len(s.adapters)),
	}

	adapterNames := make([]string, 0, len(s.adapters))
	for _, a := range s.adapters {
		adapterNames = append(adapterNames, a.Name())
	}
	status["adapters"] = adapterNames

	if s.repo == nil {
		status["crawl_jobs"] = []interface{}{}
		return status, nil
	}

	jobs, err := s.repo.ListCrawlJobs(ctx, 10)
	if err != nil {
		return nil, fmt.Errorf("list crawl jobs: %w", err)
	}

	jobList := make([]map[string]interface{}, 0, len(jobs))
	for _, j := range jobs {
		entry := map[string]interface{}{
			"id":              j.ID,
			"source_name":     j.SourceName,
			"status":          j.Status,
			"items_found":     j.ItemsFound,
			"items_processed": j.ItemsProcessed,
			"items_failed":    j.ItemsFailed,
			"created_at":      j.CreatedAt,
		}
		if j.StartedAt != nil {
			entry["started_at"] = j.StartedAt
		}
		if j.CompletedAt != nil {
			entry["completed_at"] = j.CompletedAt
		}
		if j.ErrorMessage != nil {
			entry["error_message"] = j.ErrorMessage
		}
		jobList = append(jobList, entry)
	}
	status["crawl_jobs"] = jobList

	return status, nil
}

// runSingleAdapter 단일 어댑터를 실행하고 CrawlJob 레코드를 관리한다.
func (s *Scheduler) runSingleAdapter(ctx context.Context, adapter SourceAdapter) error {
	s.logger.Info("starting adapter", "adapter", adapter.Name())

	job := &domain.CrawlJob{
		SourceName: adapter.Name(),
		SourceURL:  "",
		JobType:    "FULL",
		Status:     "RUNNING",
	}

	now := time.Now()
	job.StartedAt = &now

	// DB 레코드 생성 (repo 있을 때만)
	if s.repo != nil {
		if err := s.repo.CreateCrawlJob(ctx, job); err != nil {
			s.logger.Warn("failed to create crawl job record",
				"adapter", adapter.Name(),
				"err", err)
			// DB 기록 실패해도 크롤은 계속 진행
		}
	}

	// save 콜백: RawCrawlData를 DB에 저장
	saveFn := func(data *domain.RawCrawlData) error {
		if data.CrawlJobID == nil && job.ID > 0 {
			data.CrawlJobID = &job.ID
		}
		if s.repo != nil {
			return s.repo.SaveRawCrawlData(ctx, data)
		}
		return nil
	}

	fetched, crawlErr := adapter.Crawl(ctx, job.ID, saveFn)

	completedAt := time.Now()
	job.CompletedAt = &completedAt
	job.ItemsFound = fetched

	if crawlErr != nil {
		job.Status = "FAILED"
		msg := crawlErr.Error()
		job.ErrorMessage = &msg
		s.logger.Error("adapter crawl failed",
			"adapter", adapter.Name(),
			"fetched", fetched,
			"err", crawlErr)
	} else {
		job.Status = "COMPLETED"
		job.ItemsProcessed = fetched
		s.logger.Info("adapter crawl completed",
			"adapter", adapter.Name(),
			"fetched", fetched)
	}

	if s.repo != nil && job.ID > 0 {
		if err := s.repo.UpdateCrawlJob(ctx, job); err != nil {
			s.logger.Warn("failed to update crawl job record",
				"adapter", adapter.Name(),
				"job_id", job.ID,
				"err", err)
		}
	}

	return crawlErr
}

// hashContent SHA256 해시 (중복 방지)
func hashContent(data []byte) string {
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h)
}
