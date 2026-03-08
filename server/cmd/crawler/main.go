// cmd/crawler/main.go - AquaVerse 독립 크롤러 CLI 바이너리
//
// 환경변수:
//   AV_DATABASE_DSN    - PostgreSQL DSN (필수)
//   ANTHROPIC_API_KEY  - Claude API 키 (선택: 미설정 시 번역/AI 강화 스킵)
//   CRAWLER_SOURCE     - 실행할 소스 (fishbase|gbif|wikidata|all, 기본값: all)
//   CRAWLER_LIMIT      - 번역 대상 어종 상한 (기본값: 0 = 무제한)
//
// 사용 예시:
//   docker run --rm -e AV_DATABASE_DSN=... aquaverse-crawler
//   CRAWLER_SOURCE=fishbase go run ./cmd/crawler
package main

import (
	"context"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cashyalla/aquaverse/internal/crawler"
	"github.com/cashyalla/aquaverse/internal/pipeline"
	"github.com/cashyalla/aquaverse/internal/repository"
	"github.com/jmoiron/sqlx"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	// ── 로거 ──────────────────────────────────────────────
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// ── 환경변수 로드 ──────────────────────────────────────
	dsn := os.Getenv("AV_DATABASE_DSN")
	if dsn == "" {
		slog.Error("AV_DATABASE_DSN is required")
		os.Exit(1)
	}

	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	crawlerSource := strings.ToLower(os.Getenv("CRAWLER_SOURCE"))
	if crawlerSource == "" {
		crawlerSource = "all"
	}

	crawlerLimitStr := os.Getenv("CRAWLER_LIMIT")
	crawlerLimit := 0
	if crawlerLimitStr != "" {
		if n, err := strconv.Atoi(crawlerLimitStr); err == nil && n > 0 {
			crawlerLimit = n
		}
	}

	aiModel := os.Getenv("AI_MODEL")
	if aiModel == "" {
		aiModel = "claude-haiku-4-5-20251001"
	}

	userAgent := os.Getenv("CRAWLER_USER_AGENT")
	if userAgent == "" {
		userAgent = "AquaVerse-Crawler/1.0 (contact@aquaverse.app)"
	}

	slog.Info("AquaVerse crawler starting",
		"source", crawlerSource,
		"limit", crawlerLimit,
		"ai_enabled", anthropicKey != "",
	)

	// ── PostgreSQL 연결 ────────────────────────────────────
	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		slog.Error("db connect failed", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(3)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, pingCancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := db.PingContext(ctx); err != nil {
		slog.Error("db ping failed", "err", err)
		pingCancel()
		os.Exit(1)
	}
	pingCancel()
	slog.Info("PostgreSQL connected")

	// ── Repository 초기화 ──────────────────────────────────
	fishRepo := repository.NewFishRepository(db)

	// ── AI 컴포넌트 초기화 ────────────────────────────────
	aiEnricher := pipeline.NewAIEnricher(anthropicKey, aiModel, 2048, logger)
	translator := pipeline.NewTranslator(anthropicKey, aiModel, logger)

	// ── PipelineProcessor 초기화 ──────────────────────────
	processor := pipeline.NewPipelineProcessor(fishRepo, aiEnricher, logger)

	// ── SourceAdapter 목록 구성 ───────────────────────────
	allAdapters := []crawler.SourceAdapter{
		crawler.NewFishBaseAdapter(logger, userAgent),
		crawler.NewGBIFAdapter(logger, userAgent),
		crawler.NewWikidataAdapter(logger, userAgent),
	}

	// CRAWLER_SOURCE 필터링
	var adapters []crawler.SourceAdapter
	switch crawlerSource {
	case "all":
		adapters = allAdapters
	default:
		for _, a := range allAdapters {
			if a.Name() == crawlerSource {
				adapters = append(adapters, a)
				break
			}
		}
		if len(adapters) == 0 {
			slog.Error("unknown CRAWLER_SOURCE",
				"source", crawlerSource,
				"valid", "fishbase|gbif|wikidata|all")
			os.Exit(1)
		}
	}

	// ── Scheduler 초기화 ──────────────────────────────────
	sched := crawler.NewScheduler(fishRepo, processor, adapters, logger)

	appCtx := context.Background()

	// ── Step 1: 크롤링 실행 ────────────────────────────────
	slog.Info("step 1/4: crawling sources")
	if crawlerSource == "all" {
		if err := sched.RunAll(appCtx); err != nil {
			slog.Error("RunAll failed", "err", err)
			// 실패해도 이미 수집된 데이터를 파이프라인으로 처리 계속
		}
	} else {
		if err := sched.RunAdapter(appCtx, crawlerSource); err != nil {
			slog.Error("RunAdapter failed", "source", crawlerSource, "err", err)
		}
		// RunAdapter는 ProcessPending을 자동으로 호출하지 않으므로 직접 실행
	}

	// ── Step 2: PENDING 데이터 파이프라인 처리 ──────────────
	slog.Info("step 2/4: processing pending crawl data")
	if err := processor.ProcessPending(appCtx); err != nil {
		slog.Error("ProcessPending failed", "err", err)
	}

	// ── Step 3: 저품질 데이터 AI 보완 ────────────────────────
	slog.Info("step 3/4: enriching low quality data")
	if err := processor.EnrichLowQuality(appCtx); err != nil {
		slog.Error("EnrichLowQuality failed", "err", err)
	}

	// ── Step 4: 다국어 번역 ───────────────────────────────
	slog.Info("step 4/4: translating missing locales")
	if anthropicKey == "" {
		slog.Info("ANTHROPIC_API_KEY not set, skipping translation step")
	} else {
		if err := translateMissingLocales(appCtx, fishRepo, translator, logger, crawlerLimit); err != nil {
			slog.Error("translateMissingLocales failed", "err", err)
		}
	}

	// ── 상태 출력 ─────────────────────────────────────────
	status, err := sched.Status(appCtx)
	if err != nil {
		slog.Warn("failed to get scheduler status", "err", err)
	} else {
		slog.Info("crawler run completed", "status", status)
	}

	slog.Info("AquaVerse crawler finished")
}

// translateMissingLocales PUBLISHED 어종 중 번역이 없는 로케일을 찾아 번역한다.
// crawlerLimit > 0 이면 처리 어종 수를 제한한다.
func translateMissingLocales(
	ctx context.Context,
	fishRepo *repository.FishRepository,
	translator *pipeline.Translator,
	logger *slog.Logger,
	crawlerLimit int,
) error {
	batchSize := 20
	offset := 0
	totalTranslated := 0
	totalFish := 0

	for {
		// CRAWLER_LIMIT 상한 체크
		if crawlerLimit > 0 && totalFish >= crawlerLimit {
			logger.Info("translation limit reached", "limit", crawlerLimit)
			break
		}

		// 이번 배치에서 가져올 건수 계산
		fetchLimit := batchSize
		if crawlerLimit > 0 {
			remaining := crawlerLimit - totalFish
			if remaining < fetchLimit {
				fetchLimit = remaining
			}
		}

		fishes, err := fishRepo.ListPublishedFish(ctx, fetchLimit, offset)
		if err != nil {
			return err
		}
		if len(fishes) == 0 {
			break
		}

		for i := range fishes {
			fish := &fishes[i]

			existingLocales, err := fishRepo.ListExistingTranslationLocales(ctx, fish.ID)
			if err != nil {
				logger.Warn("failed to list existing locales",
					"fish_id", fish.ID,
					"err", err)
				continue
			}

			translations, err := translator.TranslateAllMissing(ctx, fish.ID, fish, existingLocales)
			if err != nil {
				logger.Warn("TranslateAllMissing failed",
					"fish", fish.ScientificName,
					"err", err)
				continue
			}

			for j := range translations {
				if err := fishRepo.SaveTranslation(ctx, &translations[j]); err != nil {
					logger.Error("failed to save translation",
						"fish", fish.ScientificName,
						"locale", translations[j].Locale,
						"err", err)
					continue
				}
				totalTranslated++
				logger.Info("translation saved",
					"fish", fish.ScientificName,
					"locale", translations[j].Locale)
			}

			totalFish++
		}

		offset += len(fishes)

		if len(fishes) < batchSize {
			break
		}
	}

	logger.Info("translation step completed",
		"fish_processed", totalFish,
		"translations_saved", totalTranslated)

	return nil
}
