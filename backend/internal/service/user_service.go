package service

import (
	"errors"
	"strings"
	"time"

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
	Logout(userID uint64, tokenJTI string, tokenExpiresAt time.Time) error
	GetProfile(userID uint64) (*model.User, error)
	UpdateProfile(userID uint64, req *request.UpdateProfileRequest) (*model.User, error)
}

type userService struct {
	userRepo         repository.UserRepository
	revokedTokenRepo repository.RevokedTokenRepository
	jwtCfg           config.JWTConfig
}

func NewUserService(userRepo repository.UserRepository, revokedTokenRepo repository.RevokedTokenRepository, jwtCfg config.JWTConfig) UserService {
	return &userService{
		userRepo:         userRepo,
		revokedTokenRepo: revokedTokenRepo,
		jwtCfg:           jwtCfg,
	}
}

func (s *userService) Register(req *request.RegisterRequest) (*model.User, error) {
	if _, err := s.userRepo.FindByUsername(req.Username); err == nil {
		return nil, errors.New("username already exists")
	}

	if _, err := s.userRepo.FindByEmail(req.Email); err == nil {
		return nil, errors.New("email already exists")
	}

	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

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
	user, err := s.userRepo.FindByUsername(req.Username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, errors.New("invalid username or password")
	}

	if !user.IsActive {
		return nil, errors.New("user account is deactivated")
	}

	token, err := utils.GenerateToken(user.ID, user.Username, s.jwtCfg.Secret, s.jwtCfg.ExpireHours)
	if err != nil {
		return nil, err
	}

	_ = s.userRepo.UpdateLastLogin(user.ID)

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

func (s *userService) Logout(userID uint64, tokenJTI string, tokenExpiresAt time.Time) error {
	tokenJTI = strings.TrimSpace(tokenJTI)
	if tokenJTI == "" {
		return errors.New("token jti is required")
	}

	revokedToken := &model.RevokedToken{
		UserID:         userID,
		JTI:            tokenJTI,
		TokenExpiresAt: tokenExpiresAt,
		RevokedAt:      time.Now(),
		Reason:         "logout",
	}

	if err := s.revokedTokenRepo.Create(revokedToken); err != nil {
		if repository.IsDuplicateRevokedTokenError(err) {
			return nil
		}
		return err
	}

	return nil
}

func (s *userService) GetProfile(userID uint64) (*model.User, error) {
	return s.userRepo.FindByID(userID)
}

func (s *userService) UpdateProfile(userID uint64, req *request.UpdateProfileRequest) (*model.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
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

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}
