package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/agent"
	"google.golang.org/api/iterator"
)

type notionSearchConfigRepository struct {
	client *firestore.Client
}

// NewNotionSearchConfigRepository creates a new Notion search config repository
func NewNotionSearchConfigRepository(client *firestore.Client) interfaces.NotionSearchConfigRepository {
	return &notionSearchConfigRepository{
		client: client,
	}
}

const notionSearchConfigCollection = "agent_notion_search_configs"

func (r *notionSearchConfigRepository) Create(ctx context.Context, config *agent.NotionSearchConfig) error {
	if err := config.Validate(); err != nil {
		return goerr.Wrap(err, "invalid config")
	}

	_, err := r.client.Collection(notionSearchConfigCollection).Doc(config.ID).Set(ctx, config)
	if err != nil {
		return goerr.Wrap(err, "failed to create notion search config", goerr.V("id", config.ID))
	}

	return nil
}

func (r *notionSearchConfigRepository) GetByAgentID(ctx context.Context, agentID string) ([]*agent.NotionSearchConfig, error) {
	iter := r.client.Collection(notionSearchConfigCollection).
		Where("AgentID", "==", agentID).
		OrderBy("CreatedAt", firestore.Desc).
		Documents(ctx)

	var configs []*agent.NotionSearchConfig
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, goerr.Wrap(err, "failed to iterate documents")
		}

		var config agent.NotionSearchConfig
		if err := doc.DataTo(&config); err != nil {
			return nil, goerr.Wrap(err, "failed to unmarshal config")
		}
		configs = append(configs, &config)
	}

	return configs, nil
}

func (r *notionSearchConfigRepository) GetByID(ctx context.Context, id string) (*agent.NotionSearchConfig, error) {
	doc, err := r.client.Collection(notionSearchConfigCollection).Doc(id).Get(ctx)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get notion search config", goerr.V("id", id))
	}

	var config agent.NotionSearchConfig
	if err := doc.DataTo(&config); err != nil {
		return nil, goerr.Wrap(err, "failed to unmarshal config")
	}

	return &config, nil
}

func (r *notionSearchConfigRepository) Update(ctx context.Context, config *agent.NotionSearchConfig) error {
	if err := config.Validate(); err != nil {
		return goerr.Wrap(err, "invalid config")
	}

	_, err := r.client.Collection(notionSearchConfigCollection).Doc(config.ID).Set(ctx, config)
	if err != nil {
		return goerr.Wrap(err, "failed to update notion search config", goerr.V("id", config.ID))
	}

	return nil
}

func (r *notionSearchConfigRepository) Delete(ctx context.Context, id string) error {
	_, err := r.client.Collection(notionSearchConfigCollection).Doc(id).Delete(ctx)
	if err != nil {
		return goerr.Wrap(err, "failed to delete notion search config", goerr.V("id", id))
	}

	return nil
}

func (r *notionSearchConfigRepository) GetEnabledByAgentID(ctx context.Context, agentID string) ([]*agent.NotionSearchConfig, error) {
	iter := r.client.Collection(notionSearchConfigCollection).
		Where("AgentID", "==", agentID).
		Where("Enabled", "==", true).
		OrderBy("CreatedAt", firestore.Desc).
		Documents(ctx)

	var configs []*agent.NotionSearchConfig
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, goerr.Wrap(err, "failed to iterate documents")
		}

		var config agent.NotionSearchConfig
		if err := doc.DataTo(&config); err != nil {
			return nil, goerr.Wrap(err, "failed to unmarshal config")
		}
		configs = append(configs, &config)
	}

	return configs, nil
}

func (r *notionSearchConfigRepository) ExistsByAgentIDAndDatabaseID(ctx context.Context, agentID, databaseID string) (bool, error) {
	iter := r.client.Collection(notionSearchConfigCollection).
		Where("AgentID", "==", agentID).
		Where("DatabaseID", "==", databaseID).
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
