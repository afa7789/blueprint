package testutil

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/afa/blueprint/backend/internal/domain"
)

// ---- MockUserRepo ----

type MockUserRepo struct {
	mu    sync.RWMutex
	users map[string]*domain.User // keyed by ID
}

func NewMockUserRepo() *MockUserRepo {
	return &MockUserRepo{users: make(map[string]*domain.User)}
}

func (m *MockUserRepo) FindByID(_ context.Context, id string) (*domain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.users[id]
	if !ok {
		return nil, errors.New("not found")
	}
	copy := *u
	return &copy, nil
}

func (m *MockUserRepo) FindByEmail(_ context.Context, email string) (*domain.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, u := range m.users {
		if u.Email == email {
			copy := *u
			return &copy, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *MockUserRepo) Create(_ context.Context, u *domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, existing := range m.users {
		if existing.Email == u.Email {
			return errors.New("duplicate email")
		}
	}
	now := time.Now()
	u.CreatedAt = now
	u.UpdatedAt = now
	copy := *u
	m.users[u.ID] = &copy
	return nil
}

func (m *MockUserRepo) Update(_ context.Context, u *domain.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.users[u.ID]; !ok {
		return errors.New("not found")
	}
	u.UpdatedAt = time.Now()
	copy := *u
	m.users[u.ID] = &copy
	return nil
}

func (m *MockUserRepo) Delete(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.users, id)
	return nil
}

func (m *MockUserRepo) List(_ context.Context, offset, limit int) ([]domain.User, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	all := make([]domain.User, 0, len(m.users))
	for _, u := range m.users {
		all = append(all, *u)
	}
	total := len(all)
	if offset >= total {
		return []domain.User{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return all[offset:end], total, nil
}

func (m *MockUserRepo) VerifyEmail(_ context.Context, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[userID]
	if !ok {
		return errors.New("not found")
	}
	u.EmailVerified = true
	now := time.Now()
	u.EmailVerifiedAt = &now
	return nil
}

func (m *MockUserRepo) UpdateStripeCustomerID(_ context.Context, userID, customerID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.users[userID]
	if !ok {
		return errors.New("not found")
	}
	u.StripeCustomerID = &customerID
	return nil
}

// ---- MockFeatureFlagRepo ----

type MockFeatureFlagRepo struct {
	mu    sync.RWMutex
	flags map[string]*domain.FeatureFlag
}

func NewMockFeatureFlagRepo() *MockFeatureFlagRepo {
	return &MockFeatureFlagRepo{flags: make(map[string]*domain.FeatureFlag)}
}

func (m *MockFeatureFlagRepo) GetAll(_ context.Context) ([]domain.FeatureFlag, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	all := make([]domain.FeatureFlag, 0, len(m.flags))
	for _, f := range m.flags {
		all = append(all, *f)
	}
	return all, nil
}

func (m *MockFeatureFlagRepo) GetByKey(_ context.Context, key string) (*domain.FeatureFlag, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	f, ok := m.flags[key]
	if !ok {
		return nil, errors.New("not found")
	}
	copy := *f
	return &copy, nil
}

func (m *MockFeatureFlagRepo) Set(_ context.Context, key string, enabled bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if f, ok := m.flags[key]; ok {
		f.Enabled = enabled
	} else {
		m.flags[key] = &domain.FeatureFlag{Key: key, Enabled: enabled, UpdatedAt: time.Now()}
	}
	return nil
}

// ---- MockWaitlistRepo ----

type MockWaitlistRepo struct {
	mu      sync.RWMutex
	entries []domain.WaitlistEntry
}

func NewMockWaitlistRepo() *MockWaitlistRepo {
	return &MockWaitlistRepo{}
}

func (m *MockWaitlistRepo) Add(_ context.Context, entry *domain.WaitlistEntry) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	entry.CreatedAt = time.Now()
	m.entries = append(m.entries, *entry)
	return nil
}

func (m *MockWaitlistRepo) ExistsByEmail(_ context.Context, email string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, e := range m.entries {
		if e.Email == email {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockWaitlistRepo) List(_ context.Context) ([]domain.WaitlistEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]domain.WaitlistEntry, len(m.entries))
	copy(result, m.entries)
	return result, nil
}

// ---- MockProductRepo ----

type MockProductRepo struct {
	mu       sync.RWMutex
	products map[string]*domain.Product
}

func NewMockProductRepo() *MockProductRepo {
	return &MockProductRepo{products: make(map[string]*domain.Product)}
}

func (m *MockProductRepo) FindByID(_ context.Context, id string) (*domain.Product, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.products[id]
	if !ok {
		return nil, errors.New("not found")
	}
	copy := *p
	return &copy, nil
}

func (m *MockProductRepo) List(_ context.Context, categoryID *string, activeOnly bool, offset, limit int) ([]domain.Product, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []domain.Product
	for _, p := range m.products {
		if activeOnly && !p.IsActive {
			continue
		}
		if categoryID != nil && (p.CategoryID == nil || *p.CategoryID != *categoryID) {
			continue
		}
		result = append(result, *p)
	}
	total := len(result)
	if offset >= total {
		return []domain.Product{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return result[offset:end], total, nil
}

func (m *MockProductRepo) Create(_ context.Context, p *domain.Product) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now
	copy := *p
	m.products[p.ID] = &copy
	return nil
}

func (m *MockProductRepo) Update(_ context.Context, p *domain.Product) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.products[p.ID]; !ok {
		return errors.New("not found")
	}
	p.UpdatedAt = time.Now()
	copy := *p
	m.products[p.ID] = &copy
	return nil
}

func (m *MockProductRepo) Delete(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.products, id)
	return nil
}

// ---- MockCategoryRepo ----

type MockCategoryRepo struct {
	mu         sync.RWMutex
	categories map[string]*domain.Category
}

func NewMockCategoryRepo() *MockCategoryRepo {
	return &MockCategoryRepo{categories: make(map[string]*domain.Category)}
}

func (m *MockCategoryRepo) List(_ context.Context) ([]domain.Category, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]domain.Category, 0, len(m.categories))
	for _, c := range m.categories {
		result = append(result, *c)
	}
	return result, nil
}

