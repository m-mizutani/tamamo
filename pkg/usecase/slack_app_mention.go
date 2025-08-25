package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/m-mizutani/ctxlog"
	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/gollem"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
	pkgErrors "github.com/m-mizutani/tamamo/pkg/utils/errors"
)

// generalModeUUID is a special UUID for general mode threads
var generalModeUUID = types.UUID("00000000-0000-0000-0000-000000000000")

// threadContext represents the context of the thread (internal use only)
type threadContext struct {
	isNewThread    bool          // Is this a new thread
	existingThread *slack.Thread // Existing thread (if any)
	requiresAgent  bool          // Does this require agent specification
}

// agentContext represents resolved agent information (internal use only)
type agentContext struct {
	uuid         types.UUID // Agent UUID (special UUID for general mode)
	version      string     // Agent version
	systemPrompt string     // System prompt
	llmProvider  string     // LLM provider (e.g., "gemini", "claude", "openai")
	llmModel     string     // LLM model (e.g., "gemini-2.0-flash")
}

// HandleSlackAppMention handles a slack app mention event with LLM integration
func (uc *Slack) HandleSlackAppMention(ctx context.Context, slackMsg slack.Message) error {
	logger := ctxlog.From(ctx)
	logger.Debug("slack app mention event",
		"slack message", slackMsg,
	)

	if uc.slackClient == nil {
		return goerr.New("slack client not configured")
	}

	// Check if LLM is configured
	if uc.llmClient == nil && uc.llmFactory == nil {
		logger.Warn("LLM not configured, falling back to simple response")
		return uc.handleSimpleResponse(ctx, slackMsg)
	}

	// Find first bot mention
	firstBotMention := uc.findFirstBotMention(slackMsg.Mentions)
	if firstBotMention == nil {
		return nil // No bot mention found
	}

	// Parse agent information from mention
	agentMention := uc.parseAgentFromMention(firstBotMention)

	// Analyze thread context
	threadCtx := uc.analyzeThreadContext(ctx, slackMsg)

	// Resolve agent
	agent, err := uc.resolveAgent(ctx, agentMention, threadCtx)
	if err != nil {
		return uc.handleAgentError(ctx, slackMsg, err)
	}

	// Process the bot mention with agent
	return uc.processBotMentionWithAgent(ctx, slackMsg, agentMention, agent)
}

// findFirstBotMention finds the first mention that is for the bot
func (uc *Slack) findFirstBotMention(mentions []slack.Mention) *slack.Mention {
	for _, mention := range mentions {
		if uc.slackClient.IsBotUser(mention.UserID) {
			return &mention
		}
	}
	return nil
}

// storeThreadAndMessage stores the thread and message if repository is available
func (uc *Slack) storeThreadAndMessage(ctx context.Context, slackMsg *slack.Message) types.ThreadID {
	logger := ctxlog.From(ctx)

	if uc.repository == nil {
		return ""
	}

	// Get or create thread atomically
	t, err := uc.repository.GetOrPutThread(ctx, slackMsg.TeamID, slackMsg.Channel, slackMsg.GetThreadTS())
	if err != nil {
		logger.Warn("failed to get or create thread",
			"error", err,
			"team_id", slackMsg.TeamID,
			"channel", slackMsg.Channel,
			"thread_ts", slackMsg.GetThreadTS(),
		)
		return ""
	}

	slackMsg.ThreadID = t.ID

	if err := uc.repository.PutThreadMessage(ctx, t.ID, slackMsg); err != nil {
		logger.Warn("failed to save message",
			"error", err,
			"thread_id", t.ID,
			"message_id", slackMsg.ID,
		)
	}

	return t.ID
}

