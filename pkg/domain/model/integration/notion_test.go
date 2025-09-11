package integration_test

import (
	"testing"
	"time"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/domain/model/integration"
)

func TestNewNotionIntegration(t *testing.T) {
	userID := "test-user-123"
	notion := integration.NewNotionIntegration(userID)

	gt.V(t, notion.UserID).Equal(userID)
	gt.V(t, notion.WorkspaceID).Equal("")
	gt.V(t, notion.WorkspaceName).Equal("")
	gt.V(t, notion.WorkspaceIcon).Equal("")
	gt.V(t, notion.BotID).Equal("")
	gt.V(t, notion.AccessToken).Equal("")
	gt.V(t, notion.CreatedAt).NotEqual(time.Time{})
	gt.V(t, notion.UpdatedAt).NotEqual(time.Time{})
	gt.V(t, notion.CreatedAt).Equal(notion.UpdatedAt)
}

func TestNotionIntegration_UpdateTokens(t *testing.T) {
	notion := integration.NewNotionIntegration("test-user")
	originalUpdatedAt := notion.UpdatedAt

	// Sleep to ensure time difference
	time.Sleep(time.Millisecond)

	accessToken := "test-access-token"
	notion.UpdateTokens(accessToken)

	gt.V(t, notion.AccessToken).Equal(accessToken)
	gt.B(t, notion.UpdatedAt.After(originalUpdatedAt)).True()
}

func TestNotionIntegration_UpdateWorkspaceInfo(t *testing.T) {
	notion := integration.NewNotionIntegration("test-user")
	originalUpdatedAt := notion.UpdatedAt

	// Sleep to ensure time difference
	time.Sleep(time.Millisecond)

	workspaceID := "workspace-123"
	workspaceName := "Test Workspace"
	workspaceIcon := "https://example.com/icon.png"
	botID := "bot-456"

	notion.UpdateWorkspaceInfo(workspaceID, workspaceName, workspaceIcon, botID)

	gt.V(t, notion.WorkspaceID).Equal(workspaceID)
	gt.V(t, notion.WorkspaceName).Equal(workspaceName)
	gt.V(t, notion.WorkspaceIcon).Equal(workspaceIcon)
	gt.V(t, notion.BotID).Equal(botID)
	gt.B(t, notion.UpdatedAt.After(originalUpdatedAt)).True()
}

func TestNotionIntegration_IsConnected(t *testing.T) {
	t.Run("not connected when no token", func(t *testing.T) {
		notion := integration.NewNotionIntegration("test-user")
		gt.B(t, notion.IsConnected()).False()
	})

	t.Run("connected when token exists", func(t *testing.T) {
		notion := integration.NewNotionIntegration("test-user")
		notion.UpdateTokens("test-token")
		gt.B(t, notion.IsConnected()).True()
	})

	t.Run("always connected with valid token (no expiration)", func(t *testing.T) {
		notion := integration.NewNotionIntegration("test-user")
		notion.UpdateTokens("test-token")

		// Unlike Jira, Notion tokens don't expire
		// So even after time passes, it should still be connected
		gt.B(t, notion.IsConnected()).True()
	})
}
