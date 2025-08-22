package user_test

import (
	"testing"
	"time"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/domain/model/user"
)

func TestNewUser(t *testing.T) {
	u := user.NewUser("U123456789", "Test User", "Test Display", "test@example.com", "T123456789")

	gt.NotEqual(t, u, nil)
	gt.Equal(t, u.SlackID, "U123456789")
	gt.Equal(t, u.SlackName, "Test User")
	gt.Equal(t, u.Email, "test@example.com")
	gt.Equal(t, u.TeamID, "T123456789")
	gt.True(t, u.ID.IsValid())
	gt.True(t, !u.CreatedAt.IsZero())
	gt.True(t, !u.UpdatedAt.IsZero())
}

func TestUser_UpdateSlackInfo(t *testing.T) {
	u := user.NewUser("U123456789", "Old Name", "Old Display", "old@example.com", "T123456789")
	originalUpdatedAt := u.UpdatedAt

	// Wait a bit to ensure time difference
	time.Sleep(1 * time.Millisecond)

	u.UpdateSlackInfo("New Name", "New Display", "new@example.com")

	gt.Equal(t, u.SlackName, "New Name")
	gt.Equal(t, u.DisplayName, "New Display")
	gt.Equal(t, u.Email, "new@example.com")
	gt.True(t, u.UpdatedAt.After(originalUpdatedAt))

	// SlackID and TeamID should not change
	gt.Equal(t, u.SlackID, "U123456789")
	gt.Equal(t, u.TeamID, "T123456789")
}

func TestUser_IsUpdateRequired(t *testing.T) {
	u := user.NewUser("U123456789", "Test User", "Test Display", "test@example.com", "T123456789")

	// Should not require update immediately
	gt.False(t, u.IsUpdateRequired(1*time.Hour))

	// Simulate old update time
	u.UpdatedAt = time.Now().Add(-25 * time.Hour)
	gt.True(t, u.IsUpdateRequired(24*time.Hour))

	// Should not require update with longer interval
	gt.False(t, u.IsUpdateRequired(48*time.Hour))
}

func TestUser_Validate(t *testing.T) {
	t.Run("valid user", func(t *testing.T) {
		u := user.NewUser("U123456789", "Test User", "Test Display", "test@example.com", "T123456789")
		gt.NoError(t, u.Validate())
	})

	t.Run("empty slack ID", func(t *testing.T) {
		u := user.NewUser("", "Test User", "Test Display", "test@example.com", "T123456789")
		gt.Error(t, u.Validate())
	})

	t.Run("empty team ID", func(t *testing.T) {
		u := user.NewUser("U123456789", "Test User", "Test Display", "test@example.com", "")
		gt.Error(t, u.Validate())
	})
}
