package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cashyalla/aquaverse/internal/api/handler"
	"github.com/cashyalla/aquaverse/internal/api/router"
	"github.com/cashyalla/aquaverse/internal/cache"
	"github.com/cashyalla/aquaverse/internal/config"
	"github.com/cashyalla/aquaverse/internal/crawler"
	"github.com/cashyalla/aquaverse/internal/marketplace"
	"github.com/cashyalla/aquaverse/internal/notification"
	"github.com/cashyalla/aquaverse/internal/pipeline"
	"github.com/cashyalla/aquaverse/internal/repository"
	"github.com/cashyalla/aquaverse/internal/service"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	miniogo "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	// ── 로거 ──────────────────────────────────────────────
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// ── 설정 ──────────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load failed", "err", err)
		os.Exit(1)
	}

	slog.Info("AquaVerse starting",
		"port", cfg.Server.Port,
		"env", cfg.Server.Env,
	)

	// ── PostgreSQL ─────────────────────────────────────────
	db, err := sqlx.Open("pgx", cfg.Database.DSN)
	if err != nil {
		slog.Error("db open failed", "err", err)
		os.Exit(1)
	}
	db.SetMaxOpenConns(cfg.Database.MaxOpenConn)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConn)
	db.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := db.PingContext(ctx); err != nil {
		slog.Error("db ping failed", "err", err)
		cancel()
		os.Exit(1)
	}
	cancel()
	slog.Info("PostgreSQL connected")

	// ── Redis ──────────────────────────────────────────────
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	rdbCtx, rdbCancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := rdb.Ping(rdbCtx).Err(); err != nil {
		slog.Error("redis ping failed", "err", err)
		rdbCancel()
		os.Exit(1)
	}
	rdbCancel()
	slog.Info("Redis connected")

	// ── 캐시 ──────────────────────────────────────────────
	redisCache := cache.NewRedisCache(rdb)

	// ── Repository ────────────────────────────────────────
	userRepo := repository.NewUserRepository(db)
	fishRepo := repository.NewFishRepository(db)
	commRepo := repository.NewCommunityRepository(db)
	mktRepo := repository.NewMarketplaceRepository(db)
	chatRepo := repository.NewChatRepository(db)
	citesRepo := repository.NewCitesRepository(db)

	// ── 사기 탐지기 ───────────────────────────────────────
	fraudDetector := marketplace.NewFraudDetector(redisCache)

	// ── 알림 서비스 ───────────────────────────────────────
	notifySvc := notification.NewService(cfg.Notification.FCMServerKey, redisCache, logger)

	// 알림 워커 시작
	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()
	notifySvc.StartWorker(appCtx)

	// ── 서비스 ────────────────────────────────────────────
	authSvc := service.NewAuthService(userRepo, cfg.Auth)
	fishSvc := service.NewFishService(fishRepo, redisCache)
	commSvc := service.NewCommunityService(commRepo)
	mktSvc := service.NewMarketplaceService(mktRepo, notifySvc, fraudDetector)

	// ChatHub: Redis Pub/Sub 기반 실시간 채팅
	chatHub := service.NewChatHub(rdb, chatRepo)
	go chatHub.StartSubscriber(appCtx)

	// 에스크로 서비스
	escrowSvc := service.NewEscrowService(db)

	// 합사 호환성 서비스
	compatSvc := service.NewCompatibilityService(db)

	// 이미지 처리 워커 (goroutine)
	imageWorker := service.NewImageWorker(db, rdb)
	go imageWorker.StartWorker(appCtx)

	// ── AI Enricher (크롤러 파이프라인용) ─────────────────
	aiEnricher := pipeline.NewAIEnricher(
		cfg.AI.APIKey, cfg.AI.Model, cfg.AI.MaxTokens, logger,
	)

	// ── PipelineProcessor ─────────────────────────────────
	pipelineProcessor := pipeline.NewPipelineProcessor(fishRepo, aiEnricher, logger)

	// ── 크롤러 스케줄러 ───────────────────────────────────
	fishbaseClient := crawler.NewFishBaseClient(
		logger, cfg.Crawler.UserAgent, cfg.Crawler.RequestsPerMinute,
	)
	scheduler := crawler.NewScheduler(fishbaseClient, pipelineProcessor, logger)
	scheduler.Start()
	defer scheduler.Stop()

	// ── MinIO 클라이언트 ──────────────────────────────────
	minioClient, err := miniogo.New(cfg.Storage.Endpoint, &miniogo.Options{
		Creds:  credentials.NewStaticV4(cfg.Storage.AccessKey, cfg.Storage.SecretKey, ""),
		Secure: cfg.Storage.UseSSL,
	})
	if err != nil {
		slog.Error("minio client init failed", "err", err)
		os.Exit(1)
	}
	slog.Info("MinIO client initialized", "endpoint", cfg.Storage.Endpoint)

	// ── 핸들러 ────────────────────────────────────────────
	authH := handler.NewAuthHandler(authSvc)
	fishH := handler.NewFishHandler(fishSvc)
	commH := handler.NewCommunityHandler(commSvc)
	mktH := handler.NewMarketplaceHandler(mktSvc)
	uploadH := handler.NewUploadHandler(minioClient, cfg.Storage.Bucket)
	chatH := handler.NewChatHandler(chatHub, chatRepo)
	phoneH := handler.NewPhoneHandler(db)
	metricsH := handler.NewMetricsHandler()
	citesH := handler.NewCitesHandler(citesRepo)
	escrowH := handler.NewEscrowHandler(escrowSvc)
	compatH := handler.NewCompatibilityHandler(compatSvc)

	// 수조 주치의 서비스 및 핸들러
	tankDoctorSvc := service.NewTankDoctorService(db, rdb)
	tankDoctorH := handler.NewTankDoctorHandler(tankDoctorSvc)

	// PG 결제 서비스 및 핸들러
	paymentSvc := service.NewPaymentService(db)
	paymentH := handler.NewPaymentHandler(paymentSvc)

	// ── Echo 라우터 설정 ───────────────────────────────────
	e := echo.New()
	e.HideBanner = true
	router.Setup(e, cfg, authH, fishH, commH, mktH, uploadH, chatH, phoneH, metricsH, citesH, escrowH, compatH, tankDoctorH, paymentH)
	router.SetupHealthCheck(e, db, rdb)

	// ── 그레이스풀 셧다운 ──────────────────────────────────
	go func() {
		addr := fmt.Sprintf(":%d", cfg.Server.Port)
		slog.Info("server listening", "addr", addr)
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down...")
	shutCtx, shutCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutCancel()

	if err := e.Shutdown(shutCtx); err != nil {
		slog.Error("shutdown error", "err", err)
	}
	db.Close()
	rdb.Close()
	slog.Info("server stopped")
}
