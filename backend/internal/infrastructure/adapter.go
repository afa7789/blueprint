package infrastructure

import (
	"context"
	"time"

	"github.com/afa/blueprint/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ---- User ----

type userRepo struct{ pool *pgxpool.Pool }

func NewUserRepo(pool *pgxpool.Pool) domain.UserRepository { return &userRepo{pool} }

func (r *userRepo) FindByID(ctx context.Context, id string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id,email,password_hash,name,role,created_at,updated_at FROM users WHERE id=$1`, id)
	u := &domain.User{}
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}
	return u, nil
}

func (r *userRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id,email,password_hash,name,role,created_at,updated_at FROM users WHERE email=$1`, email)
	u := &domain.User{}
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}
	return u, nil
}

func (r *userRepo) Create(ctx context.Context, u *domain.User) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO users(email,password_hash,name,role) VALUES($1,$2,$3,$4) RETURNING id,created_at,updated_at`,
		u.Email, u.PasswordHash, u.Name, u.Role,
	).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

func (r *userRepo) Update(ctx context.Context, u *domain.User) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE users SET email=$1,password_hash=$2,name=$3,role=$4,updated_at=NOW() WHERE id=$5`,
		u.Email, u.PasswordHash, u.Name, u.Role, u.ID)
	return err
}

func (r *userRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id=$1`, id)
	return err
}

func (r *userRepo) List(ctx context.Context, offset, limit int) ([]domain.User, int, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id,email,password_hash,name,role,created_at,updated_at FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset)
	if err != nil {
		return nil, 0, err
	}
	users, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.User, error) {
		var u domain.User
		return u, row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	})
	if err != nil {
		return nil, 0, err
	}
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&total); err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

// ---- FeatureFlag ----

type featureFlagRepo struct{ pool *pgxpool.Pool }

func NewFeatureFlagRepo(pool *pgxpool.Pool) domain.FeatureFlagRepository { return &featureFlagRepo{pool} }

func (r *featureFlagRepo) GetAll(ctx context.Context) ([]domain.FeatureFlag, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,key,enabled,updated_at FROM feature_flags`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.FeatureFlag, error) {
		var f domain.FeatureFlag
		return f, row.Scan(&f.ID, &f.Key, &f.Enabled, &f.UpdatedAt)
	})
}

func (r *featureFlagRepo) GetByKey(ctx context.Context, key string) (*domain.FeatureFlag, error) {
	f := &domain.FeatureFlag{}
	err := r.pool.QueryRow(ctx, `SELECT id,key,enabled,updated_at FROM feature_flags WHERE key=$1`, key).
		Scan(&f.ID, &f.Key, &f.Enabled, &f.UpdatedAt)
	return f, err
}

func (r *featureFlagRepo) Set(ctx context.Context, key string, enabled bool) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO feature_flags(key,enabled) VALUES($1,$2) ON CONFLICT(key) DO UPDATE SET enabled=$2,updated_at=NOW()`,
		key, enabled)
	return err
}

// ---- Waitlist ----

type waitlistRepo struct{ pool *pgxpool.Pool }

func NewWaitlistRepo(pool *pgxpool.Pool) domain.WaitlistRepository { return &waitlistRepo{pool} }

func (r *waitlistRepo) Add(ctx context.Context, entry *domain.WaitlistEntry) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO waitlist(email,name,source) VALUES($1,$2,$3) RETURNING id,created_at`,
		entry.Email, entry.Name, entry.Source,
	).Scan(&entry.ID, &entry.CreatedAt)
}

