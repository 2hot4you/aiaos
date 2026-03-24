package handler

import (
	"strconv"

	"github.com/2hot4you/aiaos/backend/internal/domain"
	"github.com/2hot4you/aiaos/backend/internal/middleware"
	"github.com/2hot4you/aiaos/backend/internal/service"
	"github.com/2hot4you/aiaos/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type EpisodeHandler struct {
	svc *service.EpisodeService
}

func NewEpisodeHandler(svc *service.EpisodeService) *EpisodeHandler {
	return &EpisodeHandler{svc: svc}
}

type CreateEpisodeRequest struct {
	Title string `json:"title" binding:"required"`
}

func (h *EpisodeHandler) ListBySeason(c *gin.Context) {
	sid, err := strconv.ParseInt(c.Param("sid"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid season id")
		return
	}

	items, err := h.svc.ListBySeason(sid)
	if err != nil {
		if err == domain.ErrSeasonNotFound {
			response.NotFound(c, 40402, "season not found")
			return
		}
		response.InternalError(c, "failed to list episodes")
		return
	}

	response.Success(c, items)
}

func (h *EpisodeHandler) Create(c *gin.Context) {
	sid, err := strconv.ParseInt(c.Param("sid"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid season id")
		return
	}

	var req CreateEpisodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "title is required")
		return
	}

	userID := middleware.GetUserID(c)
	ep, err := h.svc.Create(sid, req.Title, userID)
	if err != nil {
		if err == domain.ErrSeasonNotFound {
			response.NotFound(c, 40402, "season not found")
			return
		}
		response.InternalError(c, "failed to create episode")
		return
	}

	response.Success(c, ep)
}

func (h *EpisodeHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid episode id")
		return
	}

	ep, err := h.svc.GetByID(id)
	if err != nil {
		if err == domain.ErrEpisodeNotFound {
			response.NotFound(c, 40403, "episode not found")
			return
		}
		response.InternalError(c, "failed to get episode")
		return
	}

	response.Success(c, ep)
}

func (h *EpisodeHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid episode id")
		return
	}

	var req CreateEpisodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "title is required")
		return
	}

	ep, err := h.svc.Update(id, req.Title)
	if err != nil {
		if err == domain.ErrEpisodeNotFound {
			response.NotFound(c, 40403, "episode not found")
			return
		}
		response.InternalError(c, "failed to update episode")
		return
	}

	response.Success(c, ep)
}

func (h *EpisodeHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid episode id")
		return
	}

	if err := h.svc.Delete(id); err != nil {
		if err == domain.ErrEpisodeNotFound {
			response.NotFound(c, 40403, "episode not found")
			return
		}
		response.InternalError(c, "failed to delete episode")
		return
	}

	response.Success(c, nil)
}
