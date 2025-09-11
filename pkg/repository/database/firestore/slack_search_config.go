package firestore

import (
	"context"

	"cloud.google.com/go/firestore"
	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/agent"
	"google.golang.org/api/iterator"
)

type slackSearchConfigRepository struct {
	client *firestore.Client
}

// NewSlackSearchConfigRepository creates a new Slack search config repository
func NewSlackSearchConfigRepository(client *firestore.Client) interfaces.SlackSearchConfigRepository {
	return &slackSearchConfigRepository{
		client: client,
	}
}

const slackSearchConfigCollection = "agent_slack_search_configs"

func (r *slackSearchConfigRepository) Create(ctx context.Context, config *agent.SlackSearchConfig) error {
	if err := config.Validate(); err != nil {
		return goerr.Wrap(err, "invalid config")
	}

	_, err := r.client.Collection(slackSearchConfigCollection).Doc(config.ID).Set(ctx, config)
	if err != nil {
		return goerr.Wrap(err, "failed to create slack search config", goerr.V("id", config.ID))
	}

	return nil
}

func (r *slackSearchConfigRepository) GetByAgentID(ctx context.Context, agentID string) ([]*agent.SlackSearchConfig, error) {
	iter := r.client.Collection(slackSearchConfigCollection).
		Where("AgentID", "==", agentID).
		OrderBy("CreatedAt", firestore.Desc).
		Documents(ctx)

	var configs []*agent.SlackSearchConfig
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, goerr.Wrap(err, "failed to iterate documents")
		}

		var config agent.SlackSearchConfig
		if err := doc.DataTo(&config); err != nil {
			return nil, goerr.Wrap(err, "failed to unmarshal config")
		}
		configs = append(configs, &config)
	}

	return configs, nil
}

func (r *slackSearchConfigRepository) GetByID(ctx context.Context, id string) (*agent.SlackSearchConfig, error) {
	doc, err := r.client.Collection(slackSearchConfigCollection).Doc(id).Get(ctx)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get slack search config", goerr.V("id", id))
	}

	var config agent.SlackSearchConfig
	if err := doc.DataTo(&config); err != nil {
		return nil, goerr.Wrap(err, "failed to unmarshal config")
	}

	return &config, nil
}

func (r *slackSearchConfigRepository) Update(ctx context.Context, config *agent.SlackSearchConfig) error {
	if err := config.Validate(); err != nil {
		return goerr.Wrap(err, "invalid config")
	}

	_, err := r.client.Collection(slackSearchConfigCollection).Doc(config.ID).Set(ctx, config)
	if err != nil {
		return goerr.Wrap(err, "failed to update slack search config", goerr.V("id", config.ID))
	}

	return nil
}

func (r *slackSearchConfigRepository) Delete(ctx context.Context, id string) error {
	_, err := r.client.Collection(slackSearchConfigCollection).Doc(id).Delete(ctx)
	if err != nil {
		return goerr.Wrap(err, "failed to delete slack search config", goerr.V("id", id))
	}

	return nil
}

func (r *slackSearchConfigRepository) GetEnabledByAgentID(ctx context.Context, agentID string) ([]*agent.SlackSearchConfig, error) {
	iter := r.client.Collection(slackSearchConfigCollection).
		Where("AgentID", "==", agentID).
		Where("Enabled", "==", true).
		OrderBy("CreatedAt", firestore.Desc).
		Documents(ctx)

	var configs []*agent.SlackSearchConfig
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, goerr.Wrap(err, "failed to iterate documents")
		}

		var config agent.SlackSearchConfig
		if err := doc.DataTo(&config); err != nil {
			return nil, goerr.Wrap(err, "failed to unmarshal config")
		}
		configs = append(configs, &config)
	}

	return configs, nil
}

func (r *slackSearchConfigRepository) ExistsByAgentIDAndChannelID(ctx context.Context, agentID, channelID string) (bool, error) {
	iter := r.client.Collection(slackSearchConfigCollection).
		Where("AgentID", "==", agentID).
		Where("ChannelID", "==", channelID).
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
