package handler

import (
	"github.com/2hot4you/aiaos/backend/internal/domain"
	"github.com/2hot4you/aiaos/backend/internal/middleware"
	"github.com/2hot4you/aiaos/backend/internal/service"
	"github.com/2hot4you/aiaos/backend/pkg/response"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "username and password are required")
		return
	}

	result, err := h.authSvc.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		switch err {
		case domain.ErrInvalidCredential:
			response.Unauthorized(c, 40101, "用户名或密码错误")
		case domain.ErrAccountDisabled:
			response.Unauthorized(c, 40103, "账号已被禁用")
		case domain.ErrAccountLocked:
			response.Unauthorized(c, 40104, "账号已被锁定，请15分钟后重试")
		default:
			response.InternalError(c, "login failed")
		}
		return
	}

	response.Success(c, result)
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	userID := middleware.GetUserID(c)
	user, err := h.authSvc.GetCurrentUser(userID)
	if err != nil {
		response.NotFound(c, 40407, "user not found")
		return
	}
	response.Success(c, user)
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "old_password and new_password(min 8 chars) are required")
		return
	}

	userID := middleware.GetUserID(c)
	err := h.authSvc.ChangePassword(userID, req.OldPassword, req.NewPassword)
	if err != nil {
		switch err {
		case domain.ErrOldPasswordWrong:
			response.BadRequest(c, "旧密码不正确")
		case domain.ErrUserNotFound:
			response.NotFound(c, 40407, "user not found")
		default:
			response.InternalError(c, "change password failed")
		}
		return
	}

	response.Success(c, nil)
}