// handleSimpleResponse provides an error response when LLM is not configured
func (uc *Slack) handleSimpleResponse(ctx context.Context, slackMsg slack.Message) error {
	logger := ctxlog.From(ctx)

	// Find first bot mention
	firstBotMention := uc.findFirstBotMention(slackMsg.Mentions)
	if firstBotMention == nil {
		return nil // No bot mention found
	}

	// Store thread and message
	uc.storeThreadAndMessage(ctx, &slackMsg)

	// Generate error response about LLM not being configured
	responseText := "❌ **LLM not configured**\n\n" +
		"I'm unable to process your request because no Large Language Model (LLM) has been configured for this bot. " +
		"Please contact your administrator to configure an LLM provider (such as Gemini, OpenAI, etc.) to enable AI-powered responses.\n\n" +
		"Available LLM providers:\n" +
		"• Google Gemini\n" +
		"• OpenAI GPT\n" +
		"• Other gollem-supported providers"

	// Reply in thread
	if err := uc.slackClient.PostMessage(ctx, slackMsg.Channel, slackMsg.GetThreadTS(), responseText); err != nil {
		return goerr.Wrap(err, "failed to post message to slack")
	}

	logger.Info("responded to slack mention with LLM configuration error",
		"channel", slackMsg.Channel,
		"thread", slackMsg.GetThreadTS(),
		"user", slackMsg.UserID,
		"message", firstBotMention.Message,
	)
	return nil
}

// HandleSlackMessage handles a slack message event
func (uc *Slack) HandleSlackMessage(ctx context.Context, slackMsg slack.Message) error {
	ctxlog.From(ctx).Debug("slack message event",
		"channel", slackMsg.Channel,
		"thread", slackMsg.GetThreadTS(),
		"text", slackMsg.Text,
	)

	// If repository is available, check if this is in a participating thread
	if uc.repository != nil && slackMsg.ThreadTS != "" {
		// Check if we have this thread in our database (meaning we're participating)
		thread, err := uc.repository.GetThreadByTS(ctx, slackMsg.Channel, slackMsg.ThreadTS)
		if err == nil {
			// This is a participating thread, record the message
			slackMsg.ThreadID = thread.ID

			if err := uc.repository.PutThreadMessage(ctx, thread.ID, &slackMsg); err != nil {
				ctxlog.From(ctx).Warn("failed to save message in participating thread",
					"error", err,
					"thread_id", thread.ID,
					"thread_ts", slackMsg.ThreadTS,
					"message_id", slackMsg.ID,
				)
			} else {
				ctxlog.From(ctx).Debug("recorded message in participating thread",
					"thread_id", thread.ID,
					"thread_ts", slackMsg.ThreadTS,
					"user_id", slackMsg.UserID,
				)
			}
		}
		// If thread not found, just ignore (not a participating thread)
	}

	return nil
}

// parseAgentFromMention parses agent information from a slack mention
func (uc *Slack) parseAgentFromMention(mention *slack.Mention) *slack.AgentMention {
	// Convert the regular mention to agent mention using the parser
	agentMentions := slack.ParseAgentMention(fmt.Sprintf("<@%s> %s", mention.UserID, mention.Message))
	if len(agentMentions) == 0 {
		return nil
	}

	// Return the first (and should be only) agent mention
	return &agentMentions[0]
}

// analyzeThreadContext analyzes the thread context to determine if this is a new thread or existing thread
func (uc *Slack) analyzeThreadContext(ctx context.Context, slackMsg slack.Message) *threadContext {
	logger := ctxlog.From(ctx)

	// Check if repository is available
	if uc.repository == nil {
		// No repository means we can't track threads, treat as new thread requiring agent
		return &threadContext{
			isNewThread:    true,
			existingThread: nil,
			requiresAgent:  true,
		}
	}

	// Try to find existing thread
	thread, err := uc.repository.GetThreadByTS(ctx, slackMsg.Channel, slackMsg.GetThreadTS())
	if err != nil {
		logger.Debug("thread not found, treating as new thread",
			"channel", slackMsg.Channel,
			"thread_ts", slackMsg.GetThreadTS(),
		)
		// Thread not found, this is a new thread
		return &threadContext{
			isNewThread:    true,
			existingThread: nil,
			requiresAgent:  true,
		}
	}

	logger.Debug("found existing thread",
		"thread_id", thread.ID,
		"channel", slackMsg.Channel,
		"thread_ts", slackMsg.GetThreadTS(),
	)

	// Thread found, this is an existing thread
	return &threadContext{
		isNewThread:    false,
		existingThread: thread,
		requiresAgent:  false, // Agent is already determined from the existing thread
	}
}

