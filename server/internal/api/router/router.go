package router

import (
	"context"
	"net/http"
	"time"

	"github.com/cashyalla/aquaverse/internal/api/handler"
	"github.com/cashyalla/aquaverse/internal/api/middleware"
	"github.com/cashyalla/aquaverse/internal/config"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
	"github.com/redis/go-redis/v9"
)

// AppDependencies holds all handler and infrastructure references required by
// the router. Pass a pointer to Setup instead of individual parameters.
type AppDependencies struct {
	AuthH       *handler.AuthHandler
	FishH       *handler.FishHandler
	CommH       *handler.CommunityHandler
	MktH        *handler.MarketplaceHandler
	UploadH     *handler.UploadHandler
	ChatH       *handler.ChatHandler
	PhoneH      *handler.PhoneHandler
	MetricsH    *handler.MetricsHandler
	CitesH      *handler.CitesHandler
	EscrowH     *handler.EscrowHandler
	CompatH     *handler.CompatibilityHandler
	TankDoctorH *handler.TankDoctorHandler
	PaymentH    *handler.PaymentHandler
	BusinessH   *handler.BusinessHandler
	NotifH      *handler.NotificationHandler
	VideoH      *handler.VideoHandler
	SubH        *handler.SubscriptionHandler
	SitemapH    *handler.SitemapHandler
	AdminH      *handler.AdminHandler
	SocialH     *handler.SocialHandler
	TotpH       *handler.TOTPHandler
	SpeciesH    *handler.SpeciesIdentifyHandler
	CareHubH    *handler.CareHubHandler
	BadgeH      *handler.BadgeHandler
	RAGH        *handler.RAGHandler
	AuctionH    *handler.AuctionHandler
	ExpertH     *handler.ExpertHandler
	InventoryH  *handler.InventoryHandler
	Cfg         *config.Config
	Rdb         *redis.Client
}

