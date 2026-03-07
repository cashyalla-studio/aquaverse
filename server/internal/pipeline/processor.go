package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/cashyalla/aquaverse/internal/domain"
)

// FishDataRepository PipelineProcessor가 필요로 하는 repository 인터페이스
type FishDataRepository interface {
	ListPendingCrawlData(ctx context.Context, limit int) ([]domain.RawCrawlData, error)
	Save(ctx context.Context, fish *domain.FishData) error
	SaveTranslation(ctx context.Context, t *domain.FishTranslation) error
	ListLowQualityFish(ctx context.Context, maxScore float64, limit int) ([]domain.FishData, error)
	UpdatePublishStatus(ctx context.Context, id int64, status domain.PublishStatus, score float64) error
}

// PipelineProcessor crawler.DataProcessor 인터페이스를 구현하는 실제 파이프라인 처리기
type PipelineProcessor struct {
	fishRepo   FishDataRepository
	aiEnricher *AIEnricher
	logger     *slog.Logger
}

// NewPipelineProcessor PipelineProcessor 생성자
func NewPipelineProcessor(
	fishRepo FishDataRepository,
	aiEnricher *AIEnricher,
	logger *slog.Logger,
) *PipelineProcessor {
	return &PipelineProcessor{
		fishRepo:   fishRepo,
		aiEnricher: aiEnricher,
		logger:     logger,
	}
}

// ProcessPending PENDING 상태의 원시 크롤 데이터를 파싱하고 품질 점수를 산정해 저장한다.
func (p *PipelineProcessor) ProcessPending(ctx context.Context) error {
	items, err := p.fishRepo.ListPendingCrawlData(ctx, 100)
	if err != nil {
		return fmt.Errorf("list pending crawl data: %w", err)
	}

	p.logger.Info("processing pending crawl data", "count", len(items))

	successCount := 0
	for _, raw := range items {
		fish, err := parseCrawlData(&raw)
		if err != nil {
			p.logger.Warn("failed to parse crawl data",
				"id", raw.ID, "source", raw.SourceName, "err", err)
			continue
		}

		scored := ScoreFish(fish)
		fish.QualityScore = scored.Score

		if scored.ShouldReject {
			fish.PublishStatus = domain.PublishStatusRejected
		} else if scored.Score >= QualityGood {
			fish.PublishStatus = domain.PublishStatusPublished
		} else {
			fish.PublishStatus = domain.PublishStatusDraft
		}

		if err := p.fishRepo.Save(ctx, fish); err != nil {
			p.logger.Error("failed to save fish data",
				"scientific_name", fish.ScientificName, "err", err)
			continue
		}

		successCount++
		p.logger.Info("processed fish data",
			"scientific_name", fish.ScientificName,
			"quality_score", fish.QualityScore,
			"status", fish.PublishStatus)
	}

	p.logger.Info("process pending completed",
		"total", len(items),
		"success", successCount)

	return nil
}

