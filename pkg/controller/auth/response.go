package auth

// UserResponse represents user information in API responses
type UserResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
	TeamID      string `json:"team_id"`
	TeamName    string `json:"team_name"`
}

// AuthCheckResponse represents the authentication check response
type AuthCheckResponse struct {
	Authenticated bool          `json:"authenticated"`
	User          *UserResponse `json:"user,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error details
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
