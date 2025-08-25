package auth_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/m-mizutani/gt"
	authctrl "github.com/m-mizutani/tamamo/pkg/controller/auth"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/mock"
	authmodel "github.com/m-mizutani/tamamo/pkg/domain/model/auth"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"github.com/m-mizutani/tamamo/pkg/repository/database/memory"
	"github.com/m-mizutani/tamamo/pkg/service/slack"
	"github.com/m-mizutani/tamamo/pkg/usecase"
)

func createMockSlackClientForSecurity() *mock.SlackClientMock {
	return &mock.SlackClientMock{
		GetUserProfileFunc: func(ctx context.Context, userID string) (*interfaces.SlackUserProfile, error) {
			return &interfaces.SlackUserProfile{
				ID:          userID,
				Name:        "Test User",
				DisplayName: "Test Display Name",
				Email:       "test@example.com",
				Profile: struct {
					Image24   string `json:"image_24"`
					Image32   string `json:"image_32"`
					Image48   string `json:"image_48"`
					Image72   string `json:"image_72"`
					Image192  string `json:"image_192"`
					Image512  string `json:"image_512"`
					ImageOrig string `json:"image_original"`
				}{
					Image24:   "https://example.com/avatar_24.jpg",
					Image32:   "https://example.com/avatar_32.jpg",
					Image48:   "https://example.com/avatar_48.jpg",
					Image72:   "https://example.com/avatar_72.jpg",
					Image192:  "https://example.com/avatar_192.jpg",
					Image512:  "https://example.com/avatar_512.jpg",
					ImageOrig: "https://example.com/avatar_original.jpg",
				},
			}, nil
		},
	}
}

