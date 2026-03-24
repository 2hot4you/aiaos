package handler

import (
	"strconv"

	"github.com/2hot4you/aiaos/backend/internal/domain"
	"github.com/2hot4you/aiaos/backend/internal/middleware"
	"github.com/2hot4you/aiaos/backend/internal/service"
	"github.com/2hot4you/aiaos/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	svc *service.ProjectService
}

func NewProjectHandler(svc *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{svc: svc}
}

type CreateProjectRequest struct {
	Name string `json:"name" binding:"required"`
}

func (h *ProjectHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")
	sortBy := c.Query("sort_by")
	sortOrder := c.Query("sort_order")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	items, total, err := h.svc.List(page, pageSize, keyword, sortBy, sortOrder)
	if err != nil {
		response.InternalError(c, "failed to list projects")
		return
	}

	response.SuccessPage(c, items, total, page, pageSize)
}

func (h *ProjectHandler) Create(c *gin.Context) {
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "name is required")
		return
	}

	userID := middleware.GetUserID(c)
	p, err := h.svc.Create(req.Name, userID)
	if err != nil {
		response.InternalError(c, "failed to create project")
		return
	}

	response.Success(c, p)
}

func (h *ProjectHandler) Get(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}

	p, err := h.svc.GetByID(id)
	if err != nil {
		if err == domain.ErrProjectNotFound {
			response.NotFound(c, 40401, "project not found")
			return
		}
		response.InternalError(c, "failed to get project")
		return
	}

	response.Success(c, p)
}

type UpdateProjectRequest struct {
	Name string `json:"name"`
}

func (h *ProjectHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}

	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	p, err := h.svc.Update(id, req.Name)
	if err != nil {
		if err == domain.ErrProjectNotFound {
			response.NotFound(c, 40401, "project not found")
			return
		}
		response.InternalError(c, "failed to update project")
		return
	}

	response.Success(c, p)
}

func (h *ProjectHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid project id")
		return
	}

	if err := h.svc.Delete(id); err != nil {
		if err == domain.ErrProjectNotFound {
			response.NotFound(c, 40401, "project not found")
			return
		}
		response.InternalError(c, "failed to delete project")
		return
	}

	response.Success(c, nil)
}
