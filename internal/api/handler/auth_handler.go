package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/moistello/backend/internal/api/middleware"
	"github.com/moistello/backend/internal/domain/auth"
	"github.com/moistello/backend/internal/domain/user"
	"github.com/moistello/backend/pkg/response"
)

type AuthHandler struct {
	authService auth.Service
	userService user.Service
}

func NewAuthHandler(authSvc auth.Service, userSvc user.Service) *AuthHandler {
	return &AuthHandler{authService: authSvc, userService: userSvc}
}

func (h *AuthHandler) Nonce(c *gin.Context) {
	var req struct {
		WalletAddress string `json:"walletAddress" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	nonce, err := h.authService.GenerateNonce(c.Request.Context(), req.WalletAddress)
	if err != nil {
		response.InternalError(c, "failed to generate nonce")
		return
	}
	response.OK(c, gin.H{"nonce": nonce})
}

func (h *AuthHandler) Verify(c *gin.Context) {
	var req struct {
		WalletAddress string `json:"walletAddress" binding:"required"`
		Signature     string `json:"signature" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	valid, err := h.authService.VerifySignature(c.Request.Context(), req.WalletAddress, req.Signature)
	if err != nil || !valid {
		response.Unauthorized(c, "signature verification failed")
		return
	}
	u, err := h.userService.Create(c.Request.Context(), req.WalletAddress)
	if err != nil {
		response.InternalError(c, "failed to create user")
		return
	}
	tokenPair, err := h.authService.CreateSession(c.Request.Context(), u.ID)
	if err != nil {
		response.InternalError(c, "failed to create session")
		return
	}
	response.OK(c, gin.H{"token": tokenPair.AccessToken, "refreshToken": tokenPair.RefreshToken, "user": u})
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		WalletAddress string  `json:"walletAddress" binding:"required"`
		Signature     string  `json:"signature" binding:"required"`
		DisplayName   *string `json:"displayName"`
		Email         *string `json:"email"`
		CountryCode   *string `json:"countryCode"`
		Language      *string `json:"language"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	valid, err := h.authService.VerifySignature(c.Request.Context(), req.WalletAddress, req.Signature)
	if err != nil || !valid {
		response.Unauthorized(c, "signature verification failed")
		return
	}
	u, err := h.userService.Create(c.Request.Context(), req.WalletAddress)
	if err != nil {
		response.InternalError(c, "failed to create user")
		return
	}
	updates := user.UpdateProfileInput{
		DisplayName:       req.DisplayName,
		Email:             req.Email,
		CountryCode:       req.CountryCode,
		PreferredLanguage: req.Language,
	}
	u, err = h.userService.UpdateProfile(c.Request.Context(), u.ID.String(), updates)
	if err != nil {
		response.InternalError(c, "failed to update profile")
		return
	}
	tokenPair, err := h.authService.CreateSession(c.Request.Context(), u.ID)
	if err != nil {
		response.InternalError(c, "failed to create session")
		return
	}
	response.OK(c, gin.H{"token": tokenPair.AccessToken, "refreshToken": tokenPair.RefreshToken, "user": u})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refreshToken" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	tokenPair, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		response.Unauthorized(c, "invalid refresh token")
		return
	}
	response.OK(c, gin.H{"token": tokenPair.AccessToken, "refreshToken": tokenPair.RefreshToken})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID := middleware.GetUserID(c)
	u, err := h.userService.GetByID(c.Request.Context(), userID)
	if err != nil {
		response.Unauthorized(c, "user not found")
		return
	}
	response.OK(c, gin.H{"user": u})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	response.OK(c, gin.H{"success": true})
}
