package http_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/m-mizutani/gt"
	httpctrl "github.com/m-mizutani/tamamo/pkg/controller/http"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"github.com/m-mizutani/tamamo/pkg/repository/database/memory"
	"github.com/m-mizutani/tamamo/pkg/service/slack"
	"github.com/m-mizutani/tamamo/pkg/usecase"
)

// TestUserIntegration_EndToEndUserCreationAndAvatarAccess tests the complete flow
// from user creation to avatar access through HTTP endpoints
func TestUserIntegration_EndToEndUserCreationAndAvatarAccess(t *testing.T) {
	ctx := context.Background()

	// Setup repositories and services
	userRepo := memory.NewUserRepository()
	avatarService := slack.NewAvatarService(nil) // Use default avatar service
	userUseCase := usecase.NewUserUseCase(userRepo, avatarService, nil)

	// Create HTTP controller
	userController := httpctrl.NewUserController(userUseCase)

	// Create HTTP server
	server := httpctrl.New(
		httpctrl.WithUserController(userController),
	)

	t.Run("CreateUserAndAccessAvatar", func(t *testing.T) {
		// Step 1: Create a user via use case (simulating OAuth flow)
		slackID := "U123456789"
		slackName := "Test User"
		email := "test@example.com"
		teamID := "T123456789"

		createdUser, err := userUseCase.GetOrCreateUser(ctx, slackID, slackName, email, teamID)
		gt.NoError(t, err)
		gt.V(t, createdUser).NotNil()
		gt.Equal(t, createdUser.SlackID, slackID)
		gt.Equal(t, createdUser.SlackName, slackName)

		// Step 2: Access user avatar via HTTP endpoint
		req := httptest.NewRequest("GET", "/api/users/"+createdUser.ID.String()+"/avatar", nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		gt.Equal(t, w.Code, http.StatusOK)
		gt.Equal(t, w.Header().Get("Content-Type"), "image/jpeg")
		gt.V(t, len(w.Body.Bytes()) > 0).Equal(true) // Should have avatar data

		// Step 3: Access user info via HTTP endpoint
		req = httptest.NewRequest("GET", "/api/users/"+createdUser.ID.String(), nil)
		w = httptest.NewRecorder()

		server.ServeHTTP(w, req)

		gt.Equal(t, w.Code, http.StatusOK)
		gt.Equal(t, w.Header().Get("Content-Type"), "application/json")

		var response httpctrl.UserInfoResponse
		gt.NoError(t, json.NewDecoder(w.Body).Decode(&response))
		gt.Equal(t, response.ID, createdUser.ID.String())
		gt.Equal(t, response.SlackName, slackName)
		gt.Equal(t, response.Email, email)
	})

	t.Run("UserUpdateFlow", func(t *testing.T) {
		// Step 1: Create initial user
		slackID := "U987654321"
		originalName := "Original Name"
		originalEmail := "original@example.com"
		teamID := "T987654321"

		user1, err := userUseCase.GetOrCreateUser(ctx, slackID, originalName, originalEmail, teamID)
		gt.NoError(t, err)

		// Step 2: Update user info (simulating Slack info change)
		updatedName := "Updated Name"
		updatedEmail := "updated@example.com"

		user2, err := userUseCase.GetOrCreateUser(ctx, slackID, updatedName, updatedEmail, teamID)
		gt.NoError(t, err)

		// Should be same user ID but updated info
		gt.Equal(t, user1.ID, user2.ID)
		gt.Equal(t, user2.SlackName, updatedName)
		gt.Equal(t, user2.Email, updatedEmail)

		// Step 3: Verify via HTTP endpoint
		req := httptest.NewRequest("GET", "/api/users/"+user2.ID.String(), nil)
		w := httptest.NewRecorder()

		server.ServeHTTP(w, req)

		gt.Equal(t, w.Code, http.StatusOK)

		var response httpctrl.UserInfoResponse
		gt.NoError(t, json.NewDecoder(w.Body).Decode(&response))
		gt.Equal(t, response.SlackName, updatedName)
		gt.Equal(t, response.Email, updatedEmail)
	})

	t.Run("AvatarCaching", func(t *testing.T) {
		// Create a user
		user, err := userUseCase.GetOrCreateUser(ctx, "U111111111", "Cache Test", "cache@example.com", "T111111111")
		gt.NoError(t, err)

		// First request - should fetch avatar
		req := httptest.NewRequest("GET", "/api/users/"+user.ID.String()+"/avatar", nil)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		gt.Equal(t, w.Code, http.StatusOK)
		firstResponse := w.Body.Bytes()

		// Second request - should return cached avatar
		req = httptest.NewRequest("GET", "/api/users/"+user.ID.String()+"/avatar", nil)
		w = httptest.NewRecorder()
		server.ServeHTTP(w, req)
		gt.Equal(t, w.Code, http.StatusOK)
		secondResponse := w.Body.Bytes()

		// Should be same data (from cache)
		gt.Equal(t, len(firstResponse), len(secondResponse))
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		// Test with invalid user ID
		req := httptest.NewRequest("GET", "/api/users/invalid-uuid/avatar", nil)
		w := httptest.NewRecorder()
		server.ServeHTTP(w, req)
		gt.Equal(t, w.Code, http.StatusBadRequest)

		// Test with non-existent user ID
		nonExistentID := types.NewUserID(ctx)
		req = httptest.NewRequest("GET", "/api/users/"+nonExistentID.String()+"/avatar", nil)
		w = httptest.NewRecorder()
		server.ServeHTTP(w, req)
		gt.Equal(t, w.Code, http.StatusNotFound)
	})
}

// TestUserIntegration_ConcurrentAccess tests concurrent access to user endpoints
func TestUserIntegration_ConcurrentAccess(t *testing.T) {
	ctx := context.Background()

	// Setup
	userRepo := memory.NewUserRepository()
	avatarService := slack.NewAvatarService(nil)
	userUseCase := usecase.NewUserUseCase(userRepo, avatarService, nil)
	userController := httpctrl.NewUserController(userUseCase)
	server := httpctrl.New(httpctrl.WithUserController(userController))

	// Create a user
	user, err := userUseCase.GetOrCreateUser(ctx, "U999999999", "Concurrent Test", "concurrent@example.com", "T999999999")
	gt.NoError(t, err)

	// Test concurrent avatar requests
	const numRequests = 10
	results := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/api/users/"+user.ID.String()+"/avatar", nil)
			w := httptest.NewRecorder()
			server.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				results <- fmt.Errorf("expected status 200, got %d", w.Code)
				return
			}
			results <- nil
		}()
	}

	// Collect results
	for i := 0; i < numRequests; i++ {
		err := <-results
		gt.NoError(t, err)
	}
}
