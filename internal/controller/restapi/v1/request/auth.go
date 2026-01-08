package request

// Register represents the registration request.
type Register struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required,min=8" example:"password123"`
	Name     string `json:"name" validate:"required,min=2" example:"John Doe"`
}

// Login represents the login request.
type Login struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required" example:"password123"`
}

// RefreshToken represents the refresh token request.
type RefreshToken struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