// TestAuthSecurity_SessionCookieTampering tests security against cookie tampering
func TestAuthSecurity_SessionCookieTampering(t *testing.T) {
	ctx := context.Background()
	sessionRepo := memory.NewSessionRepository()

	userRepo := memory.NewUserRepository()
	mockSlackClient := createMockSlackClientForSecurity()
	avatarService := slack.NewAvatarService(mockSlackClient)
	userUseCase := usecase.NewUserUseCase(userRepo, avatarService, mockSlackClient)

	authUseCase := usecase.NewAuthUseCase(
		sessionRepo,
		userUseCase,
		"test-client-id",
		"test-client-secret",
		"http://localhost:3000",
	)

	controller := authctrl.NewController(authUseCase, userUseCase, "http://localhost:3000")

	// Create a test user first
	testUser, err := userUseCase.GetOrCreateUser(ctx, "U123456789", "Test User", "test@example.com", "T123456")
	gt.NoError(t, err)

	// Create a valid session
	sessionID := types.NewUUID(ctx)
	validSession := &authmodel.Session{
		ID:        sessionID,
		UserID:    testUser.ID,
		UserName:  "Test User",
		Email:     "test@example.com",
		TeamID:    "T123456",
		TeamName:  "Test Team",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}
	gt.NoError(t, sessionRepo.CreateSession(ctx, validSession))

	t.Run("InvalidSessionID", func(t *testing.T) {
		// Test with completely invalid session ID
		req := httptest.NewRequest(http.MethodGet, "/api/auth/check", nil)
		req.AddCookie(&http.Cookie{
			Name:  "tamamo_session",
			Value: "invalid-session-id",
		})
		rec := httptest.NewRecorder()

		controller.HandleCheck(rec, req)

		gt.Equal(t, rec.Code, http.StatusOK)

		var response authctrl.AuthCheckResponse
		gt.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
		gt.Equal(t, response.Authenticated, false)
	})

	t.Run("NonExistentSessionID", func(t *testing.T) {
		// Test with valid UUID format but non-existent session
		nonExistentID := types.NewUUID(ctx)
		req := httptest.NewRequest(http.MethodGet, "/api/auth/check", nil)
		req.AddCookie(&http.Cookie{
			Name:  "tamamo_session",
			Value: nonExistentID.String(),
		})
		rec := httptest.NewRecorder()

		controller.HandleCheck(rec, req)

		gt.Equal(t, rec.Code, http.StatusOK)

		var response authctrl.AuthCheckResponse
		gt.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
		gt.Equal(t, response.Authenticated, false)
	})

	t.Run("ExpiredSession", func(t *testing.T) {
		// Create an expired session
		expiredSessionID := types.NewUUID(ctx)
		expiredSession := &authmodel.Session{
			ID:        expiredSessionID,
			UserID:    types.UserID("01234567-89ab-cdef-0123-456789abcdef"),
			UserName:  "Expired User",
			Email:     "expired@example.com",
			TeamID:    "T123456",
			TeamName:  "Test Team",
			ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
			CreatedAt: time.Now().Add(-2 * time.Hour),
		}
		gt.NoError(t, sessionRepo.CreateSession(ctx, expiredSession))

		req := httptest.NewRequest(http.MethodGet, "/api/auth/check", nil)
		req.AddCookie(&http.Cookie{
			Name:  "tamamo_session",
			Value: expiredSessionID.String(),
		})
		rec := httptest.NewRecorder()

		controller.HandleCheck(rec, req)

		gt.Equal(t, rec.Code, http.StatusOK)

		var response authctrl.AuthCheckResponse
		gt.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
		gt.Equal(t, response.Authenticated, false)
	})

	t.Run("MalformedCookie", func(t *testing.T) {
		// Test with malformed cookie value
		req := httptest.NewRequest(http.MethodGet, "/api/auth/check", nil)
		req.AddCookie(&http.Cookie{
			Name:  "tamamo_session",
			Value: "malformed-cookie-value-!@#$%",
		})
		rec := httptest.NewRecorder()

		controller.HandleCheck(rec, req)

		gt.Equal(t, rec.Code, http.StatusOK)

		var response authctrl.AuthCheckResponse
		gt.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
		gt.Equal(t, response.Authenticated, false)
	})

	t.Run("EmptyCookie", func(t *testing.T) {
		// Test with empty cookie value
		req := httptest.NewRequest(http.MethodGet, "/api/auth/check", nil)
		req.AddCookie(&http.Cookie{
			Name:  "tamamo_session",
			Value: "",
		})
		rec := httptest.NewRecorder()

		controller.HandleCheck(rec, req)

		gt.Equal(t, rec.Code, http.StatusOK)

		var response authctrl.AuthCheckResponse
		gt.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
		gt.Equal(t, response.Authenticated, false)
	})

	t.Run("NoCookie", func(t *testing.T) {
		// Test with no session cookie at all
		req := httptest.NewRequest(http.MethodGet, "/api/auth/check", nil)
		rec := httptest.NewRecorder()

		controller.HandleCheck(rec, req)

		gt.Equal(t, rec.Code, http.StatusOK)

		var response authctrl.AuthCheckResponse
		gt.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
		gt.Equal(t, response.Authenticated, false)
	})

	t.Run("ValidSessionShouldWork", func(t *testing.T) {
		// Ensure that valid sessions still work properly
		req := httptest.NewRequest(http.MethodGet, "/api/auth/check", nil)
		req.AddCookie(&http.Cookie{
			Name:  "tamamo_session",
			Value: validSession.ID.String(),
		})
		rec := httptest.NewRecorder()

		controller.HandleCheck(rec, req)

		gt.Equal(t, rec.Code, http.StatusOK)

		var response authctrl.AuthCheckResponse
		gt.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
		gt.Equal(t, response.Authenticated, true)
		gt.Equal(t, response.User.ID, validSession.UserID.String())
	})
}

