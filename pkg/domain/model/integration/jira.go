package integration

import (
	"time"
)

type JiraIntegration struct {
	UserID         string
	CloudID        string
	SiteURL        string
	AccessToken    string
	RefreshToken   string
	TokenExpiresAt time.Time
	Scopes         []string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewJiraIntegration(userID string) *JiraIntegration {
	now := time.Now()
	return &JiraIntegration{
		UserID:    userID,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (j *JiraIntegration) UpdateTokens(accessToken, refreshToken string, expiresAt time.Time) {
	j.AccessToken = accessToken
	j.RefreshToken = refreshToken
	j.TokenExpiresAt = expiresAt
	j.UpdatedAt = time.Now()
}

func (j *JiraIntegration) UpdateSiteInfo(cloudID, siteURL string) {
	j.CloudID = cloudID
	j.SiteURL = siteURL
	j.UpdatedAt = time.Now()
}

func (j *JiraIntegration) IsTokenExpired() bool {
	return time.Now().After(j.TokenExpiresAt)
}

func (j *JiraIntegration) IsConnected() bool {
	return j.AccessToken != "" && !j.IsTokenExpired()
}
