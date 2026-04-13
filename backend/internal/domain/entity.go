package domain

import (
	"encoding/json"
	"time"
)

type User struct {
	ID               string     `json:"id"`
	Email            string     `json:"email"`
	PasswordHash     string     `json:"password_hash"`
	Name             *string    `json:"name"`
	Role             string     `json:"role"`
	EmailVerified    bool       `json:"email_verified"`
	EmailVerifiedAt  *time.Time `json:"email_verified_at,omitempty"`
	StripeCustomerID *string    `json:"stripe_customer_id,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

type SecuritySetting struct {
	ID          int       `json:"id"`
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	Description *string   `json:"description"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UserProfile struct {
	ID        string          `json:"id"`
	UserID    string          `json:"user_id"`
	Phone     *string         `json:"phone"`
	AvatarURL *string         `json:"avatar_url"`
	Address   json.RawMessage `json:"address"`
	Metadata  json.RawMessage `json:"metadata"`
}

type FeatureFlag struct {
	ID        int       `json:"id"`
	Key       string    `json:"key"`
	Enabled   bool      `json:"enabled"`
	UpdatedAt time.Time `json:"updated_at"`
}

type BlogPost struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	Slug        string          `json:"slug"`
	Content     *string         `json:"content"`
	Excerpt     *string         `json:"excerpt"`
	CoverImage  *string         `json:"cover_image"`
	Status      string          `json:"status"`
	AuthorID    *string         `json:"author_id"`
	CreatedAt   time.Time       `json:"created_at"`
	PublishedAt *time.Time      `json:"published_at"`
	Metadata    json.RawMessage `json:"metadata"`
}

type Category struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Slug     *string `json:"slug"`
	ParentID *string `json:"parent_id"`
}

type Product struct {
	ID                 string          `json:"id"`
	Name               string          `json:"name"`
	Description        *string         `json:"description"`
	Price              float64         `json:"price"`
	Stock              int             `json:"stock"`
	IsPreSale          bool            `json:"is_pre_sale"`
	PreSaleAvailableAt *time.Time      `json:"pre_sale_available_at"`
	Images             json.RawMessage `json:"images"`
	CategoryID         *string         `json:"category_id"`
	IsActive           bool            `json:"is_active"`
	CreatedAt          time.Time       `json:"created_at"`
	UpdatedAt          time.Time       `json:"updated_at"`
}

type Order struct {
	ID              string          `json:"id"`
	UserID          *string         `json:"user_id"`
	Status          string          `json:"status"`
	Total           float64         `json:"total"`
	PaymentMethod   *string         `json:"payment_method"`
	PaymentID       *string         `json:"payment_id"`
	ShippingAddress json.RawMessage `json:"shipping_address"`
	TrackingCode    *string         `json:"tracking_code"`
	ReceiptURL      *string         `json:"receipt_url"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type OrderItem struct {
	ID        string  `json:"id"`
	OrderID   string  `json:"order_id"`
	ProductID *string `json:"product_id"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unit_price"`
}

type Coupon struct {
	ID            string     `json:"id"`
	Code          string     `json:"code"`
	DiscountType  *string    `json:"discount_type"`
	DiscountValue float64    `json:"discount_value"`
	MinPurchase   *float64   `json:"min_purchase"`
	ValidFrom     *time.Time `json:"valid_from"`
	ValidUntil    *time.Time `json:"valid_until"`
	MaxUses       *int       `json:"max_uses"`
	UsedCount     int        `json:"used_count"`
	IsActive      bool       `json:"is_active"`
}

type UserGroup struct {
	ID                 string  `json:"id"`
	Name               string  `json:"name"`
	DiscountPercentage float64 `json:"discount_percentage"`
}

type EmailGroup struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type EmailSubscription struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	GroupID      *string   `json:"group_id"`
	IsActive     bool      `json:"is_active"`
	SubscribedAt time.Time `json:"subscribed_at"`
}

type Banner struct {
	ID              string     `json:"id"`
	Title           *string    `json:"title"`
	ImageURL        string     `json:"image_url"`
	LinkURL         *string    `json:"link_url"`
	TargetProfile   *string    `json:"target_profile"`
	IsActive        bool       `json:"is_active"`
	StartDate       *time.Time `json:"start_date"`
	EndDate         *time.Time `json:"end_date"`
	DisplayDuration *int       `json:"display_duration"`
	CacheKey        *string    `json:"cache_key"`
	OrderIndex      int        `json:"order_index"`
}

type LinktreeItem struct {
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	URL        string  `json:"url"`
	Icon       *string `json:"icon"`
	OrderIndex int     `json:"order_index"`
	IsActive   bool    `json:"is_active"`
}

