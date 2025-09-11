package graphql

import (
	"github.com/m-mizutani/tamamo/pkg/domain/model/agent"
	graphql1 "github.com/m-mizutani/tamamo/pkg/domain/model/graphql"
)

// convertSlackSearchConfigToGraphQL converts a domain SlackSearchConfig to GraphQL model
func convertSlackSearchConfigToGraphQL(config *agent.SlackSearchConfig) *graphql1.AgentSlackSearchConfig {
	if config == nil {
		return nil
	}

	return &graphql1.AgentSlackSearchConfig{
		ID:          config.ID,
		AgentID:     config.AgentID,
		ChannelID:   config.ChannelID,
		ChannelName: config.ChannelName,
		Description: config.Description,
		Enabled:     config.Enabled,
		CreatedAt:   config.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   config.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// convertJiraSearchConfigToGraphQL converts a domain JiraSearchConfig to GraphQL model
func convertJiraSearchConfigToGraphQL(config *agent.JiraSearchConfig) *graphql1.AgentJiraSearchConfig {
	if config == nil {
		return nil
	}

	return &graphql1.AgentJiraSearchConfig{
		ID:          config.ID,
		AgentID:     config.AgentID,
		ProjectKey:  config.ProjectKey,
		ProjectName: config.ProjectName,
		BoardID:     config.BoardID,
		BoardName:   config.BoardName,
		Description: config.Description,
		Enabled:     config.Enabled,
		CreatedAt:   config.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   config.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// convertNotionSearchConfigToGraphQL converts a domain NotionSearchConfig to GraphQL model
func convertNotionSearchConfigToGraphQL(config *agent.NotionSearchConfig) *graphql1.AgentNotionSearchConfig {
	if config == nil {
		return nil
	}

	return &graphql1.AgentNotionSearchConfig{
		ID:           config.ID,
		AgentID:      config.AgentID,
		DatabaseID:   config.DatabaseID,
		DatabaseName: config.DatabaseName,
		WorkspaceID:  config.WorkspaceID,
		Description:  config.Description,
		Enabled:      config.Enabled,
		CreatedAt:    config.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:    config.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// convertSlackSearchConfigsToGraphQL converts a slice of domain SlackSearchConfigs to GraphQL models
func convertSlackSearchConfigsToGraphQL(configs []*agent.SlackSearchConfig) []*graphql1.AgentSlackSearchConfig {
	result := make([]*graphql1.AgentSlackSearchConfig, len(configs))
	for i, config := range configs {
		result[i] = convertSlackSearchConfigToGraphQL(config)
	}
	return result
}

// convertJiraSearchConfigsToGraphQL converts a slice of domain JiraSearchConfigs to GraphQL models
func convertJiraSearchConfigsToGraphQL(configs []*agent.JiraSearchConfig) []*graphql1.AgentJiraSearchConfig {
	result := make([]*graphql1.AgentJiraSearchConfig, len(configs))
	for i, config := range configs {
		result[i] = convertJiraSearchConfigToGraphQL(config)
	}
	return result
}

// convertNotionSearchConfigsToGraphQL converts a slice of domain NotionSearchConfigs to GraphQL models
func convertNotionSearchConfigsToGraphQL(configs []*agent.NotionSearchConfig) []*graphql1.AgentNotionSearchConfig {
	result := make([]*graphql1.AgentNotionSearchConfig, len(configs))
	for i, config := range configs {
		result[i] = convertNotionSearchConfigToGraphQL(config)
	}
	return result
}
