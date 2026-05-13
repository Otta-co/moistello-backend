package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/moistello/backend/internal/api/middleware"
	"github.com/moistello/backend/internal/domain/invite"
	"github.com/moistello/backend/pkg/response"
)

type InviteHandler struct {
	inviteService invite.Service
}

func NewInviteHandler(svc invite.Service) *InviteHandler {
	return &InviteHandler{inviteService: svc}
}

func (h *InviteHandler) CreateInvite(c *gin.Context) {
	circleID := c.Param("id")
	userID := middleware.GetUserID(c)
	var input invite.GenerateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	input.CircleID = circleID
	input.UserID = userID
	inv, err := h.inviteService.Generate(c.Request.Context(), input)
	if err != nil {
		response.InternalError(c, "failed to create invite")
		return
	}
	response.Created(c, gin.H{"invite": inv})
}

func (h *InviteHandler) ListInvites(c *gin.Context) {
	circleID := c.Param("id")
	invites, err := h.inviteService.List(c.Request.Context(), circleID)
	if err != nil {
		response.InternalError(c, "failed to list invites")
		return
	}
	response.OK(c, gin.H{"invites": invites})
}

func (h *InviteHandler) RevokeInvite(c *gin.Context) {
	code := c.Param("code")
	userID := middleware.GetUserID(c)
	if err := h.inviteService.Revoke(c.Request.Context(), code, userID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"success": true})
}
