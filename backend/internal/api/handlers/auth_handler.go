package handlers

import (
	"net/http"

	"github.com/drayauction/auctionxi/internal/api/dto"
	"github.com/drayauction/auctionxi/internal/domain"
	"github.com/drayauction/auctionxi/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{Code: "VALIDATION_ERROR", Message: err.Error()},
		})
		return
	}

	user, tokens, err := h.authService.Register(c.Request.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.AuthResponse{
		User:         toUserResponse(user),
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{Code: "VALIDATION_ERROR", Message: err.Error()},
		})
		return
	}

	user, tokens, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.AuthResponse{
		User:         toUserResponse(user),
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error: dto.ErrorDetail{Code: "VALIDATION_ERROR", Message: err.Error()},
		})
		return
	}

	tokens, err := h.authService.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID, _ := c.Get("user_id")
	uid := userID.(uuid.UUID)
	_ = h.authService.Logout(c.Request.Context(), uid)
	c.Status(http.StatusNoContent)
}

func toUserResponse(u *domain.User) dto.UserResponse {
	return dto.UserResponse{
		ID:         u.ID.String(),
		Username:   u.Username,
		Email:      u.Email,
		Coins:      u.Coins,
		XP:         u.XP,
		Rank:       string(u.RankTier),
		RankPoints: u.RankPoints,
		Wins:       u.Wins,
		Losses:     u.Losses,
	}
}

func handleError(c *gin.Context, err error) {
	if appErr, ok := err.(*domain.AppError); ok {
		c.JSON(appErr.HTTPStatus, dto.ErrorResponse{
			Error: dto.ErrorDetail{Code: appErr.Code, Message: appErr.Message},
		})
		return
	}
	c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
		Error: dto.ErrorDetail{Code: "INTERNAL_ERROR", Message: "An unexpected error occurred"},
	})
}
