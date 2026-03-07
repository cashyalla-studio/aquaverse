package crawler

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"

	"github.com/robfig/cron/v3"
)

// Scheduler 크롤링 스케줄러 (coin-trading cron 패턴 재활용)
type Scheduler struct {
	cron          *cron.Cron
	fishbaseClient *FishBaseClient
	processor     DataProcessor
	logger        *slog.Logger
}

// DataProcessor 데이터 가공 파이프라인 인터페이스
type DataProcessor interface {
	ProcessPending(ctx context.Context) error
	EnrichLowQuality(ctx context.Context) error
	TranslatePending(ctx context.Context) error
}

func NewScheduler(
	fishbase *FishBaseClient,
	processor DataProcessor,
	logger *slog.Logger,
) *Scheduler {
	return &Scheduler{
		cron:          cron.New(),
		fishbaseClient: fishbase,
		processor:     processor,
		logger:        logger,
	}
}

func (s *Scheduler) Start() {
	// FishBase: 매주 월요일 02:00 (데이터 변동 적음)
	s.cron.AddFunc("0 2 * * 1", s.crawlFishBase)

	// 데이터 가공 파이프라인: 매일 07:00
	s.cron.AddFunc("0 7 * * *", s.processPending)

	// AI 재가공 (낮은 품질): 매일 08:00
	s.cron.AddFunc("0 8 * * *", s.enrichLowQuality)

	// 번역: 매일 09:00
	s.cron.AddFunc("0 9 * * *", s.translatePending)

	s.cron.Start()
	s.logger.Info("crawler scheduler started")
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
}

func (s *Scheduler) crawlFishBase() {
	ctx := context.Background()
	s.logger.Info("starting fishbase crawl")

	offset := 0
	limit := 100
	total := 0

	for {
		species, err := s.fishbaseClient.ListAquariumSpecies(ctx, offset, limit)
		if err != nil {
			s.logger.Error("fishbase crawl error", "err", err, "offset", offset)
			break
		}
		if len(species) == 0 {
			break
		}

		// TODO: RawCrawlData 저장 (repo 주입 필요)
		total += len(species)
		offset += limit

		if len(species) < limit {
			break
		}
	}

	s.logger.Info("fishbase crawl completed", "total", total)
}

func (s *Scheduler) processPending() {
	if err := s.processor.ProcessPending(context.Background()); err != nil {
		s.logger.Error("process pending error", "err", err)
	}
}

func (s *Scheduler) enrichLowQuality() {
	if err := s.processor.EnrichLowQuality(context.Background()); err != nil {
		s.logger.Error("enrich low quality error", "err", err)
	}
}

func (s *Scheduler) translatePending() {
	if err := s.processor.TranslatePending(context.Background()); err != nil {
		s.logger.Error("translate pending error", "err", err)
	}
}

// hashContent SHA256 해시 (중복 방지)
func hashContent(data []byte) string {
	h := sha256.Sum256(data)
	return fmt.Sprintf("%x", h)
}
