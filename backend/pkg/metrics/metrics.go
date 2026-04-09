package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics
	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "blueprint_http_requests_total",
		Help: "Total HTTP requests",
	}, []string{"method", "path", "status"})

	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "blueprint_http_request_duration_seconds",
		Help:    "HTTP request duration",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	// Auth metrics
	AuthLoginsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "blueprint_auth_logins_total",
		Help: "Total login attempts",
	}, []string{"status"}) // success, failed

	AuthRegistrationsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "blueprint_auth_registrations_total",
		Help: "Total registrations",
	})

	// Users
	UsersTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "blueprint_users_total",
		Help: "Total registered users",
	})

	// Store metrics
	OrdersTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "blueprint_orders_total",
		Help: "Total orders created",
	}, []string{"status", "payment_method"})

	OrdersRevenue = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "blueprint_orders_revenue_total",
		Help: "Total revenue in cents",
	}, []string{"payment_method"})

	ProductsTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "blueprint_products_total",
		Help: "Total active products",
	})

	// Blog
	BlogPostsTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "blueprint_blog_posts_total",
		Help: "Total blog posts",
	})

	// Waitlist
	WaitlistTotal = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "blueprint_waitlist_total",
		Help: "Total waitlist entries",
	})

	// Rate limiting
	RateLimitHits = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "blueprint_rate_limit_hits_total",
		Help: "Rate limit hits (429 responses)",
	}, []string{"key_type"}) // ip, email

	// Errors
	AppErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "blueprint_errors_total",
		Help: "Total application errors",
	}, []string{"source"})
)