func (m *MockCategoryRepo) Create(_ context.Context, c *domain.Category) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	copy := *c
	m.categories[c.ID] = &copy
	return nil
}

func (m *MockCategoryRepo) Update(_ context.Context, c *domain.Category) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.categories[c.ID]; !ok {
		return errors.New("not found")
	}
	copy := *c
	m.categories[c.ID] = &copy
	return nil
}

func (m *MockCategoryRepo) Delete(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.categories, id)
	return nil
}

// ---- MockOrderRepo ----

type MockOrderRepo struct {
	mu     sync.RWMutex
	orders map[string]*domain.Order
	items  map[string][]domain.OrderItem
}

func NewMockOrderRepo() *MockOrderRepo {
	return &MockOrderRepo{
		orders: make(map[string]*domain.Order),
		items:  make(map[string][]domain.OrderItem),
	}
}

func (m *MockOrderRepo) FindByID(_ context.Context, id string) (*domain.Order, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	o, ok := m.orders[id]
	if !ok {
		return nil, errors.New("not found")
	}
	copy := *o
	return &copy, nil
}

func (m *MockOrderRepo) FindByUser(_ context.Context, userID string, offset, limit int) ([]domain.Order, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []domain.Order
	for _, o := range m.orders {
		if o.UserID != nil && *o.UserID == userID {
			result = append(result, *o)
		}
	}
	total := len(result)
	if offset >= total {
		return []domain.Order{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return result[offset:end], total, nil
}

func (m *MockOrderRepo) FindByPaymentID(_ context.Context, paymentID string) (*domain.Order, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, o := range m.orders {
		if o.PaymentID != nil && *o.PaymentID == paymentID {
			copy := *o
			return &copy, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *MockOrderRepo) Create(_ context.Context, o *domain.Order, orderItems []domain.OrderItem) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	o.CreatedAt = now
	o.UpdatedAt = now
	copy := *o
	m.orders[o.ID] = &copy
	m.items[o.ID] = orderItems
	return nil
}

func (m *MockOrderRepo) UpdateStatus(_ context.Context, id, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	o, ok := m.orders[id]
	if !ok {
		return errors.New("not found")
	}
	o.Status = status
	return nil
}

func (m *MockOrderRepo) UpdatePayment(_ context.Context, id, paymentMethod, paymentID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	o, ok := m.orders[id]
	if !ok {
		return errors.New("not found")
	}
	o.PaymentMethod = &paymentMethod
	o.PaymentID = &paymentID
	return nil
}

func (m *MockOrderRepo) ListByStatus(_ context.Context, status string) ([]domain.Order, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []domain.Order
	for _, o := range m.orders {
		if status == "" || o.Status == status {
			result = append(result, *o)
		}
	}
	return result, nil
}

func (m *MockOrderRepo) AddTrackingCode(_ context.Context, id, code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	o, ok := m.orders[id]
	if !ok {
		return errors.New("not found")
	}
	o.TrackingCode = &code
	return nil
}

func (m *MockOrderRepo) UpdateReceiptURL(_ context.Context, id, url string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	o, ok := m.orders[id]
	if !ok {
		return errors.New("not found")
	}
	o.ReceiptURL = &url
	return nil
}

// ---- MockCouponRepo ----

type MockCouponRepo struct {
	mu      sync.RWMutex
	coupons map[string]*domain.Coupon // keyed by code
}

func NewMockCouponRepo() *MockCouponRepo {
	return &MockCouponRepo{coupons: make(map[string]*domain.Coupon)}
}

func (m *MockCouponRepo) FindByCode(_ context.Context, code string) (*domain.Coupon, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.coupons[code]
	if !ok {
		return nil, errors.New("not found")
	}
	copy := *c
	return &copy, nil
}

func (m *MockCouponRepo) Create(_ context.Context, c *domain.Coupon) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	copy := *c
	m.coupons[c.Code] = &copy
	return nil
}

func (m *MockCouponRepo) IncrementUsed(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, c := range m.coupons {
		if c.ID == id {
			c.UsedCount++
			return nil
		}
	}
	return errors.New("not found")
}

func (m *MockCouponRepo) List(_ context.Context) ([]domain.Coupon, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]domain.Coupon, 0, len(m.coupons))
	for _, c := range m.coupons {
		result = append(result, *c)
	}
	return result, nil
}

