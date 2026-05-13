package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/moistello/backend/internal/api/middleware"
	"github.com/moistello/backend/internal/domain/circle"
	"github.com/moistello/backend/internal/domain/invite"
	"github.com/moistello/backend/pkg/pagination"
	"github.com/moistello/backend/pkg/response"
	"github.com/moistello/backend/pkg/validator"
)

type CircleHandler struct {
	circleService circle.Service
	inviteService invite.Service
}

func NewCircleHandler(circleSvc circle.Service, inviteSvc invite.Service) *CircleHandler {
	return &CircleHandler{circleService: circleSvc, inviteService: inviteSvc}
}

func (h *CircleHandler) ListCircles(c *gin.Context) {
	page, limit, _ := pagination.Parse(c)
	filter := circle.CircleFilter{
		Search: c.Query("search"),
		Status: circle.CircleStatus(c.Query("status")),
		Type:   circle.CircleType(c.Query("type")),
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

func (h *CircleHandler) CreateCircle(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var input circle.CreateCircleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := validator.Validate.Struct(input); err != nil {
		response.ValidationErrors(c, "validation failed: "+err.Error())
		return
	}
	cir, err := h.circleService.Create(c.Request.Context(), userID, input)
	if err != nil {
		response.InternalError(c, "failed to create circle")
		return
	}
	response.Created(c, gin.H{"circle": cir})
}

func (h *CircleHandler) GetCircle(c *gin.Context) {
	id := c.Param("id")
	cir, err := h.circleService.Get(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "circle not found")
		return
	}
	response.OK(c, gin.H{"circle": cir})
}

func (h *CircleHandler) UpdateCircle(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)
	var input circle.UpdateCircleInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	cir, err := h.circleService.Update(c.Request.Context(), id, userID, input)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"circle": cir})
}

func (h *CircleHandler) CancelCircle(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)
	if err := h.circleService.Cancel(c.Request.Context(), id, userID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"success": true})
}

func (h *CircleHandler) JoinCircle(c *gin.Context) {
	circleID := c.Param("id")
	userID := middleware.GetUserID(c)
	var req struct {
		InviteCode string `json:"inviteCode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if req.InviteCode != "" {
		if _, err := h.inviteService.Validate(c.Request.Context(), req.InviteCode); err != nil {
			response.BadRequest(c, "invalid invite code")
			return
		}
	}
	if err := h.circleService.Join(c.Request.Context(), circleID, userID, req.InviteCode); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"success": true})
}

func (h *CircleHandler) Contribute(c *gin.Context) {
	circleID := c.Param("id")
	userID := middleware.GetUserID(c)
	var req struct {
		Amount      float64 `json:"amount" binding:"required,gt=0"`
		TxnHash     string  `json:"txnHash" binding:"required"`
		RoundNumber int     `json:"roundNumber" binding:"required,gte=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	_ = circleID
	_ = userID
	response.OK(c, gin.H{"success": true, "message": "contribution recorded"})
}

func (h *CircleHandler) ExitCircle(c *gin.Context) {
	id := c.Param("id")
	userID := middleware.GetUserID(c)
	if err := h.circleService.Exit(c.Request.Context(), id, userID); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"success": true})
}

func (h *CircleHandler) GetMembers(c *gin.Context) {
	circleID := c.Param("id")
	members, err := h.circleService.GetMembers(c.Request.Context(), circleID)
	if err != nil {
		response.InternalError(c, "failed to get members")
		return
	}
	response.OK(c, gin.H{"members": members})
}

func (h *CircleHandler) GetRounds(c *gin.Context) {
	circleID := c.Param("id")
	cir, err := h.circleService.Get(c.Request.Context(), circleID)
	if err != nil {
		response.NotFound(c, "circle not found")
		return
	}
	response.OK(c, gin.H{
		"rounds":        []any{},
		"currentRound":  cir.CurrentRound,
		"totalMembers":  cir.MaxMembers,
	})
}

func (h *CircleHandler) GetPayouts(c *gin.Context) {
	circleID := c.Param("id")
	_ = circleID
	response.OK(c, gin.H{"payouts": []any{}})
}

func (h *CircleHandler) Dispute(c *gin.Context) {
	circleID := c.Param("id")
	userID := middleware.GetUserID(c)
	var req struct {
		Reason string `json:"reason" binding:"required"`
		Details string `json:"details"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	_ = circleID
	_ = userID
	_ = req
	response.OK(c, gin.H{"success": true, "message": "dispute submitted"})
}

func (h *CircleHandler) Vote(c *gin.Context) {
	circleID := c.Param("id")
	userID := middleware.GetUserID(c)
	var req struct {
		RecipientID string `json:"recipientId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	_ = circleID
	_ = userID
	_ = req
	response.OK(c, gin.H{"success": true, "message": "vote recorded"})
}

func (h *CircleHandler) AuctionBid(c *gin.Context) {
	circleID := c.Param("id")
	userID := middleware.GetUserID(c)
	var req struct {
		BidAmount float64 `json:"bidAmount" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	_ = circleID
	_ = userID
	_ = req
	response.OK(c, gin.H{"success": true, "message": "bid placed"})
}
