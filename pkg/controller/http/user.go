package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
)

// UserController handles user-related HTTP requests
type UserController struct {
	userUseCase interfaces.UserUseCases
}

// UserInfoResponse represents the response structure for user information
type UserInfoResponse struct {
	ID          string    `json:"id"`
	SlackName   string    `json:"slack_name"`
	DisplayName string    `json:"display_name"`
	Email       string    `json:"email,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// NewUserController creates a new user controller
func NewUserController(userUseCase interfaces.UserUseCases) *UserController {
	return &UserController{
		userUseCase: userUseCase,
	}
}

// HandleGetUserAvatar returns user avatar data
func (c *UserController) HandleGetUserAvatar(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Validate UserID
	if !types.UserID(userID).IsValid() {
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	// Parse size parameter (optional, defaults to 48)
	size := 48
	if sizeStr := r.URL.Query().Get("size"); sizeStr != "" {
		if parsedSize, err := strconv.Atoi(sizeStr); err == nil && parsedSize > 0 && parsedSize <= 512 {
			size = parsedSize
		}
	}

	// Get avatar data from use case
	avatarData, err := c.userUseCase.GetUserAvatar(r.Context(), types.UserID(userID), size)
	if err != nil {
		// Check if it's a user not found error
		if strings.Contains(err.Error(), "user not found") {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set appropriate headers
	w.Header().Set("Content-Type", "image/jpeg") // Most Slack avatars are JPEG
	w.Header().Set("Content-Length", strconv.Itoa(len(avatarData)))
	w.Header().Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour

	// Write avatar data
	if _, err := w.Write(avatarData); err != nil {
		// Log error but don't return error response as headers are already sent
		return
	}
}

// HandleGetUserInfo returns user information
func (c *UserController) HandleGetUserInfo(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userID")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	// Validate UserID
	if !types.UserID(userID).IsValid() {
		http.Error(w, "Invalid user ID format", http.StatusBadRequest)
		return
	}

	// Get user information
	user, err := c.userUseCase.GetUserByID(r.Context(), types.UserID(userID))
	if err != nil {
		if strings.Contains(err.Error(), "user not found") {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return user info as JSON
	w.Header().Set("Content-Type", "application/json")

	// Create a safe response structure (excluding sensitive information)
	response := UserInfoResponse{
		ID:          user.ID.String(),
		SlackName:   user.SlackName,
		DisplayName: user.DisplayName,
		Email:       user.Email,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}