func (r *waitlistRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM waitlist WHERE email=$1)`, email).Scan(&exists)
	return exists, err
}

func (r *waitlistRepo) List(ctx context.Context) ([]domain.WaitlistEntry, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,email,name,source,created_at FROM waitlist ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.WaitlistEntry, error) {
		var e domain.WaitlistEntry
		return e, row.Scan(&e.ID, &e.Email, &e.Name, &e.Source, &e.CreatedAt)
	})
}

// ---- Product ----

type productRepo struct{ pool *pgxpool.Pool }

func NewProductRepo(pool *pgxpool.Pool) domain.ProductRepository { return &productRepo{pool} }

func (r *productRepo) FindByID(ctx context.Context, id string) (*domain.Product, error) {
	p := &domain.Product{}
	err := r.pool.QueryRow(ctx,
		`SELECT id,name,description,price,stock,is_pre_sale,pre_sale_available_at,images,category_id,is_active,created_at,updated_at FROM products WHERE id=$1`, id).
		Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.IsPreSale, &p.PreSaleAvailableAt, &p.Images, &p.CategoryID, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

func (r *productRepo) List(ctx context.Context, categoryID *string, activeOnly bool, offset, limit int) ([]domain.Product, int, error) {
	query := `SELECT id,name,description,price,stock,is_pre_sale,pre_sale_available_at,images,category_id,is_active,created_at,updated_at FROM products WHERE 1=1`
	args := []any{}
	n := 1
	if categoryID != nil {
		query += ` AND category_id=$` + itoa(n)
		args = append(args, *categoryID)
		n++
	}
	if activeOnly {
		query += ` AND is_active=true`
	}
	countQuery := `SELECT COUNT(*) FROM products WHERE 1=1`
	if categoryID != nil {
		countQuery += ` AND category_id=$1`
	}
	if activeOnly {
		countQuery += ` AND is_active=true`
	}

	query += ` ORDER BY created_at DESC LIMIT $` + itoa(n) + ` OFFSET $` + itoa(n+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	products, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.Product, error) {
		var p domain.Product
		return p, row.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.Stock, &p.IsPreSale, &p.PreSaleAvailableAt, &p.Images, &p.CategoryID, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	})
	if err != nil {
		return nil, 0, err
	}
	var countArgs []any
	if categoryID != nil {
		countArgs = append(countArgs, *categoryID)
	}
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}
	return products, total, nil
}

func (r *productRepo) Create(ctx context.Context, p *domain.Product) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO products(name,description,price,stock,is_pre_sale,pre_sale_available_at,images,category_id,is_active)
		 VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id,created_at,updated_at`,
		p.Name, p.Description, p.Price, p.Stock, p.IsPreSale, p.PreSaleAvailableAt, p.Images, p.CategoryID, p.IsActive,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *productRepo) Update(ctx context.Context, p *domain.Product) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE products SET name=$1,description=$2,price=$3,stock=$4,is_pre_sale=$5,pre_sale_available_at=$6,images=$7,category_id=$8,is_active=$9,updated_at=NOW() WHERE id=$10`,
		p.Name, p.Description, p.Price, p.Stock, p.IsPreSale, p.PreSaleAvailableAt, p.Images, p.CategoryID, p.IsActive, p.ID)
	return err
}

func (r *productRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM products WHERE id=$1`, id)
	return err
}

// ---- Category ----

type categoryRepo struct{ pool *pgxpool.Pool }

func NewCategoryRepo(pool *pgxpool.Pool) domain.CategoryRepository { return &categoryRepo{pool} }

func (r *categoryRepo) List(ctx context.Context) ([]domain.Category, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,name,slug,parent_id FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.Category, error) {
		var c domain.Category
		return c, row.Scan(&c.ID, &c.Name, &c.Slug, &c.ParentID)
	})
}

func (r *categoryRepo) Create(ctx context.Context, c *domain.Category) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO categories(name,slug,parent_id) VALUES($1,$2,$3) RETURNING id`,
		c.Name, c.Slug, c.ParentID,
	).Scan(&c.ID)
}

func (r *categoryRepo) Update(ctx context.Context, c *domain.Category) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE categories SET name=$1,slug=$2,parent_id=$3 WHERE id=$4`,
		c.Name, c.Slug, c.ParentID, c.ID)
	return err
}

func (r *categoryRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM categories WHERE id=$1`, id)
	return err
}

// ---- Order ----

type orderRepo struct{ pool *pgxpool.Pool }

func NewOrderRepo(pool *pgxpool.Pool) domain.OrderRepository { return &orderRepo{pool} }

func (r *orderRepo) FindByID(ctx context.Context, id string) (*domain.Order, error) {
	o := &domain.Order{}
	err := r.pool.QueryRow(ctx,
		`SELECT id,user_id,status,total,payment_method,payment_id,shipping_address,tracking_code,created_at,updated_at FROM orders WHERE id=$1`, id).
		Scan(&o.ID, &o.UserID, &o.Status, &o.Total, &o.PaymentMethod, &o.PaymentID, &o.ShippingAddress, &o.TrackingCode, &o.CreatedAt, &o.UpdatedAt)
	return o, err
}

func (r *orderRepo) FindByUser(ctx context.Context, userID string, offset, limit int) ([]domain.Order, int, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id,user_id,status,total,payment_method,payment_id,shipping_address,tracking_code,created_at,updated_at FROM orders WHERE user_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
		userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	orders, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.Order, error) {
		var o domain.Order
		return o, row.Scan(&o.ID, &o.UserID, &o.Status, &o.Total, &o.PaymentMethod, &o.PaymentID, &o.ShippingAddress, &o.TrackingCode, &o.CreatedAt, &o.UpdatedAt)
	})
	if err != nil {
		return nil, 0, err
	}
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM orders WHERE user_id=$1`, userID).Scan(&total); err != nil {
		return nil, 0, err
	}
	return orders, total, nil
}