// ---- MockBlogRepo ----

type MockBlogRepo struct {
	mu    sync.RWMutex
	posts map[string]*domain.BlogPost // keyed by ID
}

func NewMockBlogRepo() *MockBlogRepo {
	return &MockBlogRepo{posts: make(map[string]*domain.BlogPost)}
}

func (m *MockBlogRepo) FindBySlug(_ context.Context, slug string) (*domain.BlogPost, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, p := range m.posts {
		if p.Slug == slug {
			copy := *p
			return &copy, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *MockBlogRepo) FindByID(_ context.Context, id string) (*domain.BlogPost, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.posts[id]
	if !ok {
		return nil, errors.New("not found")
	}
	copy := *p
	return &copy, nil
}

func (m *MockBlogRepo) List(_ context.Context, status string, offset, limit int) ([]domain.BlogPost, int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	all := make([]domain.BlogPost, 0)
	for _, p := range m.posts {
		if status == "" || p.Status == status {
			all = append(all, *p)
		}
	}
	total := len(all)
	if offset >= total {
		return []domain.BlogPost{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return all[offset:end], total, nil
}

func (m *MockBlogRepo) Create(_ context.Context, p *domain.BlogPost) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if p.ID == "" {
		p.ID = "post-" + time.Now().Format("20060102150405")
	}
	p.CreatedAt = time.Now()
	copy := *p
	m.posts[p.ID] = &copy
	return nil
}

func (m *MockBlogRepo) Update(_ context.Context, p *domain.BlogPost) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.posts[p.ID]; !ok {
		return errors.New("not found")
	}
	copy := *p
	m.posts[p.ID] = &copy
	return nil
}

func (m *MockBlogRepo) Delete(_ context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.posts, id)
	return nil
}
