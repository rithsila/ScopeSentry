package user

import (
	"github.com/Autumn-27/ScopeSentry/internal/api/response"
	"github.com/Autumn-27/ScopeSentry/internal/runtime/testinit"
	"github.com/Autumn-27/ScopeSentry/internal/services/user"
	"github.com/gin-gonic/gin"
)

var userService user.Service

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	NewPassword string `json:"newPassword" binding:"required" example:"newStrongPassword123"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"admin"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// Login 用户登录
// @Summary      用户登录
// @Description  用户登录接口
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        loginRequest  body      LoginRequest  true  "登录信息"
// @Success      200  {object}  response.SuccessResponse{data=object{access_token=string}}
// @Failure      400  {object}  response.BadRequestResponse
// @Failure      404  {object}  response.NotFoundResponse
// @Failure      500  {object}  response.InternalServerErrorResponse
// @Router       /users/login [post]
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	token, err := userService.Login(c, req.Username, req.Password)
	if err != nil {
		response.BadRequest(c, "api.user.login.failed", err)
		return
	}

	response.Success(c, gin.H{"access_token": token}, "api.user.login.success")
}

// ChangePassword 用户修改密码
// @Summary      修改当前用户密码
// @Description  通过 JWT 获取用户信息，修改其密码
// @Tags         user
// @Accept       json
// @Produce      json
// @Security     ApiKeyAuth
// @Param        changePasswordRequest  body      ChangePasswordRequest  true  "修改密码信息"
// @Success      200  {object}  response.SuccessResponse
// @Failure      400  {object}  response.BadRequestResponse
// @Failure      401  {object}  response.UnauthorizedResponse
// @Failure      500  {object}  response.InternalServerErrorResponse
// @Router       /user/changePassword [post]
func ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "api.bad_request", err)
		return
	}

	// 从 JWT 中间件写入的上下文获取 userID
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "Unauthorized", nil)
		return
	}
	userID, ok := userIDVal.(string)
	if !ok || userID == "" {
		response.Unauthorized(c, "Invalid token payload", nil)
		return
	}

	if err := userService.ChangePassword(c, userID, req.NewPassword); err != nil {
		response.InternalServerError(c, "api.user.change_password.failed", err)
		return
	}
	response.Success(c, nil, "api.user.change_password.success")
}

func init() {
	if testinit.Enabled {
		return
	}
	userService = user.NewService()
}