func (r *orderRepo) Create(ctx context.Context, o *domain.Order, items []domain.OrderItem) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	if err := tx.QueryRow(ctx,
		`INSERT INTO orders(user_id,status,total,payment_method,payment_id,shipping_address) VALUES($1,$2,$3,$4,$5,$6) RETURNING id,created_at,updated_at`,
		o.UserID, o.Status, o.Total, o.PaymentMethod, o.PaymentID, o.ShippingAddress,
	).Scan(&o.ID, &o.CreatedAt, &o.UpdatedAt); err != nil {
		return err
	}

	for i := range items {
		items[i].OrderID = o.ID
		if err := tx.QueryRow(ctx,
			`INSERT INTO order_items(order_id,product_id,quantity,unit_price) VALUES($1,$2,$3,$4) RETURNING id`,
			items[i].OrderID, items[i].ProductID, items[i].Quantity, items[i].UnitPrice,
		).Scan(&items[i].ID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *orderRepo) UpdateStatus(ctx context.Context, id, status string) error {
	_, err := r.pool.Exec(ctx, `UPDATE orders SET status=$1,updated_at=NOW() WHERE id=$2`, status, id)
	return err
}

func (r *orderRepo) ListByStatus(ctx context.Context, status string) ([]domain.Order, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id,user_id,status,total,payment_method,payment_id,shipping_address,tracking_code,created_at,updated_at FROM orders WHERE status=$1 ORDER BY created_at DESC`,
		status)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.Order, error) {
		var o domain.Order
		return o, row.Scan(&o.ID, &o.UserID, &o.Status, &o.Total, &o.PaymentMethod, &o.PaymentID, &o.ShippingAddress, &o.TrackingCode, &o.CreatedAt, &o.UpdatedAt)
	})
}

func (r *orderRepo) FindByPaymentID(ctx context.Context, paymentID string) (*domain.Order, error) {
	o := &domain.Order{}
	err := r.pool.QueryRow(ctx,
		`SELECT id,user_id,status,total,payment_method,payment_id,shipping_address,tracking_code,created_at,updated_at FROM orders WHERE payment_id=$1`, paymentID).
		Scan(&o.ID, &o.UserID, &o.Status, &o.Total, &o.PaymentMethod, &o.PaymentID, &o.ShippingAddress, &o.TrackingCode, &o.CreatedAt, &o.UpdatedAt)
	return o, err
}

func (r *orderRepo) UpdatePayment(ctx context.Context, id, paymentMethod, paymentID string) error {
	_, err := r.pool.Exec(ctx, `UPDATE orders SET payment_method=$1,payment_id=$2,updated_at=NOW() WHERE id=$3`, paymentMethod, paymentID, id)
	return err
}

func (r *orderRepo) AddTrackingCode(ctx context.Context, id, code string) error {
	_, err := r.pool.Exec(ctx, `UPDATE orders SET tracking_code=$1,updated_at=NOW() WHERE id=$2`, code, id)
	return err
}

// ---- Coupon ----

type couponRepo struct{ pool *pgxpool.Pool }

func NewCouponRepo(pool *pgxpool.Pool) domain.CouponRepository { return &couponRepo{pool} }

func (r *couponRepo) FindByCode(ctx context.Context, code string) (*domain.Coupon, error) {
	c := &domain.Coupon{}
	err := r.pool.QueryRow(ctx,
		`SELECT id,code,discount_type,discount_value,min_purchase,valid_from,valid_until,max_uses,used_count,is_active FROM coupons WHERE code=$1`, code).
		Scan(&c.ID, &c.Code, &c.DiscountType, &c.DiscountValue, &c.MinPurchase, &c.ValidFrom, &c.ValidUntil, &c.MaxUses, &c.UsedCount, &c.IsActive)
	return c, err
}

func (r *couponRepo) Create(ctx context.Context, c *domain.Coupon) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO coupons(code,discount_type,discount_value,min_purchase,valid_from,valid_until,max_uses,is_active) VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`,
		c.Code, c.DiscountType, c.DiscountValue, c.MinPurchase, c.ValidFrom, c.ValidUntil, c.MaxUses, c.IsActive,
	).Scan(&c.ID)
}

func (r *couponRepo) IncrementUsed(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE coupons SET used_count=used_count+1 WHERE id=$1`, id)
	return err
}

func (r *couponRepo) List(ctx context.Context) ([]domain.Coupon, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id,code,discount_type,discount_value,min_purchase,valid_from,valid_until,max_uses,used_count,is_active FROM coupons ORDER BY code`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.Coupon, error) {
		var c domain.Coupon
		return c, row.Scan(&c.ID, &c.Code, &c.DiscountType, &c.DiscountValue, &c.MinPurchase, &c.ValidFrom, &c.ValidUntil, &c.MaxUses, &c.UsedCount, &c.IsActive)
	})
}

// ---- Blog ----

type blogRepo struct{ pool *pgxpool.Pool }

func NewBlogRepo(pool *pgxpool.Pool) domain.BlogRepository { return &blogRepo{pool} }

