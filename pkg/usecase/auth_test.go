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

	// Note: In real tests, we would mock the Slack OAuth service
	// For now, we'll test the session management directly
	userRepo := memory.NewUserRepository()
	userUseCase := usecase.NewUserUseCase(userRepo, nil, nil)

	uc := usecase.NewAuthUseCase(
		sessionRepo,
		userUseCase,
		"test-client-id",
		"test-client-secret",
		"http://localhost:3000",
	)

	t.Run("Create and retrieve session", func(t *testing.T) {
		// Create a test session
		session := auth.NewSession(ctx, types.UserID("01234567-89ab-cdef-0123-456789abcdef"), "Test User", "test@example.com", "T123456", "Test Team")
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
		session := auth.NewSession(ctx, types.UserID("02345678-9abc-def0-1234-56789abcdef0"), "Another User", "another@example.com", "T234567", "Another Team")
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

	userRepo := memory.NewUserRepository()
	userUseCase := usecase.NewUserUseCase(userRepo, nil, nil)

	uc := usecase.NewAuthUseCase(
		sessionRepo,
		userUseCase,
		"test-client-id",
		"test-client-secret",
		"http://localhost:3000",
	)

	// Create an expired session
	sessionID := types.NewUUID(ctx)

	session := &auth.Session{
		ID:        sessionID,
		UserID:    types.UserID("03456789-abcd-ef01-2345-6789abcdef01"),
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

// TestOAuthState_CSRF tests are removed since OAuth state is now handled via cookies