// resolveAgent resolves agent information based on the mention and thread context
func (uc *Slack) resolveAgent(ctx context.Context, agentMention *slack.AgentMention, threadCtx *threadContext) (*agentContext, error) {
	logger := ctxlog.From(ctx)

	// If this is an existing thread, use the agent information from the thread
	if !threadCtx.isNewThread && threadCtx.existingThread != nil {
		thread := threadCtx.existingThread

		// Check if thread has agent information
		if thread.AgentUUID != nil {
			logger.Debug("using agent from existing thread",
				"thread_id", thread.ID,
				"agent_uuid", *thread.AgentUUID,
				"agent_version", thread.AgentVersion,
			)

			// Check if this is general mode
			if *thread.AgentUUID == generalModeUUID {
				return &agentContext{
					uuid:         generalModeUUID,
					version:      thread.AgentVersion,
					systemPrompt: uc.getGeneralModeSystemPrompt(),
					llmProvider:  "", // Use default from factory
					llmModel:     "", // Use default from factory
				}, nil
			}

			// Get agent information from repository
			if uc.agentRepository != nil {
				_, err := uc.agentRepository.GetAgent(ctx, *thread.AgentUUID)
				if err != nil {
					return nil, goerr.Wrap(err, "failed to get agent from repository",
						goerr.V("agent_uuid", *thread.AgentUUID))
				}

				// Get agent version information
				agentVersion, err := uc.agentRepository.GetAgentVersion(ctx, *thread.AgentUUID, thread.AgentVersion)
				if err != nil {
					return nil, goerr.Wrap(err, "failed to get agent version",
						goerr.V("agent_uuid", *thread.AgentUUID),
						goerr.V("version", thread.AgentVersion))
				}

				return &agentContext{
					uuid:         *thread.AgentUUID,
					version:      thread.AgentVersion,
					systemPrompt: agentVersion.SystemPrompt,
					llmProvider:  string(agentVersion.LLMProvider),
					llmModel:     agentVersion.LLMModel,
				}, nil
			}
		}

		// Existing thread but no agent info (backward compatibility)
		// Use default system prompt
		return &agentContext{
			uuid:         generalModeUUID,
			version:      "legacy",
			systemPrompt: "You are a helpful Slack bot assistant. Respond concisely and helpfully to user questions.",
			llmProvider:  "", // Use default from factory
			llmModel:     "", // Use default from factory
		}, nil
	}

	// This is a new thread, resolve agent from the mention
	if agentMention == nil || agentMention.AgentID == "" {
		// No agent specified, use general mode
		logger.Debug("no agent specified, using general mode")
		return &agentContext{
			uuid:         generalModeUUID,
			version:      "general-v1",
			systemPrompt: uc.getGeneralModeSystemPrompt(),
			llmProvider:  "", // Use default from factory
			llmModel:     "", // Use default from factory
		}, nil
	}

	// Agent ID specified, validate and get agent information
	if uc.agentRepository == nil {
		return nil, goerr.New("agent repository not available")
	}

	logger.Debug("resolving agent", "agent_id", agentMention.AgentID)

	// Get active agent by agent ID
	agentInfo, err := uc.agentRepository.GetAgentByAgentIDActive(ctx, agentMention.AgentID)
	if err != nil {
		return nil, goerr.Wrap(slack.ErrAgentNotFound, "agent not found or archived",
			goerr.V("agent_id", agentMention.AgentID))
	}

	// Get latest version of the agent
	latestVersion, err := uc.agentRepository.GetLatestAgentVersion(ctx, agentInfo.ID)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get latest agent version",
			goerr.V("agent_uuid", agentInfo.ID),
			goerr.V("agent_id", agentMention.AgentID))
	}

	logger.Debug("resolved agent",
		"agent_id", agentMention.AgentID,
		"agent_uuid", agentInfo.ID,
		"version", latestVersion.Version,
	)

	return &agentContext{
		uuid:         agentInfo.ID,
		version:      latestVersion.Version,
		systemPrompt: latestVersion.SystemPrompt,
		llmProvider:  string(latestVersion.LLMProvider),
		llmModel:     latestVersion.LLMModel,
	}, nil
}

