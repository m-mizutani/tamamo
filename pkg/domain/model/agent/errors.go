package agent

import "github.com/m-mizutani/goerr/v2"

// Error definitions for agent-related operations
var (
	// ErrAgentNotFound is returned when a requested agent cannot be found
	ErrAgentNotFound = goerr.New("agent not found")

	// ErrAgentAlreadyExists is returned when trying to create an agent with an existing ID
	ErrAgentAlreadyExists = goerr.New("agent already exists")

	// ErrAgentArchived is returned when trying to use an archived agent
	ErrAgentArchived = goerr.New("agent is archived")

	// ErrInvalidAgentStatus is returned when an invalid status is provided
	ErrInvalidAgentStatus = goerr.New("invalid agent status")

	// ErrAgentVersionNotFound is returned when a requested agent version cannot be found
	ErrAgentVersionNotFound = goerr.New("agent version not found")

	// ErrInvalidAgentID is returned when an invalid agent ID is provided
	ErrInvalidAgentID = goerr.New("invalid agent ID")

	// ErrCannotArchiveLastAgent is returned when trying to archive the last active agent
	ErrCannotArchiveLastAgent = goerr.New("cannot archive the last active agent")
)
