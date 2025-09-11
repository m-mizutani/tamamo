package graphql

import (
	"context"
	"net/http"

	"github.com/m-mizutani/tamamo/pkg/controller/auth"
	graphql1 "github.com/m-mizutani/tamamo/pkg/domain/model/graphql"
	"github.com/m-mizutani/tamamo/pkg/domain/model/integration"
	"github.com/m-mizutani/tamamo/pkg/domain/model/user"
)

type contextKey string

const (
	// httpResponseWriterKey is the context key for storing the HTTP response writer
	httpResponseWriterKey contextKey = "http_response_writer"
)

func getCurrentUser(ctx context.Context) *user.User {
	session, ok := auth.UserFromContext(ctx)
	if !ok {
		return nil
	}
	// For simplicity, create a user from session info
	// In a real implementation, you might want to fetch full user details
	return &user.User{
		ID: session.UserID,
	}
}

func getResponseWriter(ctx context.Context) http.ResponseWriter {
	// Extract response writer from context
	// This requires middleware that sets the response writer in context
	if w := ctx.Value(httpResponseWriterKey); w != nil {
		if responseWriter, ok := w.(http.ResponseWriter); ok {
			return responseWriter
		}
	}
	return nil
}

func WithResponseWriter(ctx context.Context, w http.ResponseWriter) context.Context {
	return context.WithValue(ctx, httpResponseWriterKey, w)
}

func convertJiraIntegrationToGraphQL(integration *integration.JiraIntegration) *graphql1.JiraIntegration {
	if integration == nil {
		// Return disconnected state when no integration exists
		return &graphql1.JiraIntegration{
			ID:          "jira",
			Connected:   false,
			SiteURL:     nil,
			ConnectedAt: nil,
		}
	}

	// Return connected state with details
	return &graphql1.JiraIntegration{
		ID:          "jira",
		Connected:   true,
		SiteURL:     &integration.SiteURL,
		ConnectedAt: &integration.CreatedAt,
	}
}

func convertNotionIntegrationToGraphQL(integration *integration.NotionIntegration) *graphql1.NotionIntegration {
	if integration == nil {
		// Return disconnected state when no integration exists
		return &graphql1.NotionIntegration{
			ID:            "notion",
			Connected:     false,
			WorkspaceName: nil,
			WorkspaceIcon: nil,
			ConnectedAt:   nil,
		}
	}

	// Return connected state with details
	return &graphql1.NotionIntegration{
		ID:            "notion",
		Connected:     true,
		WorkspaceName: &integration.WorkspaceName,
		WorkspaceIcon: &integration.WorkspaceIcon,
		ConnectedAt:   &integration.CreatedAt,
	}
}