// getGeneralModeSystemPrompt returns the hardcoded system prompt for general mode
func (uc *Slack) getGeneralModeSystemPrompt() string {
	return `You are Tamamo's Guide Assistant. Your primary role is to help users understand and effectively use the Tamamo tool system.

Your main responsibilities:
1. Explain how to use Tamamo and its agent system
2. Guide users to the appropriate specialized agents for their tasks
3. Provide clear instructions on agent ID specification: @tamamo <agent_id> [message]
4. List available agents and their specific purposes
5. Help users understand thread-based conversations

When users first contact you without specifying an agent ID, immediately explain:
- How the Tamamo agent system works
- How to specify agent IDs for specialized tasks
- Available agent types and when to use each one
- That they can continue conversations in threads without re-specifying agent IDs

Always respond in the same language as the user's message. Focus on being a knowledgeable guide rather than a general conversational assistant.

After providing guidance, you can answer questions about the tool itself, but always steer users toward using appropriate specialized agents for their actual work tasks.`
}

// handleAgentError handles agent-related errors and sends appropriate error messages to Slack
func (uc *Slack) handleAgentError(ctx context.Context, slackMsg slack.Message, err error) error {
	logger := ctxlog.From(ctx)
	logger.Error("agent error occurred", "error", err)

	var errorMessage string

	// Check if this is an agent not found error
	if errors.Is(err, slack.ErrAgentNotFound) {
		// Extract agent ID from error context
		agentID := ""
		if ge := goerr.Unwrap(err); ge != nil {
			if agentIDVal := ge.Values()["agent_id"]; agentIDVal != nil {
				if agentIDStr, ok := agentIDVal.(string); ok {
					agentID = agentIDStr
				}
			}
		}

		errorMessage = uc.generateAgentErrorMessage(ctx, agentID)
	} else {
		// Generic error message
		errorMessage = "An error occurred while processing your request. Please try again later."
	}

	// Send error message to Slack
	if err := uc.slackClient.PostMessage(ctx, slackMsg.Channel, slackMsg.GetThreadTS(), errorMessage); err != nil {
		return goerr.Wrap(err, "failed to post error message to slack")
	}

	return nil
}

// generateAgentErrorMessage generates a helpful error message when agent is not found
func (uc *Slack) generateAgentErrorMessage(ctx context.Context, agentID string) string {
	logger := ctxlog.From(ctx)

	// Try to get available active agents
	var availableAgents []string
	if uc.agentRepository != nil {
		agents, _, err := uc.agentRepository.ListActiveAgents(ctx, 0, 10) // Get first 10 active agents
		if err != nil {
			logger.Warn("failed to get available active agents", "error", err)
		} else {
			availableAgents = make([]string, 0, len(agents))
			for _, agent := range agents {
				availableAgents = append(availableAgents, fmt.Sprintf("- %s: %s", agent.AgentID, agent.Description))
			}
		}
	}

	// Generate error message with suggestions
	var message strings.Builder
	if agentID != "" {
		message.WriteString(fmt.Sprintf("Agent ID '%s' not found.\n\n", agentID))
	} else {
		message.WriteString("Invalid agent specification.\n\n")
	}

	message.WriteString("Usage: @tamamo <agent_id> [message]\n\n")

	if len(availableAgents) > 0 {
		message.WriteString("Available agents:\n")
		for _, agent := range availableAgents {
			message.WriteString(agent)
			message.WriteString("\n")
		}
		message.WriteString("\n")
	}

	message.WriteString("Or use @tamamo [message] for general mode to get help with using Tamamo.")

	return message.String()
}

