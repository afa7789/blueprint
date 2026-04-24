package main

import (
	c "context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	stripe "github.com/stripe/stripe-go/v82"
	"github.com/valyala/fasthttp/fasthttpadaptor"

	"github.com/afa/blueprint/backend/internal/handlers"
	"github.com/afa/blueprint/backend/internal/infrastructure"
	"github.com/afa/blueprint/backend/internal/infrastructure/storage"
	"github.com/afa/blueprint/backend/migrations"
	"github.com/afa/blueprint/backend/pkg/config"
	"github.com/afa/blueprint/backend/pkg/database"
	applog "github.com/afa/blueprint/backend/pkg/logger"
	"github.com/afa/blueprint/backend/pkg/metrics"
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
		pool.Close()
		log.Printf("Migrations failed: %v", err)
		return
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
			var fe *fiber.Error
			if errors.As(err, &fe) {
				code = fe.Code
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

	// Prometheus metrics middleware
	app.Use(middleware.PrometheusMiddleware())

	// General API rate limit
	app.Use(middleware.RateLimit(rdb, middleware.RateLimitConfig{
		Max:     cfg.RateLimitAPI,
		Window:  time.Minute,
		KeyFunc: middleware.KeyByIP,
	}))

	// Prometheus metrics endpoint
	app.Get("/metrics", func(c *fiber.Ctx) error {
		handler := promhttp.Handler()
		fasthttpHandler := fasthttpadaptor.NewFastHTTPHandler(handler)
		fasthttpHandler(c.Context())
		return nil
	})

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

	// Storage backend (local or s3)
	storageBackend, err := storage.NewFromConfig(c.Background(), cfg)
	if err != nil {
		log.Fatalf("Storage init failed: %v", err)
	}
	log.Printf("Storage backend: %s", cfg.StorageBackend)

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
	pixConfigRepo := infrastructure.NewPixConfigRepo(pool)
	securitySettingsRepo := infrastructure.NewSecuritySettingRepo(pool)
	profileRepo := infrastructure.NewUserProfileRepo(pool)

	// Gauge metrics updater
	go func() {
		for {
			var count int64
			ctx := c.Background()
			_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&count)
			metrics.UsersTotal.Set(float64(count))
			_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM products WHERE is_active = true").Scan(&count)
			metrics.ProductsTotal.Set(float64(count))
			_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM blog_posts").Scan(&count)
			metrics.BlogPostsTotal.Set(float64(count))
			_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM waitlist").Scan(&count)
			metrics.WaitlistTotal.Set(float64(count))
			time.Sleep(60 * time.Second)
		}
	}()

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
	paymentHandler := handlers.NewPaymentHandler(orderRepo, pixConfigRepo, cfg, storageBackend)
	blogRepo := infrastructure.NewBlogRepo(pool)
	blogHandler := handlers.NewBlogHandler(blogRepo, cfg, storageBackend)
	cronJobRepo := infrastructure.NewCronJobRepo(pool)
	jobExecRepo := infrastructure.NewJobExecutionRepo(pool)
	toolRepo := infrastructure.NewAdminToolRepo(pool)
	auditLogRepo := infrastructure.NewAuditLogRepo(pool)
	appLogRepo := infrastructure.NewAppLogRepo(pool)
	logConfigRepo := infrastructure.NewLogConfigRepo(pool)

	// Structured logger
	appLogger := applog.New(appLogRepo, "server")
	appLogger.Info(c.Background(), "Server starting", map[string]interface{}{"port": cfg.Port})
	app.Use(middleware.AppLog(appLogRepo))

	// Job registry + handler
	jobRegistry := handlers.NewJobRegistry()
	jobRegistry.Register("system.noop", func(_ c.Context) (json.RawMessage, error) {
		payload, _ := json.Marshal(map[string]interface{}{
			"ok":          true,
			"executed_at": time.Now().UTC(),
		})
		return payload, nil
	})
	jobRegistry.Register("logs.cleanup", func(ctx c.Context) (json.RawMessage, error) {
		logCfg, err := logConfigRepo.Get(ctx)
		if err != nil {
			return nil, err
		}
		deleted, err := appLogRepo.Cleanup(ctx, logCfg.RetentionDays)
		if err != nil {
			return nil, err
		}
		payload, _ := json.Marshal(map[string]interface{}{
			"deleted":        deleted,
			"retention_days": logCfg.RetentionDays,
		})
		return payload, nil
	})
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
	admin := api.Group("/admin", middleware.RequireAuth(cfg), middleware.RequireRole("admin"))
	admin.Use(middleware.AuditLog(auditLogRepo))
	admin.Get("/features", flagHandler.GetAll)
	admin.Put("/features/:key", flagHandler.Toggle)

	// Legal pages (public)
	api.Get("/legal", legalHandler.ListActive)
	api.Get("/legal/:slug", legalHandler.GetBySlug)

	// Public content routes
	api.Get("/linktree", middleware.RequireFeature(flagRepo, "linktree_enabled"), func(fc *fiber.Ctx) error {
		items, err := linktreeRepo.List(fc.Context(), true)
		if err != nil {
			return fc.Status(500).JSON(fiber.Map{"error": "failed to fetch linktree"})
		}
		return fc.JSON(items)
	})
	api.Get("/banners", middleware.RequireFeature(flagRepo, "banners_enabled"), func(fc *fiber.Ctx) error {
		items, err := bannerRepo.List(fc.Context(), true)
		if err != nil {
			return fc.Status(500).JSON(fiber.Map{"error": "failed to fetch banners"})
		}
		return fc.JSON(items)
	})
	api.Get("/brand-kit", middleware.RequireFeature(flagRepo, "brand_kit_enabled"), func(fc *fiber.Ctx) error {
		bk, err := brandKitRepo.Get(fc.Context())
		if err != nil {
			return fc.Status(500).JSON(fiber.Map{"error": "failed to fetch brand kit"})
		}
		return fc.JSON(bk)
	})

	api.Post("/waitlist", middleware.RequireFeature(flagRepo, "waitlist_enabled"), waitlistHandler.AddToWaitlist)
	api.Get("/waitlist", middleware.RequireAuth(cfg), middleware.RequireRole("admin"), waitlistHandler.GetWaitlist)

	// Admin routes (admin group already created above)
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
	api.Get("/products", middleware.RequireFeature(flagRepo, "store_enabled"), storeHandler.ListProducts)
	api.Get("/products/:id", middleware.RequireFeature(flagRepo, "store_enabled"), storeHandler.GetProduct)
	api.Get("/categories", middleware.RequireFeature(flagRepo, "store_enabled"), storeHandler.ListCategories)

	// Store — auth-required routes
	api.Post("/orders", middleware.RequireFeature(flagRepo, "store_enabled"), middleware.RequireAuth(cfg), storeHandler.CreateOrder)
	api.Get("/orders/me", middleware.RequireAuth(cfg), storeHandler.ListMyOrders)

	// Payments
	api.Post("/payments/stripe", middleware.RequireFeature(flagRepo, "payments_stripe"), middleware.RequireAuth(cfg), paymentHandler.CreateStripePayment)
	api.Post("/payments/stripe/webhook", paymentHandler.StripeWebhook)
	api.Post("/payments/pix", middleware.RequireFeature(flagRepo, "payments_pix"), middleware.RequireAuth(cfg), paymentHandler.CreatePixPayment)
	api.Post("/payments/pix/:order_id/receipt", middleware.RequireFeature(flagRepo, "payments_pix"), middleware.RequireAuth(cfg), paymentHandler.UploadPixReceipt)
	admin.Put("/orders/:id/approve-pix", paymentHandler.ApprovePixPayment)
	admin.Get("/pix-config", middleware.RequireFeature(flagRepo, "payments_pix"), paymentHandler.GetPixConfig)
	admin.Put("/pix-config", middleware.RequireFeature(flagRepo, "payments_pix"), paymentHandler.UpdatePixConfig)

	// Coupon — public validate
	api.Post("/coupons/validate", middleware.RequireFeature(flagRepo, "store_enabled"), couponHandler.ValidateCoupon)

	// Store — admin routes
	admin.Get("/products", storeHandler.AdminListProducts)
	admin.Get("/categories", storeHandler.AdminListCategories)
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
	api.Get("/blog", middleware.RequireFeature(flagRepo, "blog_enabled"), blogHandler.ListPublished)
	api.Get("/blog/rss.xml", middleware.RequireFeature(flagRepo, "blog_enabled"), blogHandler.RSSFeed)
	api.Get("/blog/atom.xml", middleware.RequireFeature(flagRepo, "blog_enabled"), blogHandler.AtomFeed)
	api.Get("/blog/:slug", middleware.RequireFeature(flagRepo, "blog_enabled"), blogHandler.GetBySlug)

	// Blog — admin routes
	admin.Get("/blog", blogHandler.AdminListPosts)
	admin.Post("/blog", blogHandler.AdminCreatePost)
	admin.Put("/blog/:id", blogHandler.AdminUpdatePost)
	admin.Delete("/blog/:id", blogHandler.AdminDeletePost)
	admin.Post("/blog/:id/cover", blogHandler.AdminUploadCover)
	admin.Post("/blog/ai-generate", middleware.RequireFeature(flagRepo, "ai_blog_enabled"), blogHandler.AdminAIGenerate)

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

	// Static file serving (only meaningful when STORAGE_BACKEND=local —
	// for S3 the URLs returned are presigned and served by S3 directly).
	staticRoot := cfg.StorageLocalPath
	if staticRoot == "" {
		staticRoot = cfg.UploadDir
	}
	staticMount := cfg.StorageURLPrefix
	if staticMount == "" {
		staticMount = "/static"
	}
	app.Static(staticMount, staticRoot)

	log.Printf("Server starting on :%s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Printf("Server error: %v", err)
		pool.Close()
		return
	}
}
