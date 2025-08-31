package apperr

import (
	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
)

// Slack related keys
var (
	UserIDKey    = goerr.NewTypedKey[string]("user_id")
	ChannelIDKey = goerr.NewTypedKey[string]("channel_id")
	TeamIDKey    = goerr.NewTypedKey[string]("team_id")
	ThreadTSKey  = goerr.NewTypedKey[string]("thread_ts")
	BotIDKey     = goerr.NewTypedKey[string]("bot_id")
)

// Domain Entity related keys
var (
	ThreadIDKey  = goerr.NewTypedKey[types.ThreadID]("thread_id")
	MessageIDKey = goerr.NewTypedKey[types.MessageID]("message_id")
	AgentUUIDKey = goerr.NewTypedKey[types.UUID]("agent_uuid")
	AgentIDKey   = goerr.NewTypedKey[string]("agent_id")
	HistoryIDKey = goerr.NewTypedKey[types.HistoryID]("history_id")
)

// Processing related keys
var (
	RequestIDKey  = goerr.NewTypedKey[string]("request_id")
	FilenameKey   = goerr.NewTypedKey[string]("filename")
	ErrorCountKey = goerr.NewTypedKey[int]("error_count")
	RetryCountKey = goerr.NewTypedKey[int]("retry_count")
	OperationKey  = goerr.NewTypedKey[string]("operation")
)

// LLM related keys
var (
	LLMProviderKey = goerr.NewTypedKey[string]("llm_provider")
	LLMModelKey    = goerr.NewTypedKey[string]("llm_model")
	TokenCountKey  = goerr.NewTypedKey[int]("token_count")
)

// Firestore related keys
var (
	CollectionKey = goerr.NewTypedKey[string]("collection")
	DocumentIDKey = goerr.NewTypedKey[string]("document_id")
	ProjectIDKey  = goerr.NewTypedKey[string]("project_id")
)
