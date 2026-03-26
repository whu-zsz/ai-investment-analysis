package response

type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type UserResponse struct {
	ID                   uint64  `json:"id"`
	Username             string  `json:"username"`
	Email                string  `json:"email"`
	Phone                *string `json:"phone"`
	AvatarURL            *string `json:"avatar_url"`
	InvestmentPreference string  `json:"investment_preference"`
	TotalProfit          string  `json:"total_profit"`
	RiskTolerance        string  `json:"risk_tolerance"`
}