func Setup(e *echo.Echo, deps *AppDependencies) {
	cfg := deps.Cfg
	rdb := deps.Rdb

	// 글로벌 미들웨어
	e.Use(echomw.Logger())
	e.Use(middleware.PrometheusMetrics())
	e.Use(echomw.Recover())
	e.Use(echomw.CORS())
	e.Use(echomw.RequestID())
	e.Use(middleware.LocaleMiddleware())

	// 헬스체크 (기본)
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok", "version": "1.0.0"})
	})

	// /metrics (Prometheus scraping, 프로덕션에서는 IP 필터 권장)
	e.GET("/metrics", deps.MetricsH.Metrics)

	// Sitemap (SEO)
	e.GET("/sitemap.xml", deps.SitemapH.Sitemap)

	api := e.Group("/api/v1")

	// ── 인증 (공개) ────────────────────────────────────────
	auth := api.Group("/auth")
	auth.POST("/register", deps.AuthH.Register, middleware.AuthRateLimit(rdb))
	auth.POST("/login", deps.AuthH.Login, middleware.AuthRateLimit(rdb))
	auth.POST("/refresh", deps.AuthH.Refresh)
	auth.POST("/logout", deps.AuthH.Logout)
	auth.POST("/logout-all", deps.AuthH.LogoutAll, middleware.JWTAuth(cfg.Auth.JWTSecret))

	// TOTP 2FA (인증 필요)
	authTOTP := auth.Group("", middleware.JWTAuth(cfg.Auth.JWTSecret))
	authTOTP.GET("/totp", deps.TotpH.GetStatus)
	authTOTP.POST("/totp/enable", deps.TotpH.Enable)
	authTOTP.POST("/totp/verify", deps.TotpH.Verify)
	authTOTP.DELETE("/totp", deps.TotpH.Disable)

	// ── 열대어 백과사전 (공개) ──────────────────────────────
	fish := api.Group("/fish")
	fish.GET("", deps.FishH.List)
	fish.GET("/search", deps.FishH.Search)
	fish.GET("/families", deps.FishH.ListFamilies)
	fish.GET("/categories", deps.FishH.ListCategories)
	fish.GET("/:id", deps.FishH.Get)
	fish.GET("/:id/compatible", deps.CompatH.GetCompatibleFish)
	fish.GET("/check-compat", deps.CompatH.CheckWithClaude)

	// ── 수조 (인증 필요) ───────────────────────────────────
	tanks := api.Group("/tanks", middleware.JWTAuth(cfg.Auth.JWTSecret))
	tanks.GET("/:id/recommend", deps.CompatH.RecommendForTank)
	tanks.GET("/:id/inhabitants", deps.CompatH.GetTankInhabitants)
	tanks.POST("/:id/water-params", deps.TankDoctorH.RecordWaterParams)
	tanks.GET("/:id/water-params", deps.TankDoctorH.GetWaterHistory)
	tanks.GET("/:id/diagnosis", deps.TankDoctorH.GetDiagnosis)
	tanks.POST("/:id/ocr-params", deps.TankDoctorH.OCRWaterParams)

	// ── 커뮤니티 (게시판) ──────────────────────────────────
	boards := api.Group("/boards")
	boards.GET("", deps.CommH.ListBoards)
	boards.GET("/:boardID/posts", deps.CommH.ListPosts)
	boards.GET("/:boardID/posts/:postID", deps.CommH.GetPost)

	// 인증 필요
	authBoards := boards.Group("", middleware.JWTAuth(cfg.Auth.JWTSecret))
	authBoards.POST("/:boardID/posts", deps.CommH.CreatePost)
	authBoards.POST("/:boardID/posts/:postID/like", deps.CommH.LikePost)
	authBoards.POST("/:boardID/posts/:postID/comments", deps.CommH.CreateComment)

	// ── 마켓플레이스 (분양/입양) ───────────────────────────
	listings := api.Group("/listings")
	listings.GET("", deps.MktH.ListListings)
	listings.GET("/:id", deps.MktH.GetListing)

	authListings := listings.Group("", middleware.JWTAuth(cfg.Auth.JWTSecret))
	authListings.POST("", deps.MktH.CreateListing)
	authListings.PUT("/:id/status", deps.MktH.UpdateListingStatus)
	authListings.POST("/:id/trade", deps.MktH.InitiateTrade)

	trades := api.Group("/trades", middleware.JWTAuth(cfg.Auth.JWTSecret))
	trades.PUT("/:id/status", deps.MktH.UpdateTradeStatus)
	trades.POST("/:id/review", deps.MktH.SubmitReview)

	// WebSocket 채팅
	trades.GET("/:id/chat", deps.ChatH.Connect)
	trades.GET("/:id/chat/history", deps.ChatH.GetHistory)

	// 에스크로
	escrow := trades.Group("/:id/escrow")
	escrow.GET("", deps.EscrowH.GetStatus)
	escrow.POST("/fund", deps.EscrowH.Fund)
	escrow.POST("/release", deps.EscrowH.Release)
	escrow.POST("/refund", deps.EscrowH.Refund)

	// PG 결제
	trades.POST("/:id/payment/initiate", deps.PaymentH.InitiatePayment)
	trades.POST("/:id/payment/mock-confirm", deps.PaymentH.MockConfirm)

	// 토스페이먼츠 웹훅 (공개 엔드포인트 - 인증 없음)
	api.POST("/webhooks/payment", deps.PaymentH.Webhook)

	// CITES 멸종위기 어종 체크 (공개)
	api.GET("/cites/check", deps.CitesH.Check)

	// 전화번호 인증
	phone := api.Group("/phone", middleware.JWTAuth(cfg.Auth.JWTSecret))
	phone.POST("/send", deps.PhoneH.SendCode)
	phone.POST("/verify", deps.PhoneH.VerifyCode)

	// 어종 알림 구독
	watches := api.Group("/watches", middleware.JWTAuth(cfg.Auth.JWTSecret))
	watches.POST("", deps.MktH.WatchFish)

	// 사기 신고
	fraud := api.Group("/fraud-reports", middleware.JWTAuth(cfg.Auth.JWTSecret))
	fraud.POST("", deps.MktH.ReportFraud)

	// ── 파일 업로드 (인증 필요) ─────────────────────────────
	upload := api.Group("/upload", middleware.JWTAuth(cfg.Auth.JWTSecret))
	upload.POST("/presign", deps.UploadH.PresignUpload)

	// ── 업체 프로필 ────────────────────────────────────────
	// 공개 조회
	businesses := api.Group("/businesses")
	businesses.GET("", deps.BusinessH.ListBusinesses)
	businesses.GET("/nearby", deps.BusinessH.NearbyBusinesses)
	businesses.GET("/:id", deps.BusinessH.GetBusiness)
	businesses.GET("/:id/reviews", deps.BusinessH.GetReviews)

	// 업체 등록/수정/리뷰 (인증 필요)
	authBusinesses := businesses.Group("", middleware.JWTAuth(cfg.Auth.JWTSecret))
	authBusinesses.POST("", deps.BusinessH.CreateBusiness)
	authBusinesses.PUT("/:id", deps.BusinessH.UpdateBusiness)
	authBusinesses.POST("/:id/reviews", deps.BusinessH.AddReview)

	// ── 푸시 알림 (FCM 토큰 관리) ─────────────────────────
	notif := api.Group("/notifications", middleware.JWTAuth(cfg.Auth.JWTSecret))
	notif.POST("/fcm/register", deps.NotifH.RegisterToken)
	notif.DELETE("/fcm/unregister", deps.NotifH.UnregisterToken)

	// ── 영상 피드 (GET 공개, 나머지 인증 필요) ─────────────
	videos := api.Group("/videos")
	videos.GET("", deps.VideoH.GetFeed)

	authVideos := videos.Group("", middleware.JWTAuth(cfg.Auth.JWTSecret))
	authVideos.POST("", deps.VideoH.CreatePost)
	authVideos.POST("/:id/like", deps.VideoH.LikePost)
	authVideos.POST("/:id/view", deps.VideoH.IncrementView)
	authVideos.DELETE("/:id", deps.VideoH.DeletePost)

	// ── 구독 (플랜 조회는 공개) ────────────────────────────
	api.GET("/subscriptions/plans", deps.SubH.GetPlans)
	subscriptions := api.Group("/subscriptions", middleware.JWTAuth(cfg.Auth.JWTSecret))
	subscriptions.GET("/me", deps.SubH.GetMySubscription)
	subscriptions.POST("/subscribe", deps.SubH.Subscribe)
	subscriptions.POST("/cancel", deps.SubH.Cancel)

	// ── 관리자 (ADMIN 역할 필요) ───────────────────────────
	admin := api.Group("/admin",
		middleware.JWTAuth(cfg.Auth.JWTSecret),
		middleware.RequireRole("ADMIN"),
	)
	admin.GET("/kpi", deps.AdminH.GetKPI)
	admin.GET("/users", deps.AdminH.ListUsers)
	admin.POST("/users/:id/ban", deps.AdminH.BanUser)
	admin.POST("/users/:id/unban", deps.AdminH.UnbanUser)
	admin.GET("/audit-logs", deps.AdminH.GetAuditLogs)
	admin.GET("/cites-stats", deps.AdminH.GetCitesStats)

	// ── 소셜 그래프 ─────────────────────────────────────
	social := api.Group("/social", middleware.JWTAuth(cfg.Auth.JWTSecret))
	social.GET("/feed", deps.SocialH.GetFeed)
	social.GET("/suggestions", deps.SocialH.GetSuggestions)
	social.GET("/following", deps.SocialH.GetFollowing)
	social.POST("/users/:id/follow", deps.SocialH.Follow)
	social.DELETE("/users/:id/follow", deps.SocialH.Unfollow)

	// ── AI 어종 식별 (공개, 선택적 인증) ────────────────────
	api.POST("/species/identify", deps.SpeciesH.Identify)

	// ── 케어 허브 (인증 필요) ────────────────────────────────
	// 수조별 케어 일정 관리
	tanks.POST("/:id/schedules", deps.CareHubH.CreateSchedule)
	tanks.GET("/:id/schedules", deps.CareHubH.ListSchedules)

	// 개별 케어 일정 수정/삭제/완료
	schedules := api.Group("/schedules", middleware.JWTAuth(cfg.Auth.JWTSecret))
	schedules.PUT("/:id", deps.CareHubH.UpdateSchedule)
	schedules.DELETE("/:id", deps.CareHubH.DeleteSchedule)
	schedules.POST("/:id/complete", deps.CareHubH.CompleteSchedule)

	// 사용자 케어 현황
	users := api.Group("/users/me", middleware.JWTAuth(cfg.Auth.JWTSecret))
	users.GET("/care-today", deps.CareHubH.GetTodayTasks)
	users.GET("/streak", deps.CareHubH.GetStreak)

	// ── 뱃지 & 챌린지 ─────────────────────────────────────
	// 뱃지 정의 목록 (공개)
	api.GET("/badges", deps.BadgeH.ListBadges)

	// 내 뱃지 조회 (인증 필요)
	users.GET("/badges", deps.BadgeH.GetMyBadges)

	// 특정 사용자 뱃지 조회 (공개)
	api.GET("/users/:id/badges", deps.BadgeH.GetUserBadges)

	// 챌린지 목록 (공개)
	api.GET("/challenges", deps.BadgeH.ListChallenges)

	// 챌린지 참가 / 진행도 조회 (인증 필요)
	challenges := api.Group("/challenges", middleware.JWTAuth(cfg.Auth.JWTSecret))
	challenges.POST("/:id/join", deps.BadgeH.JoinChallenge)
	challenges.GET("/:id/progress", deps.BadgeH.GetChallengeProgress)

	// ── RAG 챗봇 (pgvector 기반) ─────────────────────────
	// 질문하기: 선택적 인증 (비로그인도 허용)
	api.POST("/chat/ask", deps.RAGH.Ask)
	// 내 세션 목록: 인증 필요
	chat := api.Group("/chat", middleware.JWTAuth(cfg.Auth.JWTSecret))
	chat.GET("/sessions", deps.RAGH.GetSessions)
	// 세션 메시지 목록: 공개
	api.GET("/chat/sessions/:id/messages", deps.RAGH.GetMessages)
	// 어종 인덱싱: 관리자 전용
	admin.POST("/fish/:id/index", deps.RAGH.IndexFish)

	// ── 실시간 경매 ───────────────────────────────────────
	// 공개 조회
	auctions := api.Group("/auctions")
	auctions.GET("", deps.AuctionH.ListAuctions)
	auctions.GET("/:id", deps.AuctionH.GetAuction)
	// WebSocket 실시간 연결 (공개, 내부에서 선택적 JWT)
	auctions.GET("/:id/ws", deps.AuctionH.Connect)
	// 인증 필요
	authAuctions := auctions.Group("", middleware.JWTAuth(cfg.Auth.JWTSecret))
	authAuctions.POST("", deps.AuctionH.CreateAuction)
	authAuctions.POST("/:id/bid", deps.AuctionH.PlaceBid)
	authAuctions.POST("/:id/end", deps.AuctionH.EndAuction)

	// ── Expert Connect ────────────────────────────────────
	// 전문가 목록/상세 (공개)
	experts := api.Group("/experts")
	experts.GET("", deps.ExpertH.ListExperts)
	experts.GET("/:id", deps.ExpertH.GetExpert)
	// 프로필 등록/수정, 상담 신청 (인증 필요)
	authExperts := experts.Group("", middleware.JWTAuth(cfg.Auth.JWTSecret))
	authExperts.PUT("/profile", deps.ExpertH.UpsertProfile)
	authExperts.POST("/:id/consult", deps.ExpertH.CreateConsultation)
	// 상담 관련 (인증 필요)
	consultations := api.Group("/consultations", middleware.JWTAuth(cfg.Auth.JWTSecret))
	consultations.GET("/me", deps.ExpertH.GetMyConsultations)
	consultations.PUT("/:id/status", deps.ExpertH.UpdateConsultationStatus)
	consultations.POST("/:id/review", deps.ExpertH.CreateReview)

	// ── Business Hub 재고 관리 ────────────────────────────
	// 재고 조회 및 업체 통계 (공개)
	businesses.GET("/:id/inventory", deps.InventoryH.ListInventory)
	businesses.GET("/:id/stats", deps.InventoryH.GetBusinessStats)
	// 재고 추가/수정/삭제 (인증 필요)
	authBusinesses.POST("/:id/inventory", deps.InventoryH.CreateInventory)
	authBusinesses.PUT("/:id/inventory/:itemId", deps.InventoryH.UpdateInventory)
	authBusinesses.DELETE("/:id/inventory/:itemId", deps.InventoryH.DeleteInventory)
}