func (r *blogRepo) FindBySlug(ctx context.Context, slug string) (*domain.BlogPost, error) {
	p := &domain.BlogPost{}
	err := r.pool.QueryRow(ctx,
		`SELECT id,title,slug,content,excerpt,cover_image,status,author_id,created_at,published_at,metadata FROM blog_posts WHERE slug=$1`, slug).
		Scan(&p.ID, &p.Title, &p.Slug, &p.Content, &p.Excerpt, &p.CoverImage, &p.Status, &p.AuthorID, &p.CreatedAt, &p.PublishedAt, &p.Metadata)
	return p, err
}

func (r *blogRepo) FindByID(ctx context.Context, id string) (*domain.BlogPost, error) {
	p := &domain.BlogPost{}
	err := r.pool.QueryRow(ctx,
		`SELECT id,title,slug,content,excerpt,cover_image,status,author_id,created_at,published_at,metadata FROM blog_posts WHERE id=$1`, id).
		Scan(&p.ID, &p.Title, &p.Slug, &p.Content, &p.Excerpt, &p.CoverImage, &p.Status, &p.AuthorID, &p.CreatedAt, &p.PublishedAt, &p.Metadata)
	return p, err
}

func (r *blogRepo) List(ctx context.Context, status string, offset, limit int) ([]domain.BlogPost, int, error) {
	collectRow := func(row pgx.CollectableRow) (domain.BlogPost, error) {
		var p domain.BlogPost
		return p, row.Scan(&p.ID, &p.Title, &p.Slug, &p.Content, &p.Excerpt, &p.CoverImage, &p.Status, &p.AuthorID, &p.CreatedAt, &p.PublishedAt, &p.Metadata)
	}

	var posts []domain.BlogPost
	var total int

	if status != "" {
		rows, err := r.pool.Query(ctx,
			`SELECT id,title,slug,content,excerpt,cover_image,status,author_id,created_at,published_at,metadata FROM blog_posts WHERE status=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`,
			status, limit, offset)
		if err != nil {
			return nil, 0, err
		}
		posts, err = pgx.CollectRows(rows, collectRow)
		if err != nil {
			return nil, 0, err
		}
		if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM blog_posts WHERE status=$1`, status).Scan(&total); err != nil {
			return nil, 0, err
		}
	} else {
		rows, err := r.pool.Query(ctx,
			`SELECT id,title,slug,content,excerpt,cover_image,status,author_id,created_at,published_at,metadata FROM blog_posts ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
			limit, offset)
		if err != nil {
			return nil, 0, err
		}
		posts, err = pgx.CollectRows(rows, collectRow)
		if err != nil {
			return nil, 0, err
		}
		if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM blog_posts`).Scan(&total); err != nil {
			return nil, 0, err
		}
	}
	return posts, total, nil
}

func (r *blogRepo) Create(ctx context.Context, p *domain.BlogPost) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO blog_posts(title,slug,content,excerpt,cover_image,status,author_id,published_at,metadata) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id,created_at`,
		p.Title, p.Slug, p.Content, p.Excerpt, p.CoverImage, p.Status, p.AuthorID, p.PublishedAt, p.Metadata,
	).Scan(&p.ID, &p.CreatedAt)
}

func (r *blogRepo) Update(ctx context.Context, p *domain.BlogPost) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE blog_posts SET title=$1,slug=$2,content=$3,excerpt=$4,cover_image=$5,status=$6,author_id=$7,published_at=$8,metadata=$9 WHERE id=$10`,
		p.Title, p.Slug, p.Content, p.Excerpt, p.CoverImage, p.Status, p.AuthorID, p.PublishedAt, p.Metadata, p.ID)
	return err
}

func (r *blogRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM blog_posts WHERE id=$1`, id)
	return err
}

// ---- Banner ----

type bannerRepo struct{ pool *pgxpool.Pool }

func NewBannerRepo(pool *pgxpool.Pool) domain.BannerRepository { return &bannerRepo{pool} }

func (r *bannerRepo) List(ctx context.Context, activeOnly bool) ([]domain.Banner, error) {
	query := `SELECT id,title,image_url,link_url,target_profile,is_active,start_date,end_date,display_duration,cache_key,order_index FROM banners`
	if activeOnly {
		query += ` WHERE is_active=true`
	}
	query += ` ORDER BY order_index`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.Banner, error) {
		var b domain.Banner
		return b, row.Scan(&b.ID, &b.Title, &b.ImageURL, &b.LinkURL, &b.TargetProfile, &b.IsActive, &b.StartDate, &b.EndDate, &b.DisplayDuration, &b.CacheKey, &b.OrderIndex)
	})
}

