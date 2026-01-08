package auth

// LoginResponse is the API response for login/register.
type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresAt    int64        `json:"expires_at"`
	User         UserResponse `json:"user"`
}

// TokenResponse is the API response for token refresh.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

// UserResponse is the API representation of a user.
type UserResponse struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}
