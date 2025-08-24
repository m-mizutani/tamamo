package agent

import (
	"regexp"

	"github.com/m-mizutani/goerr/v2"
)

var (
	// AgentID format: alphanumeric characters + '_', '-', '.' allowed except at the beginning and end
	// Examples: "agent1", "agent-1", "agent_1", "agent.1", "agent-1.2_3"
	agentIDRegex = regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9._-]*[a-zA-Z0-9])?$`)

	// Semantic versioning format: Major.Minor.Patch
	// Examples: "1.0.0", "1.2.3", "10.0.0"
	versionRegex = regexp.MustCompile(`^\d+\.\d+\.\d+$`)
)

// ValidateAgentID validates the format of an agent ID
func ValidateAgentID(agentID string) error {
	if agentID == "" {
		return goerr.New("agent ID cannot be empty")
	}

	if len(agentID) > 64 {
		return goerr.New("agent ID cannot be longer than 64 characters")
	}

	if !agentIDRegex.MatchString(agentID) {
		return goerr.New("agent ID format is invalid",
			goerr.V("format", "alphanumeric characters with '_', '-', '.' allowed except at beginning and end"),
			goerr.V("agentID", agentID))
	}

	return nil
}

// ValidateVersion validates the format of a version string
func ValidateVersion(version string) error {
	if version == "" {
		return goerr.New("version cannot be empty")
	}

	if !versionRegex.MatchString(version) {
		return goerr.New("version format is invalid",
			goerr.V("format", "semantic versioning (Major.Minor.Patch)"),
			goerr.V("version", version))
	}

	return nil
}

// ValidateAgent validates the Agent struct
func ValidateAgent(agent *Agent) error {
	if err := ValidateAgentID(agent.AgentID); err != nil {
		return goerr.Wrap(err, "invalid agent ID")
	}

	if agent.Name == "" {
		return goerr.New("agent name cannot be empty")
	}

	if len(agent.Name) > 100 {
		return goerr.New("agent name cannot be longer than 100 characters")
	}

	if len(agent.Description) > 1000 {
		return goerr.New("agent description cannot be longer than 1000 characters")
	}

	if err := ValidateStatus(agent.Status); err != nil {
		return goerr.Wrap(err, "invalid agent status")
	}

	if err := ValidateVersion(agent.Latest); err != nil {
		return goerr.Wrap(err, "invalid latest version")
	}

	return nil
}

// ValidateAgentVersion validates the AgentVersion struct
func ValidateAgentVersion(version *AgentVersion) error {
	if err := ValidateVersion(version.Version); err != nil {
		return goerr.Wrap(err, "invalid version")
	}

	// LLM provider must be valid if specified
	if version.LLMProvider != "" && !version.LLMProvider.IsValid() {
		return goerr.New("invalid LLM provider",
			goerr.V("provider", version.LLMProvider.String()))
	}

	// LLM model validation
	if version.LLMModel != "" && len(version.LLMModel) > 100 {
		return goerr.New("LLM model cannot be longer than 100 characters")
	}

	if len(version.SystemPrompt) > 50000 {
		return goerr.New("system prompt cannot be longer than 50000 characters")
	}

	return nil
}