// SetupHealthCheck DB와 Redis ping을 포함하는 강화된 헬스체크를 등록한다.
// Setup 호출 이후 별도로 호출하여 /health 라우트를 덮어쓴다.
func SetupHealthCheck(e *echo.Echo, db *sqlx.DB, rdb *redis.Client) {
	e.GET("/health", func(c echo.Context) error {
		ctx, cancel := context.WithTimeout(c.Request().Context(), 3*time.Second)
		defer cancel()

		type componentStatus struct {
			Status string `json:"status"`
			Error  string `json:"error,omitempty"`
		}

		dbStatus := componentStatus{Status: "ok"}
		if err := db.PingContext(ctx); err != nil {
			dbStatus.Status = "error"
			dbStatus.Error = err.Error()
		}

		redisStatus := componentStatus{Status: "ok"}
		if err := rdb.Ping(ctx).Err(); err != nil {
			redisStatus.Status = "error"
			redisStatus.Error = err.Error()
		}

		overall := "ok"
		httpStatus := http.StatusOK
		if dbStatus.Status != "ok" || redisStatus.Status != "ok" {
			overall = "degraded"
			httpStatus = http.StatusServiceUnavailable
		}

		return c.JSON(httpStatus, map[string]interface{}{
			"status":  overall,
			"version": "1.0.0",
			"components": map[string]interface{}{
				"database": dbStatus,
				"redis":    redisStatus,
			},
		})
	})
}
