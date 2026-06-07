package dto

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type CreateAuctionRequest struct {
	AIOpponents int    `json:"ai_opponents" binding:"required,min=1,max=5"`
	Difficulty  string `json:"difficulty" binding:"required,oneof=easy medium hard legendary"`
}

type BidRequest struct {
	Amount int64 `json:"amount" binding:"required,min=1"`
}

type AuthResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type UserResponse struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	Coins      int64  `json:"coins"`
	XP         int64  `json:"xp"`
	Rank       string `json:"rank"`
	RankPoints int    `json:"rank_points"`
	Wins       int    `json:"wins"`
	Losses     int    `json:"losses"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
