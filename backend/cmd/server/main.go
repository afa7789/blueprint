package main

import (
	c "context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	stripe "github.com/stripe/stripe-go/v82"

	"github.com/afa/blueprint/backend/internal/handlers"
	"github.com/afa/blueprint/backend/internal/infrastructure"
	"github.com/afa/blueprint/backend/migrations"
	"github.com/afa/blueprint/backend/pkg/config"
	"github.com/afa/blueprint/backend/pkg/database"
	applog "github.com/afa/blueprint/backend/pkg/logger"
	"github.com/afa/blueprint/backend/pkg/middleware"
)

func main() {
	cfg := config.Load()

	// Stripe
	stripe.Key = cfg.StripeKey

	// Database
	pool, err := database.NewPool(cfg.DBURL)
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	defer pool.Close()
	log.Println("Connected to PostgreSQL")

	// Migrations
	if err := database.RunMigrations(migrations.FS, cfg.DBMURL); err != nil {
		log.Fatalf("Migrations failed: %v", err)
	}
	log.Println("Migrations applied")

	// Redis (optional — continue if not available)
	rdb, err := database.NewRedisClient(cfg.RedisURL)
	if err != nil {
		log.Printf("Redis not available: %v (continuing without cache)", err)
	} else {
		log.Println("Connected to Redis")
	}

	// Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.FrontendURL + ", http://localhost:5173, http://localhost:3000, http://localhost, https://*.ngrok-free.app, https://*.ngrok-free.dev",
		AllowCredentials: true,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
	}))

	// Security headers and request size limit
	app.Use(middleware.SecurityHeaders())
	app.Use(middleware.RequestSizeLimit(cfg.MaxRequestBodyMB * 1024 * 1024))

	// General API rate limit
	app.Use(middleware.RateLimit(rdb, middleware.RateLimitConfig{
		Max:     cfg.RateLimitAPI,
		Window:  time.Minute,
		KeyFunc: middleware.KeyByIP,
	}))

	// Health check
	app.Get("/healthz", func(c *fiber.Ctx) error {
		if err := pool.Ping(c.Context()); err != nil {
			return c.Status(503).JSON(fiber.Map{"db": "down", "error": err.Error()})
		}
		result := fiber.Map{"db": "ok"}
		if rdb != nil {
			if err := rdb.Ping(c.Context()).Err(); err != nil {
				result["redis"] = "down"
			} else {
				result["redis"] = "ok"
			}
		}
		return c.JSON(result)
	})

	// Repositories
	userRepo := infrastructure.NewUserRepo(pool)
	flagRepo := infrastructure.NewFeatureFlagRepo(pool)
	waitlistRepo := infrastructure.NewWaitlistRepo(pool)
	bannerRepo := infrastructure.NewBannerRepo(pool)
	linktreeRepo := infrastructure.NewLinktreeRepo(pool)
	brandKitRepo := infrastructure.NewBrandKitRepo(pool)
	emailGroupRepo := infrastructure.NewEmailGroupRepo(pool)
	emailSubRepo := infrastructure.NewEmailSubscriptionRepo(pool)
	userGroupRepo := infrastructure.NewUserGroupRepo(pool)
	productRepo := infrastructure.NewProductRepo(pool)
	categoryRepo := infrastructure.NewCategoryRepo(pool)
	orderRepo := infrastructure.NewOrderRepo(pool)
	couponRepo := infrastructure.NewCouponRepo(pool)
	securitySettingsRepo := infrastructure.NewSecuritySettingRepo(pool)
	profileRepo := infrastructure.NewUserProfileRepo(pool)

	// Handlers
	authHandler := handlers.NewAuthHandler(userRepo, flagRepo, rdb, cfg)
	userHandler := handlers.NewUserHandler(userRepo, profileRepo, cfg)
	securityHandler := handlers.NewSecurityHandler(securitySettingsRepo)
	envConfigHandler := handlers.NewEnvConfigHandler(securitySettingsRepo, flagRepo, cfg)
	flagHandler := handlers.NewFeatureFlagHandler(flagRepo)
	waitlistHandler := handlers.NewWaitlistHandler(waitlistRepo)
	adminHandler := handlers.NewAdminHandler(userRepo, bannerRepo, linktreeRepo, brandKitRepo, emailGroupRepo, emailSubRepo, userGroupRepo, cfg)
	storeHandler := handlers.NewStoreHandler(productRepo, categoryRepo, orderRepo, couponRepo, cfg)
	couponHandler := handlers.NewCouponHandler(couponRepo)
	paymentHandler := handlers.NewPaymentHandler(orderRepo, cfg)
	blogRepo := infrastructure.NewBlogRepo(pool)
	blogHandler := handlers.NewBlogHandler(blogRepo, cfg)
	cronJobRepo := infrastructure.NewCronJobRepo(pool)
	jobExecRepo := infrastructure.NewJobExecutionRepo(pool)
	toolRepo := infrastructure.NewAdminToolRepo(pool)
	auditLogRepo := infrastructure.NewAuditLogRepo(pool)
	appLogRepo := infrastructure.NewAppLogRepo(pool)
	logConfigRepo := infrastructure.NewLogConfigRepo(pool)

	// Structured logger
	appLogger := applog.New(appLogRepo, "server")
	appLogger.Info(c.Background(), "Server starting", map[string]interface{}{"port": cfg.Port})

	// Job registry + handler
	jobRegistry := handlers.NewJobRegistry()
	jobsHandler := handlers.NewJobsHandler(cronJobRepo, jobExecRepo, jobRegistry)
	if err := jobsHandler.StartScheduler(c.Background()); err != nil {
		log.Printf("Failed to start job scheduler: %v", err)
	}

	// Tools + Logs + Legal handlers
	toolsHandler := handlers.NewToolsHandler(toolRepo, cfg)
	logsHandler := handlers.NewLogsHandler(appLogRepo, auditLogRepo, logConfigRepo)
	legalRepo := infrastructure.NewLegalPageRepo(pool)
	legalHandler := handlers.NewLegalHandler(legalRepo)

	// Routes
	api := app.Group("/api/v1")

	auth := api.Group("/auth")
	auth.Post("/register",
		middleware.RateLimit(rdb, middleware.RateLimitConfig{Max: cfg.RateLimitRegister, Window: time.Hour, KeyFunc: middleware.KeyByEmail}),
		authHandler.Register)
	auth.Post("/login",
		middleware.RateLimit(rdb, middleware.RateLimitConfig{Max: cfg.RateLimitAuth, Window: time.Minute, KeyFunc: middleware.KeyByEmail}),
		authHandler.Login)
	auth.Post("/refresh", authHandler.Refresh)
	auth.Post("/logout", authHandler.Logout)
	auth.Get("/me", middleware.RequireAuth(cfg), authHandler.Me)
	auth.Post("/forgot-password",
		middleware.RateLimit(rdb, middleware.RateLimitConfig{Max: cfg.RateLimitForgot, Window: time.Hour, KeyFunc: middleware.KeyByEmail}),
		authHandler.ForgotPassword)
	auth.Post("/reset-password", authHandler.ResetPassword)
	auth.Post("/verify-email", authHandler.VerifyEmail)

	// User panel routes
	user := api.Group("/user", middleware.RequireAuth(cfg))
	user.Get("/profile", userHandler.GetProfile)
	user.Put("/profile", userHandler.UpdateProfile)
	user.Put("/password", userHandler.ChangePassword)
	user.Get("/saved-cards", userHandler.ListSavedCards)
	user.Post("/saved-cards", userHandler.CreateSetupIntent)
	user.Delete("/saved-cards/:id", userHandler.DeleteSavedCard)

	api.Get("/features", flagHandler.GetAll)
	api.Put("/admin/features/:key", middleware.RequireAuth(cfg), middleware.RequireRole("admin"), flagHandler.Toggle)

	// Legal pages (public)
	api.Get("/legal", legalHandler.ListActive)
	api.Get("/legal/:slug", legalHandler.GetBySlug)

	// Public content routes
	api.Get("/linktree", func(fc *fiber.Ctx) error {
		items, err := linktreeRepo.List(fc.Context(), true)
		if err != nil {
			return fc.Status(500).JSON(fiber.Map{"error": "failed to fetch linktree"})
		}
		return fc.JSON(items)
	})
	api.Get("/banners", func(fc *fiber.Ctx) error {
		items, err := bannerRepo.List(fc.Context(), true)
		if err != nil {
			return fc.Status(500).JSON(fiber.Map{"error": "failed to fetch banners"})
		}
		return fc.JSON(items)
	})
	api.Get("/brand-kit", func(fc *fiber.Ctx) error {
		bk, err := brandKitRepo.Get(fc.Context())
		if err != nil {
			return fc.Status(500).JSON(fiber.Map{"error": "failed to fetch brand kit"})
		}
		return fc.JSON(bk)
	})

	api.Post("/waitlist", waitlistHandler.AddToWaitlist)
	api.Get("/waitlist", middleware.RequireAuth(cfg), middleware.RequireRole("admin"), waitlistHandler.GetWaitlist)

	// Admin routes
	admin := api.Group("/admin", middleware.RequireAuth(cfg), middleware.RequireRole("admin"))

	admin.Get("/users", adminHandler.ListUsers)
	admin.Put("/users/:id/role", adminHandler.UpdateUserRole)
	admin.Delete("/users/:id", adminHandler.DeleteUser)

	admin.Get("/banners", adminHandler.ListBanners)
	admin.Post("/banners", adminHandler.CreateBanner)
	admin.Put("/banners/:id", adminHandler.UpdateBanner)
	admin.Delete("/banners/:id", adminHandler.DeleteBanner)

	admin.Get("/linktree", adminHandler.ListLinktree)
	admin.Post("/linktree", adminHandler.CreateLinktreeItem)
	admin.Put("/linktree/reorder", adminHandler.ReorderLinktree)
	admin.Put("/linktree/:id", adminHandler.UpdateLinktreeItem)
	admin.Delete("/linktree/:id", adminHandler.DeleteLinktreeItem)

	admin.Get("/brand-kit", adminHandler.GetBrandKit)
	admin.Put("/brand-kit", adminHandler.UpsertBrandKit)

	admin.Get("/email-groups", adminHandler.ListEmailGroups)
	admin.Post("/email-groups", adminHandler.CreateEmailGroup)
	admin.Delete("/email-groups/:id", adminHandler.DeleteEmailGroup)
	admin.Get("/email-groups/:id/subscribers", adminHandler.ListSubscribers)

	admin.Post("/email-subscriptions", adminHandler.AddEmailSubscription)
	admin.Put("/email-subscriptions/:email/deactivate", adminHandler.DeactivateEmailSubscription)

	admin.Get("/user-groups", adminHandler.ListUserGroups)
	admin.Post("/user-groups", adminHandler.CreateUserGroup)
	admin.Delete("/user-groups/:id", adminHandler.DeleteUserGroup)

	// Store — public routes
	api.Get("/products", storeHandler.ListProducts)
	api.Get("/products/:id", storeHandler.GetProduct)
	api.Get("/categories", storeHandler.ListCategories)

	// Store — auth-required routes
	api.Post("/orders", middleware.RequireAuth(cfg), storeHandler.CreateOrder)
	api.Get("/orders/me", middleware.RequireAuth(cfg), storeHandler.ListMyOrders)

	// Payments
	api.Post("/payments/stripe", middleware.RequireAuth(cfg), paymentHandler.CreateStripePayment)
	api.Post("/payments/stripe/webhook", paymentHandler.StripeWebhook)
	api.Post("/payments/pix", middleware.RequireAuth(cfg), paymentHandler.CreatePixPayment)
	admin.Put("/orders/:id/approve-pix", paymentHandler.ApprovePixPayment)

	// Coupon — public validate
	api.Post("/coupons/validate", couponHandler.ValidateCoupon)

	// Store — admin routes
	admin.Post("/products", storeHandler.AdminCreateProduct)
	admin.Put("/products/:id", storeHandler.AdminUpdateProduct)
	admin.Delete("/products/:id", storeHandler.AdminDeleteProduct)
	admin.Post("/categories", storeHandler.AdminCreateCategory)
	admin.Put("/categories/:id", storeHandler.AdminUpdateCategory)
	admin.Delete("/categories/:id", storeHandler.AdminDeleteCategory)
	admin.Get("/orders", storeHandler.AdminListOrders)
	admin.Put("/orders/:id/status", storeHandler.AdminUpdateOrderStatus)

	// Coupon — admin routes
	admin.Get("/coupons", couponHandler.AdminListCoupons)
	admin.Post("/coupons", couponHandler.AdminCreateCoupon)
	admin.Delete("/coupons/:id", couponHandler.AdminDeleteCoupon)

	// Blog — public routes
	api.Get("/blog", blogHandler.ListPublished)
	api.Get("/blog/:slug", blogHandler.GetBySlug)

	// Blog — admin routes
	admin.Get("/blog", blogHandler.AdminListPosts)
	admin.Post("/blog", blogHandler.AdminCreatePost)
	admin.Put("/blog/:id", blogHandler.AdminUpdatePost)
	admin.Delete("/blog/:id", blogHandler.AdminDeletePost)
	admin.Post("/blog/:id/cover", blogHandler.AdminUploadCover)
	admin.Post("/blog/ai-generate", blogHandler.AdminAIGenerate)

	// Audit middleware on admin group
	admin.Use(middleware.AuditLog(auditLogRepo))

	// Jobs routes
	admin.Get("/jobs/handlers", jobsHandler.ListHandlers)
	admin.Get("/jobs", jobsHandler.ListJobs)
	admin.Post("/jobs", jobsHandler.CreateJob)
	admin.Put("/jobs/:id", jobsHandler.UpdateJob)
	admin.Put("/jobs/:id/pause", jobsHandler.PauseJob)
	admin.Put("/jobs/:id/resume", jobsHandler.ResumeJob)
	admin.Post("/jobs/:id/run", jobsHandler.RunNow)
	admin.Get("/jobs/:id/executions", jobsHandler.ListExecutions)
	admin.Post("/jobs/:id/executions/:eid/retry", jobsHandler.RetryExecution)
	admin.Delete("/jobs/:id", jobsHandler.DeleteJob)

	// Tools routes
	admin.Get("/tools", toolsHandler.ListTools)
	admin.Post("/tools", toolsHandler.CreateTool)
	admin.Put("/tools/:id", toolsHandler.UpdateTool)
	admin.Delete("/tools/:id", toolsHandler.DeleteTool)
	admin.Get("/tools/:id/ping", toolsHandler.PingTool)

	// Logs routes
	admin.Get("/logs/config", logsHandler.GetLogConfig)
	admin.Put("/logs/config", logsHandler.UpdateLogConfig)
	admin.Get("/logs/stream", logsHandler.StreamLogs)
	admin.Get("/logs", logsHandler.ListLogs)
	admin.Post("/logs/cleanup", logsHandler.CleanupLogs)
	admin.Get("/audit", logsHandler.ListAuditLogs)

	// Legal pages admin
	admin.Get("/legal", legalHandler.AdminList)
	admin.Post("/legal", legalHandler.AdminCreate)
	admin.Put("/legal/:id", legalHandler.AdminUpdate)
	admin.Delete("/legal/:id", legalHandler.AdminDelete)

	// Security settings admin
	admin.Get("/security", securityHandler.ListSettings)
	admin.Put("/security/:key", securityHandler.UpdateSetting)

	// ENV/Config panel
	admin.Get("/config/env", envConfigHandler.GetEnvStatus)
	admin.Get("/config/export", envConfigHandler.ExportEnv)
	admin.Post("/config/import", envConfigHandler.ImportEnv)

	// Static file serving
	app.Static("/static", cfg.UploadDir)

	log.Printf("Server starting on :%s", cfg.Port)
	log.Fatal(app.Listen(":" + cfg.Port))
}
