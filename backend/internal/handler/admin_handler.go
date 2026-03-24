package handler

import (
	"strconv"

	"github.com/2hot4you/aiaos/backend/internal/domain"
	"github.com/2hot4you/aiaos/backend/internal/service"
	"github.com/2hot4you/aiaos/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	svc *service.AdminService
}

func NewAdminHandler(svc *service.AdminService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

// ---- User Management ----

func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	keyword := c.Query("keyword")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	users, total, err := h.svc.ListUsers(page, pageSize, keyword)
	if err != nil {
		response.InternalError(c, "failed to list users")
		return
	}

	response.SuccessPage(c, users, total, page, pageSize)
}

type CreateUserRequest struct {
	Username    string `json:"username" binding:"required"`
	DisplayName string `json:"display_name" binding:"required"`
	Password    string `json:"password" binding:"required,min=8"`
	Role        string `json:"role" binding:"required,oneof=admin user"`
}

func (h *AdminHandler) CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request: username, display_name, password(min 8), role(admin|user) required")
		return
	}

	user, err := h.svc.CreateUser(req.Username, req.DisplayName, req.Password, req.Role)
	if err != nil {
		if err == domain.ErrUsernameExists {
			response.Conflict(c, 40901, "用户名已存在")
			return
		}
		response.InternalError(c, "failed to create user")
		return
	}

	response.Success(c, user)
}

type UpdateUserRequest struct {
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`
}

func (h *AdminHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	user, err := h.svc.UpdateUser(id, req.DisplayName, req.Role)
	if err != nil {
		if err == domain.ErrUserNotFound {
			response.NotFound(c, 40407, "user not found")
			return
		}
		response.InternalError(c, "failed to update user")
		return
	}

	response.Success(c, user)
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	if err := h.svc.DeleteUser(id); err != nil {
		response.InternalError(c, "failed to delete user")
		return
	}

	response.Success(c, nil)
}

type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

func (h *AdminHandler) ResetPassword(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "new_password(min 8 chars) is required")
		return
	}

	if err := h.svc.ResetPassword(id, req.NewPassword); err != nil {
		if err == domain.ErrUserNotFound {
			response.NotFound(c, 40407, "user not found")
			return
		}
		response.InternalError(c, "failed to reset password")
		return
	}

	response.Success(c, nil)
}

type UpdateStatusRequest struct {
	Enabled bool `json:"enabled"`
}

func (h *AdminHandler) UpdateUserStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid user id")
		return
	}

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "enabled field is required")
		return
	}

	if err := h.svc.UpdateUserStatus(id, req.Enabled); err != nil {
		if err == domain.ErrUserNotFound {
			response.NotFound(c, 40407, "user not found")
			return
		}
		response.InternalError(c, "failed to update user status")
		return
	}

	response.Success(c, nil)
}

// ---- Model Management ----

func (h *AdminHandler) ListModels(c *gin.Context) {
	models, err := h.svc.ListModels()
	if err != nil {
		response.InternalError(c, "failed to list models")
		return
	}

	response.Success(c, models)
}

type CreateModelRequest struct {
	Name            string `json:"name" binding:"required"`
	ModelType       string `json:"model_type" binding:"required,oneof=text image video"`
	Provider        string `json:"provider" binding:"required"`
	Endpoint        string `json:"endpoint" binding:"required"`
	APIKey          string `json:"api_key" binding:"required"`
	ModelIdentifier string `json:"model_identifier" binding:"required"`
	MaxConcurrency  int    `json:"max_concurrency"`
	TimeoutSeconds  int    `json:"timeout_seconds"`
	IsDefault       bool   `json:"is_default"`
}

func (h *AdminHandler) CreateModel(c *gin.Context) {
	var req CreateModelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	if req.MaxConcurrency <= 0 {
		req.MaxConcurrency = 5
	}
	if req.TimeoutSeconds <= 0 {
		req.TimeoutSeconds = 60
	}

	m := &domain.AIModelConfig{
		Name:            req.Name,
		ModelType:       req.ModelType,
		Provider:        req.Provider,
		Endpoint:        req.Endpoint,
		ModelIdentifier: req.ModelIdentifier,
		MaxConcurrency:  req.MaxConcurrency,
		TimeoutSeconds:  req.TimeoutSeconds,
		IsDefault:       req.IsDefault,
		Enabled:         true,
	}

	result, err := h.svc.CreateModel(m, req.APIKey)
	if err != nil {
		response.InternalError(c, "failed to create model")
		return
	}

	response.Success(c, result)
}

func (h *AdminHandler) UpdateModel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid model id")
		return
	}

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "invalid request body")
		return
	}

	apiKey, _ := body["api_key"].(string)
	delete(body, "api_key")

	result, err := h.svc.UpdateModel(id, body, apiKey)
	if err != nil {
		if err == domain.ErrModelNotFound {
			response.NotFound(c, 40408, "model not found")
			return
		}
		response.InternalError(c, "failed to update model")
		return
	}

	response.Success(c, result)
}

func (h *AdminHandler) DeleteModel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid model id")
		return
	}

	if err := h.svc.DeleteModel(id); err != nil {
		response.InternalError(c, "failed to delete model")
		return
	}

	response.Success(c, nil)
}
