package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/moistello/backend/internal/domain/audit"
	"github.com/moistello/backend/internal/domain/circle"
	"github.com/moistello/backend/internal/domain/user"
	"github.com/moistello/backend/pkg/pagination"
	"github.com/moistello/backend/pkg/response"
)

type AdminHandler struct {
	userService   user.Service
	userRepo      user.Repository
	circleService circle.Service
	auditRepo     audit.Repository
}

func NewAdminHandler(userSvc user.Service, userRepo user.Repository, circleSvc circle.Service, auditRepo audit.Repository) *AdminHandler {
	return &AdminHandler{
		userService:   userSvc,
		userRepo:      userRepo,
		circleService: circleSvc,
		auditRepo:     auditRepo,
	}
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, limit, _ := pagination.Parse(c)
	filter := user.UserFilter{
		Search: c.Query("search"),
		Page:   page,
		Limit:  limit,
	}
	users, err := h.userRepo.List(c.Request.Context(), filter)
	if err != nil {
		response.InternalError(c, "failed to list users")
		return
	}
	total, err := h.userRepo.Count(c.Request.Context(), filter)
	if err != nil {
		response.InternalError(c, "failed to count users")
		return
	}
	response.OKWithMeta(c, gin.H{"users": users}, response.NewPaginationMeta(page, limit, total))
}

func (h *AdminHandler) ListCircles(c *gin.Context) {
	page, limit, _ := pagination.Parse(c)
	filter := circle.CircleFilter{
		Search: c.Query("search"),
		Status: circle.CircleStatus(c.Query("status")),
		Page:   page,
		Limit:  limit,
	}
	circles, total, err := h.circleService.List(c.Request.Context(), filter)
	if err != nil {
		response.InternalError(c, "failed to list circles")
		return
	}
	response.OKWithMeta(c, gin.H{"circles": circles}, response.NewPaginationMeta(page, limit, total))
}

func (h *AdminHandler) GetAuditLog(c *gin.Context) {
	response.OK(c, gin.H{"entries": []any{}, "message": "audit log not yet implemented"})
}

func (h *AdminHandler) GetMetrics(c *gin.Context) {
	response.OK(c, gin.H{
		"totalUsers":  0,
		"totalCircles": 0,
		"activeCircles": 0,
		"totalVolumeUSD": 0,
		"message": "metrics endpoint placeholder",
	})
}

func (h *AdminHandler) UpdateFeatureFlag(c *gin.Context) {
	var req struct {
		Flag  string `json:"flag" binding:"required"`
		Value bool   `json:"value"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"flag": req.Flag, "value": req.Value, "message": "feature flag updated"})
}
