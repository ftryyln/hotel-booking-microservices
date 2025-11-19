package dto

// RegisterRequest for creating new users.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

// LoginRequest for authentication.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AuthResponse returns JWT tokens.
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// ProfileResponse shows user data.
type ProfileResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}
