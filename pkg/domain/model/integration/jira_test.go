package integration_test

import (
	"testing"
	"time"

	"github.com/m-mizutani/gt"
	"github.com/m-mizutani/tamamo/pkg/domain/model/integration"
)

func TestJiraIntegration(t *testing.T) {
	t.Run("NewJiraIntegration creates instance with user ID", func(t *testing.T) {
		userID := "user123"
		jira := integration.NewJiraIntegration(userID)

		gt.V(t, jira.UserID).Equal(userID)
		gt.V(t, jira.CreatedAt.IsZero()).Equal(false)
		gt.V(t, jira.UpdatedAt.IsZero()).Equal(false)
	})

	t.Run("UpdateTokens updates token information", func(t *testing.T) {
		jira := integration.NewJiraIntegration("user123")
		accessToken := "access_token"
		refreshToken := "refresh_token"
		expiresAt := time.Now().Add(1 * time.Hour)

		jira.UpdateTokens(accessToken, refreshToken, expiresAt)

		gt.V(t, jira.AccessToken).Equal(accessToken)
		gt.V(t, jira.RefreshToken).Equal(refreshToken)
		gt.V(t, jira.TokenExpiresAt).Equal(expiresAt)
	})

	t.Run("UpdateSiteInfo updates Jira site information", func(t *testing.T) {
		jira := integration.NewJiraIntegration("user123")
		cloudID := "cloud123"
		siteURL := "example.atlassian.net"

		jira.UpdateSiteInfo(cloudID, siteURL)

		gt.V(t, jira.CloudID).Equal(cloudID)
		gt.V(t, jira.SiteURL).Equal(siteURL)
	})

	t.Run("IsTokenExpired checks token expiration", func(t *testing.T) {
		jira := integration.NewJiraIntegration("user123")

		// Set expired token
		jira.TokenExpiresAt = time.Now().Add(-1 * time.Hour)
		gt.V(t, jira.IsTokenExpired()).Equal(true)

		// Set valid token
		jira.TokenExpiresAt = time.Now().Add(1 * time.Hour)
		gt.V(t, jira.IsTokenExpired()).Equal(false)
	})

	t.Run("IsConnected checks connection status", func(t *testing.T) {
		jira := integration.NewJiraIntegration("user123")

		// Not connected initially
		gt.V(t, jira.IsConnected()).Equal(false)

		// Connected with valid token
		jira.AccessToken = "token"
		jira.TokenExpiresAt = time.Now().Add(1 * time.Hour)
		gt.V(t, jira.IsConnected()).Equal(true)

		// Not connected with expired token
		jira.TokenExpiresAt = time.Now().Add(-1 * time.Hour)
		gt.V(t, jira.IsConnected()).Equal(false)
	})
}