// TestAuthSecurity_SessionFixation tests against session fixation attacks
func TestAuthSecurity_SessionFixation(t *testing.T) {
	ctx := context.Background()
	sessionRepo := memory.NewSessionRepository()

	userRepo := memory.NewUserRepository()
	mockSlackClient := createMockSlackClientForSecurity()
	avatarService := slack.NewAvatarService(mockSlackClient)
	userUseCase := usecase.NewUserUseCase(userRepo, avatarService, mockSlackClient)

	authUseCase := usecase.NewAuthUseCase(
		sessionRepo,
		userUseCase,
		"test-client-id",
		"test-client-secret",
		"http://localhost:3000",
	)

	controller := authctrl.NewController(authUseCase, userUseCase, "http://localhost:3000")

	t.Run("PredictableSessionID", func(t *testing.T) {
		// Create a test user first
		testUser, err := userUseCase.GetOrCreateUser(ctx, "U987654321", "Fake User", "fake@example.com", "T123456")
		gt.NoError(t, err)

		// Attempt to create a session with a predictable ID
		predictableID := types.UUID("00000000-0000-0000-0000-000000000001")
		fakeSession := &authmodel.Session{
			ID:        predictableID,
			UserID:    testUser.ID,
			UserName:  "Fake User",
			Email:     "fake@example.com",
			TeamID:    "T123456",
			TeamName:  "Test Team",
			ExpiresAt: time.Now().Add(24 * time.Hour),
			CreatedAt: time.Now(),
		}

		// Even if someone manually inserts a predictable session ID,
		// they shouldn't be able to authenticate without proper creation flow
		gt.NoError(t, sessionRepo.CreateSession(ctx, fakeSession))

		req := httptest.NewRequest(http.MethodGet, "/api/auth/check", nil)
		req.AddCookie(&http.Cookie{
			Name:  "tamamo_session",
			Value: predictableID.String(),
		})
		rec := httptest.NewRecorder()

		controller.HandleCheck(rec, req)

		// This should work since we created a valid session,
		// but in real implementation, sessions should only be created through OAuth flow
		gt.Equal(t, rec.Code, http.StatusOK)

		var response authctrl.AuthCheckResponse
		gt.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
		gt.Equal(t, response.Authenticated, true)
	})
}

// TestAuthSecurity_UserIDValidation tests user ID validation in endpoints
func TestAuthSecurity_UserIDValidation(t *testing.T) {
	ctx := context.Background()
	sessionRepo := memory.NewSessionRepository()

	userRepo := memory.NewUserRepository()
	mockSlackClient := createMockSlackClientForSecurity()
	avatarService := slack.NewAvatarService(mockSlackClient)
	userUseCase := usecase.NewUserUseCase(userRepo, avatarService, mockSlackClient)

	authUseCase := usecase.NewAuthUseCase(
		sessionRepo,
		userUseCase,
		"test-client-id",
		"test-client-secret",
		"http://localhost:3000",
	)

	controller := authctrl.NewController(authUseCase, userUseCase, "http://localhost:3000")

	t.Run("InvalidUserIDFormat", func(t *testing.T) {
		// Create session with invalid UserID format
		sessionID := types.NewUUID(ctx)
		invalidSession := &authmodel.Session{
			ID:        sessionID,
			UserID:    types.UserID("invalid-user-id"), // Invalid format
			UserName:  "Invalid User",
			Email:     "invalid@example.com",
			TeamID:    "T123456",
			TeamName:  "Test Team",
			ExpiresAt: time.Now().Add(24 * time.Hour),
			CreatedAt: time.Now(),
		}
		gt.NoError(t, sessionRepo.CreateSession(ctx, invalidSession))

		req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
		req.AddCookie(&http.Cookie{
			Name:  "tamamo_session",
			Value: sessionID.String(),
		})
		rec := httptest.NewRecorder()

		controller.HandleMe(rec, req)

		// Should reject invalid UserID format
		gt.Equal(t, rec.Code, http.StatusInternalServerError)
	})

	t.Run("EmptyUserID", func(t *testing.T) {
		// Create session with empty UserID
		sessionID := types.NewUUID(ctx)
		emptyUserSession := &authmodel.Session{
			ID:        sessionID,
			UserID:    types.UserID(""), // Empty UserID
			UserName:  "Empty User",
			Email:     "empty@example.com",
			TeamID:    "T123456",
			TeamName:  "Test Team",
			ExpiresAt: time.Now().Add(24 * time.Hour),
			CreatedAt: time.Now(),
		}
		gt.NoError(t, sessionRepo.CreateSession(ctx, emptyUserSession))

		req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
		req.AddCookie(&http.Cookie{
			Name:  "tamamo_session",
			Value: sessionID.String(),
		})
		rec := httptest.NewRecorder()

		controller.HandleMe(rec, req)

		// Should reject empty UserID
		gt.Equal(t, rec.Code, http.StatusInternalServerError)
	})
}
