package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/moistello/backend/internal/api/middleware"
	"github.com/moistello/backend/internal/domain/notification"
	"github.com/moistello/backend/pkg/pagination"
	"github.com/moistello/backend/pkg/response"
)

type NotificationHandler struct {
	notificationService notification.Service
}

func NewNotificationHandler(svc notification.Service) *NotificationHandler {
	return &NotificationHandler{notificationService: svc}
}

func (h *NotificationHandler) ListNotifications(c *gin.Context) {
	userID := middleware.GetUserID(c)
	unreadOnly := c.Query("unread") == "true"
	page, limit, _ := pagination.Parse(c)
	notifications, total, err := h.notificationService.List(c.Request.Context(), userID, page, limit, unreadOnly)
	if err != nil {
		response.InternalError(c, "failed to list notifications")
		return
	}
	response.OKWithMeta(c, gin.H{"notifications": notifications}, response.NewPaginationMeta(page, limit, total))
}

func (h *NotificationHandler) MarkRead(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)
	if err := h.notificationService.MarkRead(c.Request.Context(), id, userID); err != nil {
		response.InternalError(c, "failed to mark notification as read")
		return
	}
	response.OK(c, gin.H{"success": true})
}

func (h *NotificationHandler) MarkAllRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if err := h.notificationService.MarkAllRead(c.Request.Context(), userID); err != nil {
		response.InternalError(c, "failed to mark all notifications as read")
		return
	}
	response.OK(c, gin.H{"success": true})
}

func (h *NotificationHandler) UpdatePreferences(c *gin.Context) {
	var req struct {
		Channels []string `json:"channels"`
		Muted    bool     `json:"muted"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"preferences": gin.H{"channels": req.Channels, "muted": req.Muted}})
}
