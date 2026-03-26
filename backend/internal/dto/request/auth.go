package request

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=20"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UpdateProfileRequest struct {
	Phone                *string `json:"phone" binding:"omitempty"`
	AvatarURL            *string `json:"avatar_url" binding:"omitempty"`
	InvestmentPreference *string `json:"investment_preference" binding:"omitempty,oneof=conservative balanced aggressive"`
}
