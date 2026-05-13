package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/moistello/backend/internal/api/middleware"
	"github.com/moistello/backend/internal/domain/contribution"
	"github.com/moistello/backend/pkg/pagination"
	"github.com/moistello/backend/pkg/response"
)

type ContributionHandler struct {
	contribService contribution.Service
	contribRepo    contribution.Repository
}

func NewContributionHandler(svc contribution.Service, repo contribution.Repository) *ContributionHandler {
	return &ContributionHandler{contribService: svc, contribRepo: repo}
}

func (h *ContributionHandler) ListContributions(c *gin.Context) {
	userIDFilter := c.Query("userId")
	circleIDFilter := c.Query("circleId")
	page, limit, _ := pagination.Parse(c)

	if circleIDFilter != "" {
		contribs, total, err := h.contribService.GetCircleHistory(c.Request.Context(), circleIDFilter, page, limit)
		if err != nil {
			response.InternalError(c, "failed to list contributions")
			return
		}
		response.OKWithMeta(c, gin.H{"contributions": contribs}, response.NewPaginationMeta(page, limit, total))
		return
	}

	if userIDFilter != "" {
		contribs, total, err := h.contribService.GetUserHistory(c.Request.Context(), userIDFilter, page, limit)
		if err != nil {
			response.InternalError(c, "failed to list contributions")
			return
		}
		response.OKWithMeta(c, gin.H{"contributions": contribs}, response.NewPaginationMeta(page, limit, total))
		return
	}

	userID := middleware.GetUserID(c)
	contribs, total, err := h.contribService.GetUserHistory(c.Request.Context(), userID, page, limit)
	if err != nil {
		response.InternalError(c, "failed to list contributions")
		return
	}
	response.OKWithMeta(c, gin.H{"contributions": contribs}, response.NewPaginationMeta(page, limit, total))
}

func (h *ContributionHandler) GetContribution(c *gin.Context) {
	id := c.Param("id")
	uid, err := uuid.Parse(id)
	if err != nil {
		response.BadRequest(c, "invalid contribution ID")
		return
	}
	contrib, err := h.contribRepo.FindByID(c.Request.Context(), uid)
	if err != nil {
		response.NotFound(c, "contribution not found")
		return
	}
	response.OK(c, gin.H{"contribution": contrib})
}