func (r *bannerRepo) Create(ctx context.Context, b *domain.Banner) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO banners(title,image_url,link_url,target_profile,is_active,start_date,end_date,display_duration,cache_key,order_index) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id`,
		b.Title, b.ImageURL, b.LinkURL, b.TargetProfile, b.IsActive, b.StartDate, b.EndDate, b.DisplayDuration, b.CacheKey, b.OrderIndex,
	).Scan(&b.ID)
}

func (r *bannerRepo) Update(ctx context.Context, b *domain.Banner) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE banners SET title=$1,image_url=$2,link_url=$3,target_profile=$4,is_active=$5,start_date=$6,end_date=$7,display_duration=$8,cache_key=$9,order_index=$10 WHERE id=$11`,
		b.Title, b.ImageURL, b.LinkURL, b.TargetProfile, b.IsActive, b.StartDate, b.EndDate, b.DisplayDuration, b.CacheKey, b.OrderIndex, b.ID)
	return err
}

func (r *bannerRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM banners WHERE id=$1`, id)
	return err
}

// ---- Linktree ----

type linktreeRepo struct{ pool *pgxpool.Pool }

func NewLinktreeRepo(pool *pgxpool.Pool) domain.LinktreeRepository { return &linktreeRepo{pool} }

func (r *linktreeRepo) List(ctx context.Context, activeOnly bool) ([]domain.LinktreeItem, error) {
	query := `SELECT id,title,url,icon,order_index,is_active FROM linktree_items`
	if activeOnly {
		query += ` WHERE is_active=true`
	}
	query += ` ORDER BY order_index`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.LinktreeItem, error) {
		var item domain.LinktreeItem
		return item, row.Scan(&item.ID, &item.Title, &item.URL, &item.Icon, &item.OrderIndex, &item.IsActive)
	})
}

func (r *linktreeRepo) Create(ctx context.Context, item *domain.LinktreeItem) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO linktree_items(title,url,icon,order_index,is_active) VALUES($1,$2,$3,$4,$5) RETURNING id`,
		item.Title, item.URL, item.Icon, item.OrderIndex, item.IsActive,
	).Scan(&item.ID)
}

func (r *linktreeRepo) Update(ctx context.Context, item *domain.LinktreeItem) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE linktree_items SET title=$1,url=$2,icon=$3,order_index=$4,is_active=$5 WHERE id=$6`,
		item.Title, item.URL, item.Icon, item.OrderIndex, item.IsActive, item.ID)
	return err
}

func (r *linktreeRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM linktree_items WHERE id=$1`, id)
	return err
}

func (r *linktreeRepo) Reorder(ctx context.Context, items []domain.LinktreeItem) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck
	for _, item := range items {
		if _, err := tx.Exec(ctx, `UPDATE linktree_items SET order_index=$1 WHERE id=$2`, item.OrderIndex, item.ID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// ---- BrandKit ----

type brandKitRepo struct{ pool *pgxpool.Pool }

func NewBrandKitRepo(pool *pgxpool.Pool) domain.BrandKitRepository { return &brandKitRepo{pool} }

func (r *brandKitRepo) Get(ctx context.Context) (*domain.BrandKit, error) {
	bk := &domain.BrandKit{}
	err := r.pool.QueryRow(ctx,
		`SELECT id,primary_color,secondary_color,logo_url,favicon_url,font_family,updated_at FROM brand_kit LIMIT 1`).
		Scan(&bk.ID, &bk.PrimaryColor, &bk.SecondaryColor, &bk.LogoURL, &bk.FaviconURL, &bk.FontFamily, &bk.UpdatedAt)
	return bk, err
}

func (r *brandKitRepo) Upsert(ctx context.Context, bk *domain.BrandKit) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO brand_kit(id,primary_color,secondary_color,logo_url,favicon_url,font_family,updated_at)
		 VALUES($1,$2,$3,$4,$5,$6,NOW())
		 ON CONFLICT(id) DO UPDATE SET primary_color=$2,secondary_color=$3,logo_url=$4,favicon_url=$5,font_family=$6,updated_at=NOW()
		 RETURNING updated_at`,
		bk.ID, bk.PrimaryColor, bk.SecondaryColor, bk.LogoURL, bk.FaviconURL, bk.FontFamily,
	).Scan(&bk.UpdatedAt)
}

// ---- EmailGroup ----

type emailGroupRepo struct{ pool *pgxpool.Pool }

func NewEmailGroupRepo(pool *pgxpool.Pool) domain.EmailGroupRepository { return &emailGroupRepo{pool} }

func (r *emailGroupRepo) List(ctx context.Context) ([]domain.EmailGroup, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,name,description FROM email_groups ORDER BY name`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.EmailGroup, error) {
		var g domain.EmailGroup
		return g, row.Scan(&g.ID, &g.Name, &g.Description)
	})
}

func (r *emailGroupRepo) Create(ctx context.Context, g *domain.EmailGroup) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO email_groups(name,description) VALUES($1,$2) RETURNING id`,
		g.Name, g.Description,
	).Scan(&g.ID)
}

func (r *emailGroupRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM email_groups WHERE id=$1`, id)
	return err
}

// ---- EmailSubscription ----

type emailSubRepo struct{ pool *pgxpool.Pool }

func NewEmailSubscriptionRepo(pool *pgxpool.Pool) domain.EmailSubscriptionRepository {
	return &emailSubRepo{pool}
}