// EnrichLowQuality 품질 점수가 낮은 데이터를 AI로 보완한다.
func (p *PipelineProcessor) EnrichLowQuality(ctx context.Context) error {
	items, err := p.fishRepo.ListLowQualityFish(ctx, QualityFair, 50)
	if err != nil {
		return fmt.Errorf("list low quality fish: %w", err)
	}

	p.logger.Info("enriching low quality fish", "count", len(items))

	enrichedCount := 0
	for i := range items {
		fish := &items[i]

		// 현재 품질 점수 산정으로 누락 필드 목록 확보
		scored := ScoreFish(fish)
		if len(scored.MissingFields) == 0 {
			continue
		}

		enriched, err := p.aiEnricher.Enrich(ctx, fish, scored.MissingFields)
		if err != nil {
			p.logger.Warn("AI enrichment failed",
				"fish", fish.ScientificName, "err", err)
			continue
		}

		ApplyEnrichment(fish, enriched)
		fish.AIEnriched = true

		// 재채점
		rescored := ScoreFish(fish)
		fish.QualityScore = rescored.Score

		if rescored.ShouldReject {
			fish.PublishStatus = domain.PublishStatusRejected
		} else if rescored.Score >= QualityGood {
			fish.PublishStatus = domain.PublishStatusPublished
		} else {
			fish.PublishStatus = domain.PublishStatusDraft
		}

		if err := p.fishRepo.UpdatePublishStatus(ctx, fish.ID, fish.PublishStatus, fish.QualityScore); err != nil {
			p.logger.Error("failed to update publish status after enrichment",
				"fish", fish.ScientificName, "err", err)
			continue
		}

		if err := p.fishRepo.Save(ctx, fish); err != nil {
			p.logger.Error("failed to save enriched fish data",
				"fish", fish.ScientificName, "err", err)
			continue
		}

		enrichedCount++
		p.logger.Info("AI enrichment applied",
			"fish", fish.ScientificName,
			"old_score", scored.Score,
			"new_score", rescored.Score)
	}

	p.logger.Info("enrich low quality completed",
		"total", len(items),
		"enriched", enrichedCount)

	return nil
}

// TranslatePending PUBLISHED 상태이지만 번역이 없는 항목을 AI로 번역한다.
func (p *PipelineProcessor) TranslatePending(ctx context.Context) error {
	// 번역이 필요한 항목: PUBLISHED 상태이며 기본 한국어 번역이 없는 항목
	// 현재 FishDataRepository에 번역 대상 목록 조회 메서드가 없으므로
	// 낮은 품질 데이터 중 PUBLISHED 상태인 것을 대상으로 한국어 번역을 시도한다.
	items, err := p.fishRepo.ListLowQualityFish(ctx, QualityExcellent, 30)
	if err != nil {
		return fmt.Errorf("list fish for translation: %w", err)
	}

	translatedCount := 0
	for i := range items {
		fish := &items[i]
		if fish.PublishStatus != domain.PublishStatusPublished {
			continue
		}

		// AI를 활용해 한국어 번역 생성
		missing := []string{"common_name_ko", "care_notes_ko", "diet_notes_ko", "breeding_notes_ko"}
		enriched, err := p.aiEnricher.Enrich(ctx, fish, missing)
		if err != nil {
			p.logger.Warn("translation failed",
				"fish", fish.ScientificName, "err", err)
			continue
		}

		translation := &domain.FishTranslation{
			FishDataID:        fish.ID,
			Locale:            domain.Locale("ko"),
			TranslationSource: "ai",
		}

		if enriched.CareNotes != nil {
			translation.CareNotes = enriched.CareNotes
		}
		if enriched.DietNotes != nil {
			translation.DietNotes = enriched.DietNotes
		}
		if enriched.BreedingNotes != nil {
			translation.BreedingNotes = enriched.BreedingNotes
		}

		if err := p.fishRepo.SaveTranslation(ctx, translation); err != nil {
			p.logger.Error("failed to save translation",
				"fish", fish.ScientificName, "err", err)
			continue
		}

		translatedCount++
		p.logger.Info("translation saved",
			"fish", fish.ScientificName,
			"locale", translation.Locale)
	}

	p.logger.Info("translate pending completed",
		"total", len(items),
		"translated", translatedCount)

	return nil
}

// parseCrawlData RawCrawlData의 raw_json을 FishData로 파싱한다.
func parseCrawlData(raw *domain.RawCrawlData) (*domain.FishData, error) {
	var fish domain.FishData
	if err := json.Unmarshal(raw.RawJSON, &fish); err != nil {
		return nil, fmt.Errorf("unmarshal raw crawl json: %w", err)
	}
	if fish.ScientificName == "" {
		return nil, fmt.Errorf("scientific_name is required")
	}
	if fish.PrimarySource == "" {
		fish.PrimarySource = raw.SourceName
	}
	if fish.SourceURL == nil && raw.SourceURL != nil {
		fish.SourceURL = raw.SourceURL
	}
	return &fish, nil
}
