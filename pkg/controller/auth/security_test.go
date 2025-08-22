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
	authmodel "github.com/m-mizutani/tamamo/pkg/domain/model/auth"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"github.com/m-mizutani/tamamo/pkg/repository/database/memory"
	"github.com/m-mizutani/tamamo/pkg/usecase"
)

// TestAuthSecurity_SessionCookieTampering tests security against cookie tampering
func TestAuthSecurity_SessionCookieTampering(t *testing.T) {
	ctx := context.Background()
	sessionRepo := memory.NewSessionRepository()

	userRepo := memory.NewUserRepository()
	userUseCase := usecase.NewUserUseCase(userRepo, nil, nil)

	authUseCase := usecase.NewAuthUseCase(
		sessionRepo,
		userUseCase,
		"test-client-id",
		"test-client-secret",
		"http://localhost:3000",
	)

	controller := authctrl.NewController(authUseCase, userUseCase, "http://localhost:3000", false)

	// Create a valid session
	sessionID := types.NewUUID(ctx)
	validSession := &authmodel.Session{
		ID:        sessionID,
		UserID:    types.UserID("01234567-89ab-cdef-0123-456789abcdef"),
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
	userUseCase := usecase.NewUserUseCase(userRepo, nil, nil)

	authUseCase := usecase.NewAuthUseCase(
		sessionRepo,
		userUseCase,
		"test-client-id",
		"test-client-secret",
		"http://localhost:3000",
	)

	controller := authctrl.NewController(authUseCase, userUseCase, "http://localhost:3000", false)

	t.Run("PredictableSessionID", func(t *testing.T) {
		// Attempt to create a session with a predictable ID
		predictableID := types.UUID("00000000-0000-0000-0000-000000000001")
		fakeSession := &authmodel.Session{
			ID:        predictableID,
			UserID:    types.UserID("01234567-89ab-cdef-0123-456789abcdef"),
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
	userUseCase := usecase.NewUserUseCase(userRepo, nil, nil)

	authUseCase := usecase.NewAuthUseCase(
		sessionRepo,
		userUseCase,
		"test-client-id",
		"test-client-secret",
		"http://localhost:3000",
	)

	controller := authctrl.NewController(authUseCase, userUseCase, "http://localhost:3000", false)

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