// processBotMentionWithAgent processes a bot mention with agent information
func (uc *Slack) processBotMentionWithAgent(ctx context.Context, slackMsg slack.Message, agentMention *slack.AgentMention, agent *agentContext) error {
	logger := ctxlog.From(ctx)

	// Store thread with agent information and get thread ID
	threadID := uc.storeThreadWithAgent(ctx, &slackMsg, &agent.uuid, agent.version)

	// Determine the user message
	userMessage := ""
	if agentMention != nil {
		userMessage = agentMention.Message
	}

	// Start chat conversation with agent-specific system prompt
	if err := uc.chatWithAgent(ctx, slackMsg, threadID, userMessage, agent); err != nil {
		// Log the error with context
		pkgErrors.Handle(ctx, err)

		// Notify user about the error
		errMsg := "I apologize, but I'm experiencing issues processing your request. Please try again later."
		if slackErr := uc.slackClient.PostMessage(ctx, slackMsg.Channel, slackMsg.GetThreadTS(), errMsg); slackErr != nil {
			logger.Error("failed to post error message to slack",
				"error", slackErr,
				"original_error", err,
			)
		}

		return err
	}

	return nil
}

// storeThreadWithAgent stores the thread with agent information and returns thread ID
func (uc *Slack) storeThreadWithAgent(ctx context.Context, slackMsg *slack.Message, agentUUID *types.UUID, agentVersion string) types.ThreadID {
	logger := ctxlog.From(ctx)

	if uc.repository == nil {
		return ""
	}

	// Get or create thread atomically with agent information
	t, err := uc.repository.GetOrPutThreadWithAgent(ctx, slackMsg.TeamID, slackMsg.Channel, slackMsg.GetThreadTS(), agentUUID, agentVersion)
	if err != nil {
		logger.Warn("failed to get or create thread with agent",
			"error", err,
			"team_id", slackMsg.TeamID,
			"channel", slackMsg.Channel,
			"thread_ts", slackMsg.GetThreadTS(),
			"agent_uuid", agentUUID,
			"agent_version", agentVersion,
		)
		return ""
	}

	slackMsg.ThreadID = t.ID

	if err := uc.repository.PutThreadMessage(ctx, t.ID, slackMsg); err != nil {
		logger.Warn("failed to save message",
			"error", err,
			"thread_id", t.ID,
			"message_id", slackMsg.ID,
		)
	}

	return t.ID
}