func (r *emailSubRepo) Add(ctx context.Context, sub *domain.EmailSubscription) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO email_subscriptions(email,group_id,is_active) VALUES($1,$2,$3) RETURNING id,subscribed_at`,
		sub.Email, sub.GroupID, sub.IsActive,
	).Scan(&sub.ID, &sub.SubscribedAt)
}

func (r *emailSubRepo) ListByGroup(ctx context.Context, groupID string) ([]domain.EmailSubscription, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id,email,group_id,is_active,subscribed_at FROM email_subscriptions WHERE group_id=$1 ORDER BY subscribed_at DESC`,
		groupID)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.EmailSubscription, error) {
		var s domain.EmailSubscription
		return s, row.Scan(&s.ID, &s.Email, &s.GroupID, &s.IsActive, &s.SubscribedAt)
	})
}

func (r *emailSubRepo) Deactivate(ctx context.Context, email string) error {
	_, err := r.pool.Exec(ctx, `UPDATE email_subscriptions SET is_active=false WHERE email=$1`, email)
	return err
}

// ---- UserGroup ----

type userGroupRepo struct{ pool *pgxpool.Pool }

func NewUserGroupRepo(pool *pgxpool.Pool) domain.UserGroupRepository { return &userGroupRepo{pool} }

func (r *userGroupRepo) List(ctx context.Context) ([]domain.UserGroup, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,name,discount_percentage FROM user_groups ORDER BY name`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.UserGroup, error) {
		var g domain.UserGroup
		return g, row.Scan(&g.ID, &g.Name, &g.DiscountPercentage)
	})
}

func (r *userGroupRepo) Create(ctx context.Context, g *domain.UserGroup) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO user_groups(name,discount_percentage) VALUES($1,$2) RETURNING id`,
		g.Name, g.DiscountPercentage,
	).Scan(&g.ID)
}

func (r *userGroupRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM user_groups WHERE id=$1`, id)
	return err
}

// ---- CronJob ----

type cronJobRepo struct{ pool *pgxpool.Pool }

func NewCronJobRepo(pool *pgxpool.Pool) domain.CronJobRepository { return &cronJobRepo{pool} }

func (r *cronJobRepo) List(ctx context.Context) ([]domain.CronJob, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id,name,schedule,handler,is_active,last_run_at,next_run_at,created_at FROM cron_jobs ORDER BY name`)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.CronJob, error) {
		var j domain.CronJob
		return j, row.Scan(&j.ID, &j.Name, &j.Schedule, &j.Handler, &j.IsActive, &j.LastRunAt, &j.NextRunAt, &j.CreatedAt)
	})
}

func (r *cronJobRepo) FindByID(ctx context.Context, id string) (*domain.CronJob, error) {
	j := &domain.CronJob{}
	err := r.pool.QueryRow(ctx,
		`SELECT id,name,schedule,handler,is_active,last_run_at,next_run_at,created_at FROM cron_jobs WHERE id=$1`, id).
		Scan(&j.ID, &j.Name, &j.Schedule, &j.Handler, &j.IsActive, &j.LastRunAt, &j.NextRunAt, &j.CreatedAt)
	return j, err
}

func (r *cronJobRepo) Create(ctx context.Context, job *domain.CronJob) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO cron_jobs(name,schedule,handler,is_active) VALUES($1,$2,$3,$4) RETURNING id,created_at`,
		job.Name, job.Schedule, job.Handler, job.IsActive,
	).Scan(&job.ID, &job.CreatedAt)
}

func (r *cronJobRepo) Update(ctx context.Context, job *domain.CronJob) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE cron_jobs SET name=$1,schedule=$2,handler=$3,is_active=$4 WHERE id=$5`,
		job.Name, job.Schedule, job.Handler, job.IsActive, job.ID)
	return err
}

func (r *cronJobRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM cron_jobs WHERE id=$1`, id)
	return err
}

func (r *cronJobRepo) UpdateLastRun(ctx context.Context, id string, lastRun time.Time, nextRun *time.Time) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE cron_jobs SET last_run_at=$1,next_run_at=$2 WHERE id=$3`,
		lastRun, nextRun, id)
	return err
}

// ---- JobExecution ----

type jobExecutionRepo struct{ pool *pgxpool.Pool }

func NewJobExecutionRepo(pool *pgxpool.Pool) domain.JobExecutionRepository {
	return &jobExecutionRepo{pool}
}

func (r *jobExecutionRepo) Create(ctx context.Context, exec *domain.JobExecution) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO job_executions(job_id,status,started_at) VALUES($1,$2,$3) RETURNING id`,
		exec.JobID, exec.Status, exec.StartedAt,
	).Scan(&exec.ID)
}

func (r *jobExecutionRepo) Update(ctx context.Context, exec *domain.JobExecution) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE job_executions SET status=$1,finished_at=$2,duration_ms=$3,error=$4,output=$5 WHERE id=$6`,
		exec.Status, exec.FinishedAt, exec.DurationMs, exec.Error, exec.Output, exec.ID)
	return err
}

