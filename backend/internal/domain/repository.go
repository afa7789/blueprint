package domain

import (
	"context"
	"time"
)

type UserRepository interface {
	FindByID(ctx context.Context, id string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, u *User) error
	Update(ctx context.Context, u *User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, offset, limit int) ([]User, int, error)
	VerifyEmail(ctx context.Context, userID string) error
	UpdateStripeCustomerID(ctx context.Context, userID, customerID string) error
}

type SecuritySettingRepository interface {
	GetAll(ctx context.Context) ([]SecuritySetting, error)
	GetByKey(ctx context.Context, key string) (*SecuritySetting, error)
	Update(ctx context.Context, key, value string) error
}

type FeatureFlagRepository interface {
	GetAll(ctx context.Context) ([]FeatureFlag, error)
	GetByKey(ctx context.Context, key string) (*FeatureFlag, error)
	Set(ctx context.Context, key string, enabled bool) error
}

type WaitlistRepository interface {
	Add(ctx context.Context, entry *WaitlistEntry) error
	ExistsByEmail(ctx context.Context, email string) (bool, error)
	List(ctx context.Context) ([]WaitlistEntry, error)
}

type ProductRepository interface {
	FindByID(ctx context.Context, id string) (*Product, error)
	List(ctx context.Context, categoryID *string, activeOnly bool, offset, limit int) ([]Product, int, error)
	Create(ctx context.Context, p *Product) error
	Update(ctx context.Context, p *Product) error
	Delete(ctx context.Context, id string) error
}

type CategoryRepository interface {
	List(ctx context.Context) ([]Category, error)
	Create(ctx context.Context, c *Category) error
	Update(ctx context.Context, c *Category) error
	Delete(ctx context.Context, id string) error
}

type OrderRepository interface {
	FindByID(ctx context.Context, id string) (*Order, error)
	FindByUser(ctx context.Context, userID string, offset, limit int) ([]Order, int, error)
	FindByPaymentID(ctx context.Context, paymentID string) (*Order, error)
	Create(ctx context.Context, o *Order, items []OrderItem) error
	UpdateStatus(ctx context.Context, id, status string) error
	UpdatePayment(ctx context.Context, id, paymentMethod, paymentID string) error
	ListByStatus(ctx context.Context, status string) ([]Order, error)
	AddTrackingCode(ctx context.Context, id, code string) error
	UpdateReceiptURL(ctx context.Context, id, url string) error
}

type CouponRepository interface {
	FindByCode(ctx context.Context, code string) (*Coupon, error)
	Create(ctx context.Context, c *Coupon) error
	IncrementUsed(ctx context.Context, id string) error
	List(ctx context.Context) ([]Coupon, error)
}

type BlogRepository interface {
	FindBySlug(ctx context.Context, slug string) (*BlogPost, error)
	FindByID(ctx context.Context, id string) (*BlogPost, error)
	List(ctx context.Context, status string, offset, limit int) ([]BlogPost, int, error)
	Create(ctx context.Context, p *BlogPost) error
	Update(ctx context.Context, p *BlogPost) error
	Delete(ctx context.Context, id string) error
}

type BannerRepository interface {
	List(ctx context.Context, activeOnly bool) ([]Banner, error)
	Create(ctx context.Context, b *Banner) error
	Update(ctx context.Context, b *Banner) error
	Delete(ctx context.Context, id string) error
}

type LinktreeRepository interface {
	List(ctx context.Context, activeOnly bool) ([]LinktreeItem, error)
	Create(ctx context.Context, item *LinktreeItem) error
	Update(ctx context.Context, item *LinktreeItem) error
	Delete(ctx context.Context, id string) error
	Reorder(ctx context.Context, items []LinktreeItem) error
}

type BrandKitRepository interface {
	Get(ctx context.Context) (*BrandKit, error)
	Upsert(ctx context.Context, bk *BrandKit) error
}

type EmailGroupRepository interface {
	List(ctx context.Context) ([]EmailGroup, error)
	Create(ctx context.Context, g *EmailGroup) error
	Delete(ctx context.Context, id string) error
}

type EmailSubscriptionRepository interface {
	Add(ctx context.Context, sub *EmailSubscription) error
	ListByGroup(ctx context.Context, groupID string) ([]EmailSubscription, error)
	Deactivate(ctx context.Context, email string) error
}

type UserGroupRepository interface {
	List(ctx context.Context) ([]UserGroup, error)
	Create(ctx context.Context, g *UserGroup) error
	Delete(ctx context.Context, id string) error
}

type CronJobRepository interface {
	List(ctx context.Context) ([]CronJob, error)
	FindByID(ctx context.Context, id string) (*CronJob, error)
	Create(ctx context.Context, job *CronJob) error
	Update(ctx context.Context, job *CronJob) error
	Delete(ctx context.Context, id string) error
	UpdateLastRun(ctx context.Context, id string, lastRun time.Time, nextRun *time.Time) error
}

type JobExecutionRepository interface {
	Create(ctx context.Context, exec *JobExecution) error
	Update(ctx context.Context, exec *JobExecution) error
	FindByID(ctx context.Context, id string) (*JobExecution, error)
	ListByJob(ctx context.Context, jobID string, offset, limit int) ([]JobExecution, int, error)
}

type AdminToolRepository interface {
	List(ctx context.Context, activeOnly bool) ([]AdminTool, error)
	Create(ctx context.Context, tool *AdminTool) error
	Update(ctx context.Context, tool *AdminTool) error
	Delete(ctx context.Context, id string) error
}

type AuditLogRepository interface {
	Create(ctx context.Context, log *AuditLog) error
	List(ctx context.Context, userID, action, resource *string, from, to *time.Time, offset, limit int) ([]AuditLog, int, error)
}

type AppLogRepository interface {
	Create(ctx context.Context, log *AppLog) error
	List(ctx context.Context, level, source, search *string, from, to *time.Time, offset, limit int) ([]AppLog, int, error)
	LatestSince(ctx context.Context, sinceID int64, limit int) ([]AppLog, error)
	Cleanup(ctx context.Context, olderThanDays int) (int64, error)
}

type LogConfigRepository interface {
	Get(ctx context.Context) (*LogConfig, error)
	Update(ctx context.Context, cfg *LogConfig) error
}

type LegalPageRepository interface {
	FindBySlug(ctx context.Context, slug string) (*LegalPage, error)
	List(ctx context.Context, activeOnly bool) ([]LegalPage, error)
	Create(ctx context.Context, page *LegalPage) error
	Update(ctx context.Context, page *LegalPage) error
	Delete(ctx context.Context, id string) error
}

type UserProfileRepository interface {
	FindByUserID(ctx context.Context, userID string) (*UserProfile, error)
	Upsert(ctx context.Context, p *UserProfile) error
}

type PixConfigRepository interface {
	Get(ctx context.Context) (*PixConfig, error)
	Update(ctx context.Context, cfg *PixConfig) error
}