// chatWithAgent handles the conversation with LLM using agent-specific configuration
func (uc *Slack) chatWithAgent(ctx context.Context, slackMsg slack.Message, threadID types.ThreadID, userMessage string, agent *agentContext) error {
	logger := ctxlog.From(ctx)

	// Load conversation history if thread exists
	var history *gollem.History
	if threadID.IsValid() && uc.repository != nil && uc.storageRepo != nil {
		logger.Debug("attempting to load history for thread",
			"thread_id", threadID,
		)
		// Get the latest history for this thread
		latestHistory, err := uc.repository.GetLatestHistory(ctx, threadID)
		if err != nil {
			if errors.Is(err, slack.ErrHistoryNotFound) {
				// It's normal for new threads to not have history yet
				logger.Debug("no existing history for thread",
					"thread_id", threadID,
				)
			} else {
				// Log other errors as warnings, as this might indicate a problem
				logger.Warn("failed to get latest history, starting new conversation",
					"error", err,
					"thread_id", threadID,
				)
			}
		} else if latestHistory == nil {
			logger.Debug("no history found for thread",
				"thread_id", threadID,
			)
		} else {
			// Load gollem history from storage
			storedHistory, err := uc.storageRepo.LoadHistoryJSON(ctx, threadID, latestHistory.ID)
			if err != nil {
				logger.Warn("failed to load history from storage, but ignore it and start without history",
					"error", err,
					"thread_id", threadID,
					"history_id", latestHistory.ID,
				)
			} else {
				history = &storedHistory
				logger.Debug("loaded conversation history",
					"thread_id", threadID,
					"history_id", latestHistory.ID,
					"message_count", history.ToCount(),
				)
			}
		}
	} else {
		logger.Debug("conditions not met for loading history",
			"thread_id_valid", threadID.IsValid(),
			"has_repository", uc.repository != nil,
			"has_storage_repo", uc.storageRepo != nil,
		)
	}

	// Get the appropriate LLM client
	llmClient, err := uc.getLLMClient(ctx, agent, slackMsg)
	if err != nil {
		return goerr.Wrap(err, "failed to get LLM client")
	}

	// Create session with agent-specific system prompt and history if available
	sessionOptions := []gollem.SessionOption{
		gollem.WithSessionSystemPrompt(agent.systemPrompt),
	}

	if history != nil {
		sessionOptions = append(sessionOptions, gollem.WithSessionHistory(history))
	}

	// Create a new session for this conversation
	session, err := llmClient.NewSession(ctx, sessionOptions...)
	if err != nil {
		return goerr.Wrap(err, "failed to create LLM session",
			goerr.V("thread_id", threadID),
			goerr.V("channel", slackMsg.Channel),
			goerr.V("user", slackMsg.UserID),
			goerr.V("agent_uuid", agent.uuid),
		)
	}

	// Generate content directly through session
	resp, err := session.GenerateContent(ctx, gollem.Text(userMessage))
	if err != nil {
		return goerr.Wrap(err, "failed to generate content with LLM",
			goerr.V("thread_id", threadID),
			goerr.V("message", userMessage),
			goerr.V("channel", slackMsg.Channel),
			goerr.V("agent_uuid", agent.uuid),
		)
	}

	var responseText string
	if resp != nil && len(resp.Texts) > 0 {
		responseText = resp.Texts[0]
	}

	// If no response was captured, use a fallback
	if responseText == "" {
		responseText = "(no response)"
	}

	// Send response to Slack with agent-specific display
	if err := uc.postMessageWithAgentDisplay(ctx, slackMsg.Channel, slackMsg.GetThreadTS(), responseText, agent); err != nil {
		return goerr.Wrap(err, "failed to post message to slack")
	}

	logger.Info("responded to slack mention with LLM",
		"channel", slackMsg.Channel,
		"thread", slackMsg.GetThreadTS(),
		"user", slackMsg.UserID,
		"message", userMessage,
		"agent_uuid", agent.uuid,
		"agent_version", agent.version,
	)

	// Save updated history to storage for future use
	if threadID.IsValid() && uc.repository != nil && uc.storageRepo != nil && session != nil {
		// Get the session's history
		updatedHistory := session.History()
		if updatedHistory != nil && updatedHistory.ToCount() > 0 {
			// Create history record with consistent ID
			historyRecord := slack.NewHistory(ctx, threadID)
			historyID := historyRecord.ID

			// Save gollem history to storage
			if err := uc.storageRepo.SaveHistoryJSON(ctx, threadID, historyID, updatedHistory); err != nil {
				logger.Warn("failed to save history to storage",
					"error", err,
					"thread_id", threadID,
					"history_id", historyID,
				)
			} else {
				// Save history record to repository
				if err := uc.repository.PutHistory(ctx, historyRecord); err != nil {
					logger.Warn("failed to save history record",
						"error", err,
						"thread_id", threadID,
						"history_id", historyID,
					)
				} else {
					logger.Debug("saved session history",
						"thread_id", threadID,
						"history_id", historyID,
						"message_count", updatedHistory.ToCount(),
						"created_at", historyRecord.CreatedAt,
					)
				}
			}
		}
	}

	return nil
}