type BrandKit struct {
	ID               string `json:"id"`
	PrimaryColor     string `json:"primary_color"`
	SecondaryColor   string `json:"secondary_color"`
	AccentColor      string `json:"accent_color"`
	AccentBg         string `json:"accent_bg"`
	AccentBorder     string `json:"accent_border"`
	TextColor        string `json:"text_color"`
	TextHeadingColor string `json:"text_heading_color"`
	BgColor          string `json:"bg_color"`
	BorderColor      string `json:"border_color"`
	CodeBgColor      string `json:"code_bg_color"`

	// Dark mode
	DarkAccentColor      string `json:"dark_accent_color"`
	DarkAccentBg         string `json:"dark_accent_bg"`
	DarkAccentBorder     string `json:"dark_accent_border"`
	DarkTextColor        string `json:"dark_text_color"`
	DarkTextHeadingColor string `json:"dark_text_heading_color"`
	DarkBgColor          string `json:"dark_bg_color"`
	DarkBorderColor      string `json:"dark_border_color"`
	DarkCodeBgColor      string `json:"dark_code_bg_color"`

	// Fonts
	LogoURL      *string `json:"logo_url"`
	FaviconURL   *string `json:"favicon_url"`
	FontFamily   *string `json:"font_family"`
	HeadingFont  *string `json:"heading_font"`
	MonoFont     *string `json:"mono_font"`
	BaseFontSize string  `json:"base_font_size"`

	UpdatedAt time.Time `json:"updated_at"`
}

type HealthCheck struct {
	ID          string          `json:"id"`
	ServiceName string          `json:"service_name"`
	Status      string          `json:"status"`
	LatencyMs   *int            `json:"latency_ms"`
	Details     json.RawMessage `json:"details"`
	CheckedAt   time.Time       `json:"checked_at"`
}

type WaitlistEntry struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      *string   `json:"name"`
	Source    *string   `json:"source"`
	CreatedAt time.Time `json:"created_at"`
}

// CronJob represents a scheduled background job
type CronJob struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Schedule  string     `json:"schedule"`
	Handler   string     `json:"handler"`
	IsActive  bool       `json:"is_active"`
	LastRunAt *time.Time `json:"last_run_at"`
	NextRunAt *time.Time `json:"next_run_at"`
	CreatedAt time.Time  `json:"created_at"`
}

// JobExecution represents a single run of a cron job
type JobExecution struct {
	ID         string          `json:"id"`
	JobID      string          `json:"job_id"`
	Status     string          `json:"status"` // running, success, failed
	StartedAt  time.Time       `json:"started_at"`
	FinishedAt *time.Time      `json:"finished_at"`
	DurationMs *int            `json:"duration_ms"`
	Error      *string         `json:"error"`
	Output     json.RawMessage `json:"output"`
}

// AdminTool represents a link to an external admin tool
type AdminTool struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	URL         string  `json:"url"`
	Icon        *string `json:"icon"`
	Category    *string `json:"category"`
	IsActive    bool    `json:"is_active"`
	MinRole     string  `json:"min_role"`
	OrderIndex  int     `json:"order_index"`
}

// AuditLog represents an admin action audit entry
type AuditLog struct {
	ID         string          `json:"id"`
	UserID     *string         `json:"user_id"`
	Action     string          `json:"action"`
	Resource   *string         `json:"resource"`
	ResourceID *string         `json:"resource_id"`
	Details    json.RawMessage `json:"details"`
	IPAddress  *string         `json:"ip_address"`
	CreatedAt  time.Time       `json:"created_at"`
}

// AppLog represents an application log entry
type AppLog struct {
	ID        int64           `json:"id"`
	Level     string          `json:"level"`
	Message   string          `json:"message"`
	Source    *string         `json:"source"`
	Metadata  json.RawMessage `json:"metadata"`
	CreatedAt time.Time       `json:"created_at"`
}

// LogConfig represents log retention configuration
type LogConfig struct {
	ID            int    `json:"id"`
	RetentionDays int    `json:"retention_days"`
	MinLevel      string `json:"min_level"`
}

// PixConfig holds the admin's PIX payment configuration
type PixConfig struct {
	ID          int       `json:"id"`
	PixKey      string    `json:"pix_key"`
	KeyType     string    `json:"key_type"` // cpf, email, phone, random
	Beneficiary string    `json:"beneficiary"`
	City        string    `json:"city"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// LegalPage represents a legal page (Terms, Privacy, etc.)
type LegalPage struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	IsActive  bool      `json:"is_active"`
	UpdatedAt time.Time `json:"updated_at"`
}
