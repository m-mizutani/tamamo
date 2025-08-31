package apperr

import "github.com/m-mizutani/goerr/v2"

// Agent related errors
var (
	ErrAgentNotFound = goerr.New("agent not found",
		goerr.T(ErrTagAgentNotFound)).ID("ERR_AGENT_NOT_FOUND")

	ErrInvalidAgentID = goerr.New("invalid agent ID format",
		goerr.T(ErrTagValidation)).ID("ERR_INVALID_AGENT_ID")

	ErrAgentCreationFailed = goerr.New("failed to create agent",
		goerr.T(ErrTagInternal)).ID("ERR_AGENT_CREATION_FAILED")
)

// LLM related errors
var (
	ErrLLMNotConfigured = goerr.New("LLM not configured",
		goerr.T(ErrTagInternal)).ID("ERR_LLM_NOT_CONFIGURED")

	ErrLLMProviderNotSupported = goerr.New("LLM provider not supported",
		goerr.T(ErrTagValidation)).ID("ERR_LLM_PROVIDER_NOT_SUPPORTED")

	ErrLLMAPIFailed = goerr.New("LLM API call failed",
		goerr.T(ErrTagLLMError)).ID("ERR_LLM_API_FAILED")
)

// Slack related errors
var (
	ErrSlackAPIFailed = goerr.New("Slack API call failed",
		goerr.T(ErrTagSlackAPI)).ID("ERR_SLACK_API_FAILED")

	ErrSlackAuthenticationFailed = goerr.New("Slack authentication failed",
		goerr.T(ErrTagUnauthorized)).ID("ERR_SLACK_AUTH_FAILED")

	ErrSlackChannelNotFound = goerr.New("Slack channel not found",
		goerr.T(ErrTagNotFound)).ID("ERR_SLACK_CHANNEL_NOT_FOUND")
)

// Firestore related errors
var (
	ErrFirestoreConnection = goerr.New("Firestore connection failed",
		goerr.T(ErrTagFirestore)).ID("ERR_FIRESTORE_CONNECTION")

	ErrFirestoreDocumentNotFound = goerr.New("Firestore document not found",
		goerr.T(ErrTagNotFound)).ID("ERR_FIRESTORE_DOC_NOT_FOUND")

	ErrFirestoreOperationFailed = goerr.New("Firestore operation failed",
		goerr.T(ErrTagFirestore)).ID("ERR_FIRESTORE_OP_FAILED")
)

// Thread related errors
var (
	ErrThreadNotFound = goerr.New("thread not found",
		goerr.T(ErrTagThreadNotFound)).ID("ERR_THREAD_NOT_FOUND")

	ErrThreadCreationFailed = goerr.New("failed to create thread",
		goerr.T(ErrTagInternal)).ID("ERR_THREAD_CREATION_FAILED")
)

// Message related errors
var (
	ErrMessageNotFound = goerr.New("message not found",
		goerr.T(ErrTagMessageNotFound)).ID("ERR_MESSAGE_NOT_FOUND")

	ErrInvalidMessage = goerr.New("invalid message format",
		goerr.T(ErrTagValidation)).ID("ERR_INVALID_MESSAGE")
)