// getLLMClient retrieves the appropriate LLM client based on agent configuration
func (uc *Slack) getLLMClient(ctx context.Context, agent *agentContext, slackMsg slack.Message) (gollem.LLMClient, error) {
	logger := ctxlog.From(ctx)

	if uc.llmFactory != nil && agent.llmProvider != "" && agent.llmModel != "" {
		// Use factory to get provider-specific client
		llmClient, err := uc.llmFactory.CreateClient(ctx, agent.llmProvider, agent.llmModel)
		if err != nil {
			logger.Warn("failed to create LLM client from factory, attempting fallback",
				"provider", agent.llmProvider,
				"model", agent.llmModel,
				"error", err,
			)

			// Try fallback if enabled
			fallbackClient, fallbackErr := uc.llmFactory.GetFallbackClient(ctx)
			if fallbackErr != nil {
				return nil, goerr.Wrap(err, "failed to create LLM client and fallback also failed",
					goerr.V("provider", agent.llmProvider),
					goerr.V("model", agent.llmModel),
					goerr.V("fallback_error", fallbackErr),
				)
			}

			// Send warning to Slack about fallback
			warningMsg := fmt.Sprintf("⚠️ Failed to use %s/%s, falling back to default provider", agent.llmProvider, agent.llmModel)
			_ = uc.slackClient.PostMessage(ctx, slackMsg.Channel, slackMsg.GetThreadTS(), warningMsg)

			return fallbackClient, nil
		}
		return llmClient, nil
	} else if uc.llmClient != nil {
		// Use legacy client if factory not available
		return uc.llmClient, nil
	}

	return nil, goerr.New("no LLM client available")
}

// postMessageWithAgentDisplay posts a message to Slack with agent-specific display settings
func (uc *Slack) postMessageWithAgentDisplay(ctx context.Context, channelID, threadTS, text string, agent *agentContext) error {
	// Use basic PostMessage if no agent context or for general mode
	if agent == nil || agent.uuid == generalModeUUID {
		return uc.slackClient.PostMessage(ctx, channelID, threadTS, text)
	}

	// Get agent information for custom display
	agentInfo, err := uc.getAgentDisplayInfo(ctx, agent)
	if err != nil {
		// Log warning but fallback to basic message
		ctxlog.From(ctx).Warn("failed to get agent display info, using basic message",
			"agent_uuid", agent.uuid,
			"error", err,
		)
		return uc.slackClient.PostMessage(ctx, channelID, threadTS, text)
	}

	// Post message with agent-specific options
	return uc.slackClient.PostMessageWithOptions(ctx, channelID, threadTS, text, agentInfo)
}

// getAgentDisplayInfo retrieves agent display information for Slack messages
func (uc *Slack) getAgentDisplayInfo(ctx context.Context, agent *agentContext) (*interfaces.SlackMessageOptions, error) {
	if uc.agentRepository == nil {
		return nil, goerr.New("agent repository not available")
	}

	// Get agent information
	agentInfo, err := uc.agentRepository.GetAgent(ctx, agent.uuid)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get agent info")
	}

	options := &interfaces.SlackMessageOptions{
		Username: agentInfo.Name, // Use agent name as display name
	}

	// Get agent image URL if agent has an image
	if agentInfo.ImageID != nil && uc.agentImageRepo != nil && uc.serverBaseURL != "" {
		// Use FRONTEND_URL as base for public access to agent images
		imageURL := uc.serverBaseURL + "/api/agents/" + agentInfo.ID.String() + "/image?size=72"
		options.IconURL = imageURL

		ctxlog.From(ctx).Debug("using agent custom image",
			"agent_id", agentInfo.ID,
			"image_id", agentInfo.ImageID,
			"image_url", imageURL,
		)
	}

	return options, nil
}
