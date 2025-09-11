package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/agent"
	"google.golang.org/api/iterator"
)

type jiraSearchConfigRepository struct {
	client *firestore.Client
}

// NewJiraSearchConfigRepository creates a new Jira search config repository
func NewJiraSearchConfigRepository(client *firestore.Client) interfaces.JiraSearchConfigRepository {
	return &jiraSearchConfigRepository{
		client: client,
	}
}

const jiraSearchConfigCollection = "agent_jira_search_configs"

func (r *jiraSearchConfigRepository) Create(ctx context.Context, config *agent.JiraSearchConfig) error {
	if err := config.Validate(); err != nil {
		return goerr.Wrap(err, "invalid config")
	}

	_, err := r.client.Collection(jiraSearchConfigCollection).Doc(config.ID).Set(ctx, config)
	if err != nil {
		return goerr.Wrap(err, "failed to create jira search config", goerr.V("id", config.ID))
	}

	return nil
}

func (r *jiraSearchConfigRepository) GetByAgentID(ctx context.Context, agentID string) ([]*agent.JiraSearchConfig, error) {
	iter := r.client.Collection(jiraSearchConfigCollection).
		Where("AgentID", "==", agentID).
		OrderBy("CreatedAt", firestore.Desc).
		Documents(ctx)

	var configs []*agent.JiraSearchConfig
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, goerr.Wrap(err, "failed to iterate documents")
		}

		var config agent.JiraSearchConfig
		if err := doc.DataTo(&config); err != nil {
			return nil, goerr.Wrap(err, "failed to unmarshal config")
		}
		configs = append(configs, &config)
	}

	return configs, nil
}

func (r *jiraSearchConfigRepository) GetByID(ctx context.Context, id string) (*agent.JiraSearchConfig, error) {
	doc, err := r.client.Collection(jiraSearchConfigCollection).Doc(id).Get(ctx)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get jira search config", goerr.V("id", id))
	}

	var config agent.JiraSearchConfig
	if err := doc.DataTo(&config); err != nil {
		return nil, goerr.Wrap(err, "failed to unmarshal config")
	}

	return &config, nil
}

func (r *jiraSearchConfigRepository) Update(ctx context.Context, config *agent.JiraSearchConfig) error {
	if err := config.Validate(); err != nil {
		return goerr.Wrap(err, "invalid config")
	}

	_, err := r.client.Collection(jiraSearchConfigCollection).Doc(config.ID).Set(ctx, config)
	if err != nil {
		return goerr.Wrap(err, "failed to update jira search config", goerr.V("id", config.ID))
	}

	return nil
}

func (r *jiraSearchConfigRepository) Delete(ctx context.Context, id string) error {
	_, err := r.client.Collection(jiraSearchConfigCollection).Doc(id).Delete(ctx)
	if err != nil {
		return goerr.Wrap(err, "failed to delete jira search config", goerr.V("id", id))
	}

	return nil
}

func (r *jiraSearchConfigRepository) GetEnabledByAgentID(ctx context.Context, agentID string) ([]*agent.JiraSearchConfig, error) {
	iter := r.client.Collection(jiraSearchConfigCollection).
		Where("AgentID", "==", agentID).
		Where("Enabled", "==", true).
		OrderBy("CreatedAt", firestore.Desc).
		Documents(ctx)

	var configs []*agent.JiraSearchConfig
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, goerr.Wrap(err, "failed to iterate documents")
		}

		var config agent.JiraSearchConfig
		if err := doc.DataTo(&config); err != nil {
			return nil, goerr.Wrap(err, "failed to unmarshal config")
		}
		configs = append(configs, &config)
	}

	return configs, nil
}

func (r *jiraSearchConfigRepository) ExistsByAgentIDAndProjectKey(ctx context.Context, agentID, projectKey string) (bool, error) {
	iter := r.client.Collection(jiraSearchConfigCollection).
		Where("AgentID", "==", agentID).
		Where("ProjectKey", "==", projectKey).
		Limit(1).
		Documents(ctx)

	_, err := iter.Next()
	if err == iterator.Done {
		return false, nil
	}
	if err != nil {
		return false, goerr.Wrap(err, "failed to check existence")
	}

	return true, nil
}
