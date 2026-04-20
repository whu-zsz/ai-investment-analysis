package handler

import (
	dtoRequest "stock-analysis-backend/internal/dto/request"
	dtoResponse "stock-analysis-backend/internal/dto/response"
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/service"
	"stock-analysis-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// Register godoc
// @Summary 用户注册
// @Description 注册新用户
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dtoRequest.RegisterRequest true "注册请求"
// @Success 200 {object} response.Response{data=dtoResponse.UserResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/auth/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req dtoRequest.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	user, err := h.userService.Register(&req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, h.toUserResponse(user))
}

// Login godoc
// @Summary 用户登录
// @Description 用户登录获取token
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dtoRequest.LoginRequest true "登录请求"
// @Success 200 {object} response.Response{data=dtoResponse.LoginResponse}
// @Failure 401 {object} response.Response
// @Router /api/v1/auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req dtoRequest.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	loginResp, err := h.userService.Login(&req)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.Success(c, loginResp)
}

// Logout godoc
// @Summary 用户退出登录
// @Description 鉴权通过后返回退出确认，不执行服务端 token 撤销
// @Tags 认证
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Failure 401 {object} response.Response
// @Router /api/v1/auth/logout [post]
func (h *UserHandler) Logout(c *gin.Context) {
	response.Success(c, nil)
}

// GetProfile godoc
// @Summary 获取用户信息
// @Description 获取当前用户详细信息
// @Tags 用户
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=dtoResponse.UserResponse}
// @Failure 401 {object} response.Response
// @Router /api/v1/user/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.GetUint64("user_id")

	user, err := h.userService.GetProfile(userID)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	response.Success(c, h.toUserResponse(user))
}

// UpdateProfile godoc
// @Summary 更新用户信息
// @Description 更新当前用户信息
// @Tags 用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dtoRequest.UpdateProfileRequest true "更新请求"
// @Success 200 {object} response.Response{data=dtoResponse.UserResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/user/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetUint64("user_id")

	var req dtoRequest.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	user, err := h.userService.UpdateProfile(userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, h.toUserResponse(user))
}

func (h *UserHandler) toUserResponse(user *model.User) dtoResponse.UserResponse {
	return dtoResponse.UserResponse{
		ID:                   user.ID,
		Username:             user.Username,
		Email:                user.Email,
		Phone:                user.Phone,
		AvatarURL:            user.AvatarURL,
		InvestmentPreference: user.InvestmentPreference,
		TotalProfit:          user.TotalProfit.String(),
		RiskTolerance:        user.RiskTolerance,
	}
}