func (r *jobExecutionRepo) FindByID(ctx context.Context, id string) (*domain.JobExecution, error) {
	e := &domain.JobExecution{}
	err := r.pool.QueryRow(ctx,
		`SELECT id,job_id,status,started_at,finished_at,duration_ms,error,output FROM job_executions WHERE id=$1`, id).
		Scan(&e.ID, &e.JobID, &e.Status, &e.StartedAt, &e.FinishedAt, &e.DurationMs, &e.Error, &e.Output)
	return e, err
}

func (r *jobExecutionRepo) ListByJob(ctx context.Context, jobID string, offset, limit int) ([]domain.JobExecution, int, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id,job_id,status,started_at,finished_at,duration_ms,error,output FROM job_executions WHERE job_id=$1 ORDER BY started_at DESC LIMIT $2 OFFSET $3`,
		jobID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	execs, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.JobExecution, error) {
		var e domain.JobExecution
		return e, row.Scan(&e.ID, &e.JobID, &e.Status, &e.StartedAt, &e.FinishedAt, &e.DurationMs, &e.Error, &e.Output)
	})
	if err != nil {
		return nil, 0, err
	}
	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM job_executions WHERE job_id=$1`, jobID).Scan(&total); err != nil {
		return nil, 0, err
	}
	return execs, total, nil
}

// ---- AdminTool ----

type adminToolRepo struct{ pool *pgxpool.Pool }

func NewAdminToolRepo(pool *pgxpool.Pool) domain.AdminToolRepository { return &adminToolRepo{pool} }

func (r *adminToolRepo) List(ctx context.Context, activeOnly bool) ([]domain.AdminTool, error) {
	query := `SELECT id,name,description,url,icon,category,is_active,min_role,order_index FROM admin_tools`
	if activeOnly {
		query += ` WHERE is_active=true`
	}
	query += ` ORDER BY order_index`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.AdminTool, error) {
		var t domain.AdminTool
		return t, row.Scan(&t.ID, &t.Name, &t.Description, &t.URL, &t.Icon, &t.Category, &t.IsActive, &t.MinRole, &t.OrderIndex)
	})
}

func (r *adminToolRepo) Create(ctx context.Context, tool *domain.AdminTool) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO admin_tools(name,description,url,icon,category,is_active,min_role,order_index) VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING id`,
		tool.Name, tool.Description, tool.URL, tool.Icon, tool.Category, tool.IsActive, tool.MinRole, tool.OrderIndex,
	).Scan(&tool.ID)
}

func (r *adminToolRepo) Update(ctx context.Context, tool *domain.AdminTool) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE admin_tools SET name=$1,description=$2,url=$3,icon=$4,category=$5,is_active=$6,min_role=$7,order_index=$8 WHERE id=$9`,
		tool.Name, tool.Description, tool.URL, tool.Icon, tool.Category, tool.IsActive, tool.MinRole, tool.OrderIndex, tool.ID)
	return err
}

func (r *adminToolRepo) Delete(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM admin_tools WHERE id=$1`, id)
	return err
}

// ---- AuditLog ----

type auditLogRepo struct{ pool *pgxpool.Pool }

func NewAuditLogRepo(pool *pgxpool.Pool) domain.AuditLogRepository { return &auditLogRepo{pool} }

func (r *auditLogRepo) Create(ctx context.Context, log *domain.AuditLog) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO audit_logs(user_id,action,resource,resource_id,details,ip_address) VALUES($1,$2,$3,$4,$5,$6) RETURNING id,created_at`,
		log.UserID, log.Action, log.Resource, log.ResourceID, log.Details, log.IPAddress,
	).Scan(&log.ID, &log.CreatedAt)
}

