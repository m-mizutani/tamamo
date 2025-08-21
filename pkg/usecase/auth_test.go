package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/domain/model/auth"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"github.com/m-mizutani/tamamo/pkg/repository/database/memory"
	"github.com/m-mizutani/tamamo/pkg/usecase"
)

func TestAuthUseCase_SessionLifecycle(t *testing.T) {
	ctx := context.Background()
	sessionRepo := memory.NewSessionRepository()
	oauthStateRepo := memory.NewOAuthStateRepository()

	// Note: In real tests, we would mock the Slack OAuth service
	// For now, we'll test the session management directly
	uc := usecase.NewAuthUseCase(
		sessionRepo,
		oauthStateRepo,
		"test-client-id",
		"test-client-secret",
		"http://localhost:3000",
	)

	t.Run("Create and retrieve session", func(t *testing.T) {
		// Create a test session
		session := auth.NewSession(ctx, "U123456", "Test User", "test@example.com", "T123456", "Test Team")
		gt.NoError(t, sessionRepo.CreateSession(ctx, session))

		// Retrieve the session
		retrieved, err := uc.GetSession(ctx, session.ID.String())
		gt.NoError(t, err)
		gt.Equal(t, retrieved.ID, session.ID)
		gt.Equal(t, retrieved.UserID, session.UserID)
		gt.Equal(t, retrieved.Email, session.Email)
	})

	t.Run("Session not found", func(t *testing.T) {
		_, err := uc.GetSession(ctx, "non-existent-session")
		gt.Error(t, err)
		gt.Equal(t, err, auth.ErrSessionNotFound)
	})

	t.Run("Logout deletes session", func(t *testing.T) {
		// Create a test session
		session := auth.NewSession(ctx, "U234567", "Another User", "another@example.com", "T234567", "Another Team")
		gt.NoError(t, sessionRepo.CreateSession(ctx, session))

		// Logout
		gt.NoError(t, uc.Logout(ctx, session.ID.String()))

		// Try to retrieve the session
		_, err := uc.GetSession(ctx, session.ID.String())
		gt.Error(t, err)
		gt.Equal(t, err, auth.ErrSessionNotFound)
	})

	t.Run("Logout with non-existent session succeeds", func(t *testing.T) {
		// Logout with a non-existent session should not error
		err := uc.Logout(ctx, "non-existent-session")
		gt.NoError(t, err)
	})
}

func TestAuthUseCase_ExpiredSession(t *testing.T) {
	ctx := context.Background()
	sessionRepo := memory.NewSessionRepository()
	oauthStateRepo := memory.NewOAuthStateRepository()

	uc := usecase.NewAuthUseCase(
		sessionRepo,
		oauthStateRepo,
		"test-client-id",
		"test-client-secret",
		"http://localhost:3000",
	)

	// Create an expired session
	sessionID := types.NewUUID(ctx)

	session := &auth.Session{
		ID:        sessionID,
		UserID:    "U345678",
		UserName:  "Expired User",
		Email:     "expired@example.com",
		TeamID:    "T345678",
		TeamName:  "Expired Team",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		CreatedAt: time.Now().Add(-2 * time.Hour),
	}

	gt.NoError(t, sessionRepo.CreateSession(ctx, session))

	// Try to retrieve expired session
	_, getErr := uc.GetSession(ctx, session.ID.String())
	gt.Error(t, getErr)
	gt.Equal(t, getErr, auth.ErrSessionExpired)
}

func TestOAuthState_CSRF(t *testing.T) {
	ctx := context.Background()
	oauthStateRepo := memory.NewOAuthStateRepository()

	t.Run("Valid state", func(t *testing.T) {
		state, err := auth.NewOAuthState()
		gt.NoError(t, err)
		gt.NoError(t, oauthStateRepo.SaveState(ctx, state))

		// Validate and delete state
		err = oauthStateRepo.ValidateAndDeleteState(ctx, state.State)
		gt.NoError(t, err)

		// State should be deleted after validation
		err = oauthStateRepo.ValidateAndDeleteState(ctx, state.State)
		gt.Error(t, err)
		gt.Equal(t, err, auth.ErrStateNotFound)
	})

	t.Run("Expired state", func(t *testing.T) {
		// Create an expired state
		expiredState := &auth.OAuthState{
			State:     "expired-state-token",
			ExpiresAt: time.Now().Add(-1 * time.Minute), // Expired 1 minute ago
			CreatedAt: time.Now().Add(-6 * time.Minute),
		}

		// Note: SaveState will immediately clean up expired states in memory implementation
		gt.NoError(t, oauthStateRepo.SaveState(ctx, expiredState))

		// Try to validate expired state - it should be not found since it was cleaned up
		err := oauthStateRepo.ValidateAndDeleteState(ctx, expiredState.State)
		gt.Error(t, err)
		// Memory implementation cleans up expired states on save, so it returns NotFound
		gt.Equal(t, err, auth.ErrStateNotFound)
	})

	t.Run("Invalid state", func(t *testing.T) {
		err := oauthStateRepo.ValidateAndDeleteState(ctx, "invalid-state-token")
		gt.Error(t, err)
		gt.Equal(t, err, auth.ErrStateNotFound)
	})
}
