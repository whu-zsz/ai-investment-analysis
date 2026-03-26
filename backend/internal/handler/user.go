package handler

import (
	dtoResponse "stock-analysis-backend/internal/dto/response"
	"stock-analysis-backend/internal/dto/request"
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
// @Param request body request.RegisterRequest true "注册请求"
// @Success 200 {object} response.Response{data=dtoResponse.UserResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/auth/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req request.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	user, err := h.userService.Register(&req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, dtoResponse.UserResponse{
		ID:                   user.ID,
		Username:             user.Username,
		Email:                user.Email,
		Phone:                user.Phone,
		AvatarURL:            user.AvatarURL,
		InvestmentPreference: user.InvestmentPreference,
		TotalProfit:          user.TotalProfit.String(),
		RiskTolerance:        user.RiskTolerance,
	})
}

// Login godoc
// @Summary 用户登录
// @Description 用户登录获取token
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body request.LoginRequest true "登录请求"
// @Success 200 {object} response.Response{data=dtoResponse.LoginResponse}
// @Failure 401 {object} response.Response
// @Router /api/v1/auth/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req request.LoginRequest
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

	response.Success(c, dtoResponse.UserResponse{
		ID:                   user.ID,
		Username:             user.Username,
		Email:                user.Email,
		Phone:                user.Phone,
		AvatarURL:            user.AvatarURL,
		InvestmentPreference: user.InvestmentPreference,
		TotalProfit:          user.TotalProfit.String(),
		RiskTolerance:        user.RiskTolerance,
	})
}

// UpdateProfile godoc
// @Summary 更新用户信息
// @Description 更新当前用户信息
// @Tags 用户
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.UpdateProfileRequest true "更新请求"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /api/v1/user/profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetUint64("user_id")

	var req request.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.userService.UpdateProfile(userID, &req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}
