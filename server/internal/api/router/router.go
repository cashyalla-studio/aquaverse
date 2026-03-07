package router

import (
	"net/http"

	"github.com/cashyalla/aquaverse/internal/api/handler"
	"github.com/cashyalla/aquaverse/internal/api/middleware"
	"github.com/cashyalla/aquaverse/internal/config"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

func Setup(
	e *echo.Echo,
	cfg *config.Config,
	authH *handler.AuthHandler,
	fishH *handler.FishHandler,
	commH *handler.CommunityHandler,
	mktH *handler.MarketplaceHandler,
) {
	// 글로벌 미들웨어
	e.Use(echomw.Logger())
	e.Use(echomw.Recover())
	e.Use(echomw.CORS())
	e.Use(echomw.RequestID())
	e.Use(middleware.LocaleMiddleware())

	// 헬스체크
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"status": "ok", "version": "1.0.0"})
	})

	api := e.Group("/api/v1")

	// ── 인증 (공개) ────────────────────────────────────────
	auth := api.Group("/auth")
	auth.POST("/register", authH.Register)
	auth.POST("/login", authH.Login)
	auth.POST("/refresh", authH.Refresh)
	auth.POST("/logout", authH.Logout)
	auth.POST("/logout-all", authH.LogoutAll, middleware.JWTAuth(cfg.Auth.JWTSecret))

	// ── 열대어 백과사전 (공개) ──────────────────────────────
	fish := api.Group("/fish")
	fish.GET("", fishH.List)
	fish.GET("/search", fishH.Search)
	fish.GET("/families", fishH.ListFamilies)
	fish.GET("/:id", fishH.Get)

	// ── 커뮤니티 (게시판) ──────────────────────────────────
	boards := api.Group("/boards")
	boards.GET("", commH.ListBoards)
	boards.GET("/:boardID/posts", commH.ListPosts)
	boards.GET("/:boardID/posts/:postID", commH.GetPost)

	// 인증 필요
	authBoards := boards.Group("", middleware.JWTAuth(cfg.Auth.JWTSecret))
	authBoards.POST("/:boardID/posts", commH.CreatePost)
	authBoards.POST("/:boardID/posts/:postID/like", commH.LikePost)
	authBoards.POST("/:boardID/posts/:postID/comments", commH.CreateComment)

	// ── 마켓플레이스 (분양/입양) ───────────────────────────
	listings := api.Group("/listings")
	listings.GET("", mktH.ListListings)
	listings.GET("/:id", mktH.GetListing)

	authListings := listings.Group("", middleware.JWTAuth(cfg.Auth.JWTSecret))
	authListings.POST("", mktH.CreateListing)
	authListings.PUT("/:id/status", mktH.UpdateListingStatus)
	authListings.POST("/:id/trade", mktH.InitiateTrade)

	trades := api.Group("/trades", middleware.JWTAuth(cfg.Auth.JWTSecret))
	trades.PUT("/:id/status", mktH.UpdateTradeStatus)
	trades.POST("/:id/review", mktH.SubmitReview)

	// 어종 알림 구독
	watches := api.Group("/watches", middleware.JWTAuth(cfg.Auth.JWTSecret))
	watches.POST("", mktH.WatchFish)

	// 사기 신고
	fraud := api.Group("/fraud-reports", middleware.JWTAuth(cfg.Auth.JWTSecret))
	fraud.POST("", mktH.ReportFraud)

	// ── 관리자 (ADMIN 역할 필요) ───────────────────────────
	admin := api.Group("/admin",
		middleware.JWTAuth(cfg.Auth.JWTSecret),
		middleware.RequireRole("ADMIN"),
	)
	_ = admin // 관리자 핸들러는 별도 구현
}
