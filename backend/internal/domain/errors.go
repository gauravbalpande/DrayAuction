package domain

import "net/http"

type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	HTTPStatus int    `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

func New(code, message string, status int) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: status}
}

var (
	ErrNotFound           = New("NOT_FOUND", "Resource not found", http.StatusNotFound)
	ErrUnauthorized       = New("UNAUTHORIZED", "Authentication required", http.StatusUnauthorized)
	ErrInvalidCredentials = New("INVALID_CREDENTIALS", "Invalid email or password", http.StatusUnauthorized)
	ErrEmailTaken         = New("EMAIL_TAKEN", "Email already registered", http.StatusConflict)
	ErrUsernameTaken      = New("USERNAME_TAKEN", "Username already taken", http.StatusConflict)
	ErrValidation         = New("VALIDATION_ERROR", "Invalid request data", http.StatusBadRequest)
	ErrInsufficientBudget = New("INSUFFICIENT_BUDGET", "Bid exceeds remaining budget", http.StatusBadRequest)
	ErrAuctionNotLive     = New("AUCTION_NOT_LIVE", "Auction is not in live state", http.StatusBadRequest)
	ErrBidTooLow          = New("BID_TOO_LOW", "Bid amount is too low", http.StatusBadRequest)
	ErrRateLimited        = New("RATE_LIMITED", "Too many requests", http.StatusTooManyRequests)
)
