package auth_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func createMockSlackClient() *mock.SlackClientMock {
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

func TestAuthController_Login(t *testing.T) {
	sessionRepo := memory.NewSessionRepository()
	userRepo := memory.NewUserRepository()
	mockSlackClient := createMockSlackClient()
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

	// Test login redirect
	req := httptest.NewRequest(http.MethodGet, "/api/auth/login", nil)
	rec := httptest.NewRecorder()

	controller.HandleLogin(rec, req)

	gt.Equal(t, rec.Code, http.StatusTemporaryRedirect)
	location := rec.Header().Get("Location")
	gt.True(t, location != "")
	gt.True(t, strings.Contains(location, "slack.com"))
}

func TestAuthController_Callback(t *testing.T) {
	sessionRepo := memory.NewSessionRepository()

	userRepo := memory.NewUserRepository()
	mockSlackClient := createMockSlackClient()
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

	t.Run("Missing parameters", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/auth/callback", nil)
		rec := httptest.NewRecorder()

		controller.HandleCallback(rec, req)

		gt.Equal(t, rec.Code, http.StatusTemporaryRedirect)
		location := rec.Header().Get("Location")
		gt.True(t, strings.Contains(location, "error"))
	})

	t.Run("OAuth error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/auth/callback?error=access_denied", nil)
		rec := httptest.NewRecorder()

		controller.HandleCallback(rec, req)

		gt.Equal(t, rec.Code, http.StatusTemporaryRedirect)
		location := rec.Header().Get("Location")
		gt.True(t, strings.Contains(location, "error=auth_failed"))
	})
}

func TestAuthController_Me(t *testing.T) {
	ctx := context.Background()
	sessionRepo := memory.NewSessionRepository()

	userRepo := memory.NewUserRepository()
	mockSlackClient := createMockSlackClient()
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

	t.Run("No session", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
		rec := httptest.NewRecorder()

		controller.HandleMe(rec, req)

		gt.Equal(t, rec.Code, http.StatusUnauthorized)
	})

	t.Run("Valid session", func(t *testing.T) {
		// Create a session
		sessionID := types.NewUUID(ctx)
		session := &authmodel.Session{
			ID:        sessionID,
			UserID:    types.UserID("01234567-89ab-cdef-0123-456789abcdef"),
			UserName:  "Test User",
			Email:     "test@example.com",
			TeamID:    "T123456",
			TeamName:  "Test Team",
			ExpiresAt: time.Now().Add(24 * time.Hour),
			CreatedAt: time.Now(),
		}

		gt.NoError(t, sessionRepo.CreateSession(ctx, session))

		// Create request with session cookie
		req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
		req.AddCookie(&http.Cookie{
			Name:  "tamamo_session",
			Value: sessionID.String(),
		})
		rec := httptest.NewRecorder()

		controller.HandleMe(rec, req)

		gt.Equal(t, rec.Code, http.StatusOK)

		var response authctrl.UserResponse
		gt.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
		gt.Equal(t, response.ID, session.UserID.String())
		gt.Equal(t, response.Name, session.UserName)
		gt.Equal(t, response.Email, session.Email)
	})
}

func TestAuthController_Check(t *testing.T) {
	ctx := context.Background()
	sessionRepo := memory.NewSessionRepository()

	userRepo := memory.NewUserRepository()
	mockSlackClient := createMockSlackClient()
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

	t.Run("Not authenticated", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/auth/check", nil)
		rec := httptest.NewRecorder()

		controller.HandleCheck(rec, req)

		gt.Equal(t, rec.Code, http.StatusOK)

		var response authctrl.AuthCheckResponse
		gt.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
		gt.Equal(t, response.Authenticated, false)
	})

	t.Run("Authenticated", func(t *testing.T) {
		// Create a user first
		createdUser, err := userUseCase.GetOrCreateUser(ctx, "U123456789", "Test User", "test@example.com", "T123456")
		gt.NoError(t, err)

		// Create a session
		sessionID := types.NewUUID(ctx)
		session := &authmodel.Session{
			ID:        sessionID,
			UserID:    createdUser.ID,
			UserName:  "Test User",
			Email:     "test@example.com",
			TeamID:    "T123456",
			TeamName:  "Test Team",
			ExpiresAt: time.Now().Add(24 * time.Hour),
			CreatedAt: time.Now(),
		}

		gt.NoError(t, sessionRepo.CreateSession(ctx, session))

		// Create request with session cookie
		req := httptest.NewRequest(http.MethodGet, "/api/auth/check", nil)
		req.AddCookie(&http.Cookie{
			Name:  "tamamo_session",
			Value: sessionID.String(),
		})
		rec := httptest.NewRecorder()

		controller.HandleCheck(rec, req)

		gt.Equal(t, rec.Code, http.StatusOK)

		var response authctrl.AuthCheckResponse
		gt.NoError(t, json.NewDecoder(rec.Body).Decode(&response))
		gt.Equal(t, response.Authenticated, true)
		gt.Equal(t, response.User.ID, createdUser.ID.String())
	})
}

func TestAuthController_Logout(t *testing.T) {
	ctx := context.Background()
	sessionRepo := memory.NewSessionRepository()

	userRepo := memory.NewUserRepository()
	mockSlackClient := createMockSlackClient()
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

	// Create a session
	sessionID := types.NewUUID(ctx)
	session := &authmodel.Session{
		ID:        sessionID,
		UserID:    "U123456",
		UserName:  "Test User",
		Email:     "test@example.com",
		TeamID:    "T123456",
		TeamName:  "Test Team",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}

	gt.NoError(t, sessionRepo.CreateSession(ctx, session))

	// Create request with session cookie
	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	req.AddCookie(&http.Cookie{
		Name:  "tamamo_session",
		Value: sessionID.String(),
	})
	rec := httptest.NewRecorder()

	controller.HandleLogout(rec, req)

	gt.Equal(t, rec.Code, http.StatusOK)

	// Check that session is deleted
	_, err := sessionRepo.GetSession(ctx, sessionID.String())
	gt.Error(t, err)
	gt.Equal(t, err, authmodel.ErrSessionNotFound)

	// Check that cookie is cleared
	cookies := rec.Result().Cookies()
	gt.True(t, len(cookies) > 0)
	sessionCookie := cookies[0]
	gt.Equal(t, sessionCookie.Name, "tamamo_session")
	gt.Equal(t, sessionCookie.MaxAge, -1)
}
