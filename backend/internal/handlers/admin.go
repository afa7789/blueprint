package handlers

import (
	"strconv"

	"github.com/afa/blueprint/backend/internal/domain"
	"github.com/afa/blueprint/backend/pkg/config"
	"github.com/gofiber/fiber/v2"
)

type AdminHandler struct {
	users       domain.UserRepository
	banners     domain.BannerRepository
	linktree    domain.LinktreeRepository
	brandKit    domain.BrandKitRepository
	emailGroups domain.EmailGroupRepository
	emailSubs   domain.EmailSubscriptionRepository
	userGroups  domain.UserGroupRepository
	cfg         *config.Config
}

func NewAdminHandler(
	users domain.UserRepository,
	banners domain.BannerRepository,
	linktree domain.LinktreeRepository,
	brandKit domain.BrandKitRepository,
	emailGroups domain.EmailGroupRepository,
	emailSubs domain.EmailSubscriptionRepository,
	userGroups domain.UserGroupRepository,
	cfg *config.Config,
) *AdminHandler {
	return &AdminHandler{
		users:       users,
		banners:     banners,
		linktree:    linktree,
		brandKit:    brandKit,
		emailGroups: emailGroups,
		emailSubs:   emailSubs,
		userGroups:  userGroups,
		cfg:         cfg,
	}
}

// pagination helpers
func paginate(c *fiber.Ctx) (page, limit, offset int) {
	page, _ = strconv.Atoi(c.Query("page", "1"))
	limit, _ = strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	offset = (page - 1) * limit
	return
}

// ---- Users ----

func (h *AdminHandler) ListUsers(c *fiber.Ctx) error {
	page, limit, offset := paginate(c)
	users, total, err := h.users.List(c.Context(), offset, limit)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": users, "total": total, "page": page, "limit": limit})
}

func (h *AdminHandler) UpdateUserRole(c *fiber.Ctx) error {
	id := c.Params("id")
	var body struct {
		Role string `json:"role"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	switch body.Role {
	case "admin", "operator", "user":
	default:
		return fiber.NewError(fiber.StatusBadRequest, "invalid role")
	}
	user, err := h.users.FindByID(c.Context(), id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "user not found")
	}
	user.Role = body.Role
	if err := h.users.Update(c.Context(), user); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(user)
}

func (h *AdminHandler) DeleteUser(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := h.users.Delete(c.Context(), id); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ---- Banners ----

func (h *AdminHandler) ListBanners(c *fiber.Ctx) error {
	banners, err := h.banners.List(c.Context(), false)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": banners})
}

func (h *AdminHandler) CreateBanner(c *fiber.Ctx) error {
	var b domain.Banner
	if err := c.BodyParser(&b); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if err := h.banners.Create(c.Context(), &b); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(b)
}

func (h *AdminHandler) UpdateBanner(c *fiber.Ctx) error {
	var b domain.Banner
	if err := c.BodyParser(&b); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	b.ID = c.Params("id")
	if err := h.banners.Update(c.Context(), &b); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(b)
}

func (h *AdminHandler) DeleteBanner(c *fiber.Ctx) error {
	if err := h.banners.Delete(c.Context(), c.Params("id")); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ---- Linktree ----

func (h *AdminHandler) ListLinktree(c *fiber.Ctx) error {
	items, err := h.linktree.List(c.Context(), false)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": items})
}

func (h *AdminHandler) CreateLinktreeItem(c *fiber.Ctx) error {
	var item domain.LinktreeItem
	if err := c.BodyParser(&item); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if err := h.linktree.Create(c.Context(), &item); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(item)
}

func (h *AdminHandler) UpdateLinktreeItem(c *fiber.Ctx) error {
	var item domain.LinktreeItem
	if err := c.BodyParser(&item); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	item.ID = c.Params("id")
	if err := h.linktree.Update(c.Context(), &item); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(item)
}

func (h *AdminHandler) DeleteLinktreeItem(c *fiber.Ctx) error {
	if err := h.linktree.Delete(c.Context(), c.Params("id")); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (h *AdminHandler) ReorderLinktree(c *fiber.Ctx) error {
	var items []domain.LinktreeItem
	if err := c.BodyParser(&items); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if err := h.linktree.Reorder(c.Context(), items); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ---- Brand Kit ----

func (h *AdminHandler) GetBrandKit(c *fiber.Ctx) error {
	bk, err := h.brandKit.Get(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(bk)
}

func (h *AdminHandler) UpsertBrandKit(c *fiber.Ctx) error {
	var bk domain.BrandKit
	if err := c.BodyParser(&bk); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if err := h.brandKit.Upsert(c.Context(), &bk); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(bk)
}

// ---- Email Groups ----

func (h *AdminHandler) ListEmailGroups(c *fiber.Ctx) error {
	groups, err := h.emailGroups.List(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": groups})
}

func (h *AdminHandler) CreateEmailGroup(c *fiber.Ctx) error {
	var g domain.EmailGroup
	if err := c.BodyParser(&g); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if err := h.emailGroups.Create(c.Context(), &g); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(g)
}

func (h *AdminHandler) DeleteEmailGroup(c *fiber.Ctx) error {
	if err := h.emailGroups.Delete(c.Context(), c.Params("id")); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ---- Email Subscriptions ----

func (h *AdminHandler) ListSubscribers(c *fiber.Ctx) error {
	subs, err := h.emailSubs.ListByGroup(c.Context(), c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": subs})
}

func (h *AdminHandler) AddEmailSubscription(c *fiber.Ctx) error {
	var sub domain.EmailSubscription
	if err := c.BodyParser(&sub); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if err := h.emailSubs.Add(c.Context(), &sub); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(sub)
}

func (h *AdminHandler) DeactivateEmailSubscription(c *fiber.Ctx) error {
	email := c.Params("email")
	if err := h.emailSubs.Deactivate(c.Context(), email); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ---- User Groups ----

func (h *AdminHandler) ListUserGroups(c *fiber.Ctx) error {
	groups, err := h.userGroups.List(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"data": groups})
}

func (h *AdminHandler) CreateUserGroup(c *fiber.Ctx) error {
	var g domain.UserGroup
	if err := c.BodyParser(&g); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if err := h.userGroups.Create(c.Context(), &g); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(g)
}

func (h *AdminHandler) DeleteUserGroup(c *fiber.Ctx) error {
	if err := h.userGroups.Delete(c.Context(), c.Params("id")); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.SendStatus(fiber.StatusNoContent)
}
