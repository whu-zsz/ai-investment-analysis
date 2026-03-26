package service

import (
	"errors"
	"stock-analysis-backend/internal/config"
	"stock-analysis-backend/internal/dto/request"
	"stock-analysis-backend/internal/dto/response"
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/repository"
	"stock-analysis-backend/internal/utils"
)

type UserService interface {
	Register(req *request.RegisterRequest) (*model.User, error)
	Login(req *request.LoginRequest) (*response.LoginResponse, error)
	GetProfile(userID uint64) (*model.User, error)
	UpdateProfile(userID uint64, req *request.UpdateProfileRequest) error
}

type userService struct {
	userRepo repository.UserRepository
	jwtCfg   config.JWTConfig
}

func NewUserService(userRepo repository.UserRepository, jwtCfg config.JWTConfig) UserService {
	return &userService{
		userRepo: userRepo,
		jwtCfg:   jwtCfg,
	}
}

func (s *userService) Register(req *request.RegisterRequest) (*model.User, error) {
	// 检查用户名是否已存在
	if _, err := s.userRepo.FindByUsername(req.Username); err == nil {
		return nil, errors.New("username already exists")
	}

	// 检查邮箱是否已存在
	if _, err := s.userRepo.FindByEmail(req.Email); err == nil {
		return nil, errors.New("email already exists")
	}

	// 密码加密
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// 创建用户
	user := &model.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: passwordHash,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *userService) Login(req *request.LoginRequest) (*response.LoginResponse, error) {
	// 查找用户
	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	// 验证密码
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, errors.New("invalid username or password")
	}

	// 检查用户是否激活
	if !user.IsActive {
		return nil, errors.New("user account is deactivated")
	}

	// 生成JWT token
	token, err := utils.GenerateToken(user.ID, user.Username, s.jwtCfg.Secret, s.jwtCfg.ExpireHours)
	if err != nil {
		return nil, err
	}

	// 更新最后登录时间
	_ = s.userRepo.UpdateLastLogin(user.ID)

	// 返回响应
	return &response.LoginResponse{
		Token: token,
		User: response.UserResponse{
			ID:                   user.ID,
			Username:             user.Username,
			Email:                user.Email,
			Phone:                user.Phone,
			AvatarURL:            user.AvatarURL,
			InvestmentPreference: user.InvestmentPreference,
			TotalProfit:          user.TotalProfit.String(),
			RiskTolerance:        user.RiskTolerance,
		},
	}, nil
}

func (s *userService) GetProfile(userID uint64) (*model.User, error) {
	return s.userRepo.FindByID(userID)
}

func (s *userService) UpdateProfile(userID uint64, req *request.UpdateProfileRequest) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}

	if req.Phone != nil {
		user.Phone = req.Phone
	}
	if req.AvatarURL != nil {
		user.AvatarURL = req.AvatarURL
	}
	if req.InvestmentPreference != nil {
		user.InvestmentPreference = *req.InvestmentPreference
	}

	return s.userRepo.Update(user)
}
