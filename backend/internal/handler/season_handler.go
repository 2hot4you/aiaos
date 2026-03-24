package handler

import (
	"strconv"

	"github.com/2hot4you/aiaos/backend/internal/domain"
	"github.com/2hot4you/aiaos/backend/internal/middleware"
	"github.com/2hot4you/aiaos/backend/internal/service"
	"github.com/2hot4you/aiaos/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type SeasonHandler struct {
	svc *service.SeasonService
}

func NewSeasonHandler(svc *service.SeasonService) *SeasonHandler {
	return &SeasonHandler{svc: svc}
}

type CreateSeasonRequest struct {
	Title string `json:"title" binding:"required"`
}

func (h *SeasonHandler) ListByProject(c *gin.Context) {
	pid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}

	items, err := h.svc.ListByProject(pid)
	if err != nil {
		if err == domain.ErrProjectNotFound {
			response.NotFound(c, 40401, "project not found")
			return
		}
		response.InternalError(c, "failed to list seasons")
		return
	}

	response.Success(c, items)
}

func (h *SeasonHandler) Create(c *gin.Context) {
	pid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}

	var req CreateSeasonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "title is required")
		return
	}

	userID := middleware.GetUserID(c)
	s, err := h.svc.Create(pid, req.Title, userID)
	if err != nil {
		if err == domain.ErrProjectNotFound {
			response.NotFound(c, 40401, "project not found")
			return
		}
		response.InternalError(c, "failed to create season")
		return
	}

	response.Success(c, s)
}

func (h *SeasonHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid season id")
		return
	}

	var req CreateSeasonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "title is required")
		return
	}

	s, err := h.svc.Update(id, req.Title)
	if err != nil {
		if err == domain.ErrSeasonNotFound {
			response.NotFound(c, 40402, "season not found")
			return
		}
		response.InternalError(c, "failed to update season")
		return
	}

	response.Success(c, s)
}

func (h *SeasonHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid season id")
		return
	}

	if err := h.svc.Delete(id); err != nil {
		if err == domain.ErrSeasonNotFound {
			response.NotFound(c, 40402, "season not found")
			return
		}
		response.InternalError(c, "failed to delete season")
		return
	}

	response.Success(c, nil)
}
