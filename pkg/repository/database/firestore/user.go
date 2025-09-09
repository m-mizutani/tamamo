package firestore

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/integration"
	"github.com/m-mizutani/tamamo/pkg/domain/model/user"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// userDoc represents the Firestore document structure for users
type userDoc struct {
	ID          string    `firestore:"id"`
	SlackID     string    `firestore:"slack_id"`
	SlackName   string    `firestore:"slack_name"`
	DisplayName string    `firestore:"display_name"`
	Email       string    `firestore:"email"`
	TeamID      string    `firestore:"team_id"`
	CreatedAt   time.Time `firestore:"created_at"`
	UpdatedAt   time.Time `firestore:"updated_at"`
}

// toUserDoc converts a User entity to a Firestore document
func toUserDoc(u *user.User) *userDoc {
	return &userDoc{
		ID:          u.ID.String(),
		SlackID:     u.SlackID,
		SlackName:   u.SlackName,
		DisplayName: u.DisplayName,
		Email:       u.Email,
		TeamID:      u.TeamID,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

// toUser converts a Firestore document to a User entity
func (d *userDoc) toUser() *user.User {
	return &user.User{
		ID:          types.UserID(d.ID),
		SlackID:     d.SlackID,
		SlackName:   d.SlackName,
		DisplayName: d.DisplayName,
		Email:       d.Email,
		TeamID:      d.TeamID,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

// GetByID retrieves a user by their UUID
func (c *Client) GetByID(ctx context.Context, id types.UserID) (*user.User, error) {
	doc, err := c.client.Collection("users").Doc(id.String()).Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, goerr.Wrap(err, "user not found", goerr.V("user_id", id))
		}
		return nil, goerr.Wrap(err, "failed to get user", goerr.V("user_id", id))
	}

	var userDoc userDoc
	if err := doc.DataTo(&userDoc); err != nil {
		return nil, goerr.Wrap(err, "failed to parse user document", goerr.V("user_id", id))
	}

	return userDoc.toUser(), nil
}

// GetBySlackIDAndTeamID retrieves a user by their Slack ID and Team ID
func (c *Client) GetBySlackIDAndTeamID(ctx context.Context, slackID, teamID string) (*user.User, error) {
	iter := c.client.Collection("users").
		Where("slack_id", "==", slackID).
		Where("team_id", "==", teamID).
		Limit(1).
		Documents(ctx)

	doc, err := iter.Next()
	if err != nil {
		if err == iterator.Done {
			return nil, goerr.New("user not found", goerr.V("slack_id", slackID), goerr.V("team_id", teamID))
		}
		return nil, goerr.Wrap(err, "failed to query user", goerr.V("slack_id", slackID), goerr.V("team_id", teamID))
	}

	var userDoc userDoc
	if err := doc.DataTo(&userDoc); err != nil {
		return nil, goerr.Wrap(err, "failed to parse user document", goerr.V("slack_id", slackID), goerr.V("team_id", teamID))
	}

	return userDoc.toUser(), nil
}

// Create creates a new user
func (c *Client) Create(ctx context.Context, u *user.User) error {
	if err := u.Validate(); err != nil {
		return goerr.Wrap(err, "invalid user")
	}

	doc := toUserDoc(u)
	_, err := c.client.Collection("users").Doc(u.ID.String()).Set(ctx, doc)
	if err != nil {
		return goerr.Wrap(err, "failed to create user", goerr.V("user_id", u.ID))
	}

	return nil
}

// Update updates an existing user
func (c *Client) Update(ctx context.Context, u *user.User) error {
	if err := u.Validate(); err != nil {
		return goerr.Wrap(err, "invalid user")
	}

	doc := toUserDoc(u)
	_, err := c.client.Collection("users").Doc(u.ID.String()).Set(ctx, doc)
	if err != nil {
		return goerr.Wrap(err, "failed to update user", goerr.V("user_id", u.ID))
	}

	return nil
}

// jiraIntegrationDoc represents the Firestore document structure for Jira integrations
type jiraIntegrationDoc struct {
	UserID         string    `firestore:"user_id"`
	CloudID        string    `firestore:"cloud_id"`
	SiteURL        string    `firestore:"site_url"`
	AccessToken    string    `firestore:"access_token"`
	RefreshToken   string    `firestore:"refresh_token"`
	TokenExpiresAt time.Time `firestore:"token_expires_at"`
	Scopes         []string  `firestore:"scopes"`
	CreatedAt      time.Time `firestore:"created_at"`
	UpdatedAt      time.Time `firestore:"updated_at"`
}

// toJiraIntegrationDoc converts a JiraIntegration entity to a Firestore document
func toJiraIntegrationDoc(j *integration.JiraIntegration) *jiraIntegrationDoc {
	return &jiraIntegrationDoc{
		UserID:         j.UserID,
		CloudID:        j.CloudID,
		SiteURL:        j.SiteURL,
		AccessToken:    j.AccessToken,
		RefreshToken:   j.RefreshToken,
		TokenExpiresAt: j.TokenExpiresAt,
		Scopes:         j.Scopes,
		CreatedAt:      j.CreatedAt,
		UpdatedAt:      j.UpdatedAt,
	}
}

// toJiraIntegration converts a Firestore document to a JiraIntegration entity
func (d *jiraIntegrationDoc) toJiraIntegration() *integration.JiraIntegration {
	return &integration.JiraIntegration{
		UserID:         d.UserID,
		CloudID:        d.CloudID,
		SiteURL:        d.SiteURL,
		AccessToken:    d.AccessToken,
		RefreshToken:   d.RefreshToken,
		TokenExpiresAt: d.TokenExpiresAt,
		Scopes:         d.Scopes,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
	}
}

// SaveJiraIntegration saves a Jira integration to Firestore
func (c *Client) SaveJiraIntegration(ctx context.Context, integration *integration.JiraIntegration) error {
	doc := toJiraIntegrationDoc(integration)
	_, err := c.client.Collection("users").Doc(integration.UserID).Collection("integrations").Doc("jira").Set(ctx, doc)
	if err != nil {
		return goerr.Wrap(err, "failed to save Jira integration", goerr.V("user_id", integration.UserID))
	}
	return nil
}

// GetJiraIntegration retrieves a Jira integration from Firestore
func (c *Client) GetJiraIntegration(ctx context.Context, userID string) (*integration.JiraIntegration, error) {
	doc, err := c.client.Collection("users").Doc(userID).Collection("integrations").Doc("jira").Get(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil, nil // Return nil when not found (not connected)
		}
		return nil, goerr.Wrap(err, "failed to get Jira integration", goerr.V("user_id", userID))
	}

	var jiraDoc jiraIntegrationDoc
	if err := doc.DataTo(&jiraDoc); err != nil {
		return nil, goerr.Wrap(err, "failed to parse Jira integration document", goerr.V("user_id", userID))
	}

	return jiraDoc.toJiraIntegration(), nil
}

// DeleteJiraIntegration deletes a Jira integration from Firestore
func (c *Client) DeleteJiraIntegration(ctx context.Context, userID string) error {
	_, err := c.client.Collection("users").Doc(userID).Collection("integrations").Doc("jira").Delete(ctx)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return nil // Already deleted
		}
		return goerr.Wrap(err, "failed to delete Jira integration", goerr.V("user_id", userID))
	}
	return nil
}

// NewUserRepository creates a new user repository using Firestore client
func NewUserRepository(client *firestore.Client) interfaces.UserRepository {
	return &Client{
		client:     client,
		projectID:  "", // These will be filled from the existing client
		databaseID: "",
	}
}
