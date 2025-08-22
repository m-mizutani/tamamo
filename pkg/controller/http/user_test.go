package http_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/m-mizutani/gt"
	httpctrl "github.com/m-mizutani/tamamo/pkg/controller/http"
	"github.com/m-mizutani/tamamo/pkg/domain/mock"
	"github.com/m-mizutani/tamamo/pkg/domain/model/user"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
)

func TestUserController_HandleGetUserAvatar(t *testing.T) {
	testUserID := types.UserID("01234567-89ab-cdef-0123-456789abcdef")

	// Setup mock use case
	mockUserUseCase := &mock.UserUseCasesMock{
		GetUserAvatarFunc: func(ctx context.Context, userID types.UserID, size int) ([]byte, error) {
			if userID == testUserID {
				return []byte("fake-avatar-data"), nil
			}
			return nil, errors.New("user not found")
		},
	}

	// Create controller
	userCtrl := httpctrl.NewUserController(mockUserUseCase)

	// Create router
	r := chi.NewRouter()
	r.Get("/api/users/{userID}/avatar", userCtrl.HandleGetUserAvatar)

	t.Run("ValidUserID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/users/"+string(testUserID)+"/avatar", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		gt.Equal(t, w.Code, http.StatusOK)
		gt.Equal(t, w.Header().Get("Content-Type"), "image/jpeg")
		gt.Equal(t, w.Body.String(), "fake-avatar-data")

		// Check that mock was called
		calls := mockUserUseCase.GetUserAvatarCalls()
		gt.Equal(t, len(calls), 1)
		gt.Equal(t, calls[0].UserID, testUserID)
		gt.Equal(t, calls[0].Size, 48) // default size
	})

	t.Run("CustomSize", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/users/"+string(testUserID)+"/avatar?size=72", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		gt.Equal(t, w.Code, http.StatusOK)

		// Check that custom size was passed
		calls := mockUserUseCase.GetUserAvatarCalls()
		gt.Equal(t, len(calls), 2) // Previous call + this call
		gt.Equal(t, calls[1].Size, 72)
	})

	t.Run("InvalidUserID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/users/invalid-id/avatar", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		gt.Equal(t, w.Code, http.StatusBadRequest)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/users/00000000-0000-0000-0000-000000000000/avatar", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		gt.Equal(t, w.Code, http.StatusNotFound)
	})
}

func TestUserController_HandleGetUserInfo(t *testing.T) {
	testUserID := types.UserID("01234567-89ab-cdef-0123-456789abcdef")

	// Setup mock use case
	testUser := &user.User{
		ID:        testUserID,
		SlackName: "Test User",
	}

	mockUserUseCase := &mock.UserUseCasesMock{
		GetUserByIDFunc: func(ctx context.Context, userID types.UserID) (*user.User, error) {
			if userID == testUserID {
				return testUser, nil
			}
			return nil, errors.New("user not found")
		},
	}

	// Create controller
	userCtrl := httpctrl.NewUserController(mockUserUseCase)

	// Create router
	r := chi.NewRouter()
	r.Get("/api/users/{userID}", userCtrl.HandleGetUserInfo)

	t.Run("ValidUserID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/users/"+string(testUserID), nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		gt.Equal(t, w.Code, http.StatusOK)
		gt.Equal(t, w.Header().Get("Content-Type"), "application/json")

		// Check that mock was called
		calls := mockUserUseCase.GetUserByIDCalls()
		gt.Equal(t, len(calls), 1)
		gt.Equal(t, calls[0].UserID, testUserID)
	})

	t.Run("InvalidUserID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/users/invalid-id", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		gt.Equal(t, w.Code, http.StatusBadRequest)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/users/00000000-0000-0000-0000-000000000000", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		gt.Equal(t, w.Code, http.StatusNotFound)
	})
}