func (r *auditLogRepo) List(ctx context.Context, userID, action, resource *string, from, to *time.Time, offset, limit int) ([]domain.AuditLog, int, error) {
	where := []string{}
	args := []any{}
	n := 1

	if userID != nil {
		where = append(where, "user_id=$"+itoa(n))
		args = append(args, *userID)
		n++
	}
	if action != nil {
		where = append(where, "action=$"+itoa(n))
		args = append(args, *action)
		n++
	}
	if resource != nil {
		where = append(where, "resource=$"+itoa(n))
		args = append(args, *resource)
		n++
	}
	if from != nil {
		where = append(where, "created_at>=$"+itoa(n))
		args = append(args, *from)
		n++
	}
	if to != nil {
		where = append(where, "created_at<=$"+itoa(n))
		args = append(args, *to)
		n++
	}

	base := `SELECT id,user_id,action,resource,resource_id,details,ip_address,created_at FROM audit_logs`
	countBase := `SELECT COUNT(*) FROM audit_logs`
	if len(where) > 0 {
		clause := " WHERE " + joinWhere(where)
		base += clause
		countBase += clause
	}

	rows, err := r.pool.Query(ctx, base+" ORDER BY created_at DESC LIMIT $"+itoa(n)+" OFFSET $"+itoa(n+1),
		append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	logs, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.AuditLog, error) {
		var l domain.AuditLog
		return l, row.Scan(&l.ID, &l.UserID, &l.Action, &l.Resource, &l.ResourceID, &l.Details, &l.IPAddress, &l.CreatedAt)
	})
	if err != nil {
		return nil, 0, err
	}
	var total int
	if err := r.pool.QueryRow(ctx, countBase, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

// ---- AppLog ----

type appLogRepo struct{ pool *pgxpool.Pool }

func NewAppLogRepo(pool *pgxpool.Pool) domain.AppLogRepository { return &appLogRepo{pool} }

func (r *appLogRepo) Create(ctx context.Context, log *domain.AppLog) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO app_logs(level,message,source,metadata) VALUES($1,$2,$3,$4) RETURNING id,created_at`,
		log.Level, log.Message, log.Source, log.Metadata,
	).Scan(&log.ID, &log.CreatedAt)
}

func (r *appLogRepo) List(ctx context.Context, level, source, search *string, from, to *time.Time, offset, limit int) ([]domain.AppLog, int, error) {
	where := []string{}
	args := []any{}
	n := 1

	if level != nil {
		where = append(where, "level=$"+itoa(n))
		args = append(args, *level)
		n++
	}
	if source != nil {
		where = append(where, "source=$"+itoa(n))
		args = append(args, *source)
		n++
	}
	if search != nil {
		where = append(where, "message ILIKE $"+itoa(n))
		args = append(args, "%"+*search+"%")
		n++
	}
	if from != nil {
		where = append(where, "created_at>=$"+itoa(n))
		args = append(args, *from)
		n++
	}
	if to != nil {
		where = append(where, "created_at<=$"+itoa(n))
		args = append(args, *to)
		n++
	}

	base := `SELECT id,level,message,source,metadata,created_at FROM app_logs`
	countBase := `SELECT COUNT(*) FROM app_logs`
	if len(where) > 0 {
		clause := " WHERE " + joinWhere(where)
		base += clause
		countBase += clause
	}

	rows, err := r.pool.Query(ctx, base+" ORDER BY created_at DESC LIMIT $"+itoa(n)+" OFFSET $"+itoa(n+1),
		append(args, limit, offset)...)
	if err != nil {
		return nil, 0, err
	}
	logs, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.AppLog, error) {
		var l domain.AppLog
		return l, row.Scan(&l.ID, &l.Level, &l.Message, &l.Source, &l.Metadata, &l.CreatedAt)
	})
	if err != nil {
		return nil, 0, err
	}
	var total int
	if err := r.pool.QueryRow(ctx, countBase, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

func (r *appLogRepo) LatestSince(ctx context.Context, sinceID int64, limit int) ([]domain.AppLog, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id,level,message,source,metadata,created_at FROM app_logs WHERE id > $1 ORDER BY id ASC LIMIT $2`,
		sinceID, limit)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.AppLog, error) {
		var l domain.AppLog
		return l, row.Scan(&l.ID, &l.Level, &l.Message, &l.Source, &l.Metadata, &l.CreatedAt)
	})
}

func (r *appLogRepo) Cleanup(ctx context.Context, olderThanDays int) (int64, error) {
	rows, err := r.pool.Query(ctx,
		`DELETE FROM app_logs WHERE created_at < NOW() - INTERVAL '1 day' * $1 RETURNING id`,
		olderThanDays)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	var count int64
	for rows.Next() {
		count++
	}
	return count, rows.Err()
}

// ---- LogConfig ----

type logConfigRepo struct{ pool *pgxpool.Pool }

func NewLogConfigRepo(pool *pgxpool.Pool) domain.LogConfigRepository { return &logConfigRepo{pool} }

func (r *logConfigRepo) Get(ctx context.Context) (*domain.LogConfig, error) {
	cfg := &domain.LogConfig{}
	err := r.pool.QueryRow(ctx, `SELECT id,retention_days,min_level FROM log_config LIMIT 1`).
		Scan(&cfg.ID, &cfg.RetentionDays, &cfg.MinLevel)
	return cfg, err
}

func (r *logConfigRepo) Update(ctx context.Context, cfg *domain.LogConfig) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE log_config SET retention_days=$1,min_level=$2 WHERE id=$3`,
		cfg.RetentionDays, cfg.MinLevel, cfg.ID)
	return err
}

// joinWhere joins WHERE clause conditions with AND
func joinWhere(parts []string) string {
	result := parts[0]
	for _, p := range parts[1:] {
		result += " AND " + p
	}
	return result
}

// itoa converts int to string for query building without importing strconv everywhere
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := [20]byte{}
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[pos:])
}
