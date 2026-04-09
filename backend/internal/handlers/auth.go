package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/afa/blueprint/backend/internal/domain"
	"github.com/afa/blueprint/backend/pkg/config"
	"github.com/afa/blueprint/backend/pkg/middleware"
)

type AuthHandler struct {
	users domain.UserRepository
	cfg   *config.Config
}

func NewAuthHandler(users domain.UserRepository, cfg *config.Config) *AuthHandler {
	return &AuthHandler{users: users, cfg: cfg}
}

type userResponse struct {
	ID        string     `json:"id"`
	Email     string     `json:"email"`
	Name      *string    `json:"name"`
	Role      string     `json:"role"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func toUserResponse(u *domain.User) userResponse {
	return userResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func (h *AuthHandler) generateTokens(userID, role string) (accessToken, refreshToken string, err error) {
	accessClaims := middleware.Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.cfg.JWTExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	access := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = access.SignedString([]byte(h.cfg.JWTSecret))
	if err != nil {
		return
	}

	refreshClaims := middleware.Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(h.cfg.RefreshExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	refresh := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refresh.SignedString([]byte(h.cfg.JWTSecret))
	return
}

func (h *AuthHandler) setTokenCookies(c *fiber.Ctx, accessToken, refreshToken string) {
	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		HTTPOnly: true,
		Secure:   h.cfg.Env != "development",
		SameSite: "Lax",
		MaxAge:   int(h.cfg.JWTExpiry.Seconds()),
		Path:     "/",
	})
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		HTTPOnly: true,
		Secure:   h.cfg.Env != "development",
		SameSite: "Lax",
		MaxAge:   int(h.cfg.RefreshExpiry.Seconds()),
		Path:     "/api/v1/auth",
	})
}

func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req struct {
		Email    string  `json:"email"`
		Password string  `json:"password"`
		Name     *string `json:"name"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	if req.Email == "" || req.Password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "email and password are required"})
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "internal error"})
	}

	u := &domain.User{
		ID:           uuid.NewString(),
		Email:        req.Email,
		PasswordHash: string(hash),
		Name:         req.Name,
		Role:         "user",
	}
	if err := h.users.Create(c.Context(), u); err != nil {
		return c.Status(409).JSON(fiber.Map{"error": "email already in use"})
	}

	accessToken, refreshToken, err := h.generateTokens(u.ID, u.Role)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "could not generate tokens"})
	}
	h.setTokenCookies(c, accessToken, refreshToken)

	return c.Status(201).JSON(fiber.Map{
		"user":          toUserResponse(u),
		"access_token":  accessToken,
	})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}

	u, err := h.users.FindByEmail(c.Context(), req.Email)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid credentials"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		return c.Status(401).JSON(fiber.Map{"error": "invalid credentials"})
	}

	accessToken, refreshToken, err := h.generateTokens(u.ID, u.Role)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "could not generate tokens"})
	}
	h.setTokenCookies(c, accessToken, refreshToken)

	return c.JSON(fiber.Map{
		"user":         toUserResponse(u),
		"access_token": accessToken,
	})
}

func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	tokenStr := c.Cookies("refresh_token")
	if tokenStr == "" {
		return c.Status(401).JSON(fiber.Map{"error": "missing refresh token"})
	}

	claims := &middleware.Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(h.cfg.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return c.Status(401).JSON(fiber.Map{"error": "invalid refresh token"})
	}

	accessToken, refreshToken, err := h.generateTokens(claims.UserID, claims.Role)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "could not generate tokens"})
	}
	h.setTokenCookies(c, accessToken, refreshToken)

	return c.JSON(fiber.Map{"access_token": accessToken})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	c.Cookie(&fiber.Cookie{
		Name:    "access_token",
		Value:   "",
		MaxAge:  -1,
		Path:    "/",
	})
	c.Cookie(&fiber.Cookie{
		Name:    "refresh_token",
		Value:   "",
		MaxAge:  -1,
		Path:    "/api/v1/auth",
	})
	return c.JSON(fiber.Map{"message": "logged out"})
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID, _ := c.Locals("user_id").(string)
	u, err := h.users.FindByID(c.Context(), userID)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "user not found"})
	}
	return c.JSON(toUserResponse(u))
}

func (h *AuthHandler) ForgotPassword(c *fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	// Placeholder: in production, send reset email
	return c.JSON(fiber.Map{"message": "if that email exists, a reset link has been sent"})
}

func (h *AuthHandler) ResetPassword(c *fiber.Ctx) error {
	var req struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	// Placeholder: in production, validate token and update password
	return c.JSON(fiber.Map{"message": "password reset successful"})
}
