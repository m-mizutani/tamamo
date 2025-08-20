package usecase

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/agent"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
)

type agentUseCaseImpl struct {
	agentRepo interfaces.AgentRepository
}

// NewAgentUseCases creates a new agent use case implementation
func NewAgentUseCases(agentRepo interfaces.AgentRepository) interfaces.AgentUseCases {
	return &agentUseCaseImpl{
		agentRepo: agentRepo,
	}
}

// CreateAgent creates a new agent with its initial version
func (u *agentUseCaseImpl) CreateAgent(ctx context.Context, req *interfaces.CreateAgentRequest) (*agent.Agent, error) {
	if req == nil {
		return nil, goerr.New("create agent request cannot be nil")
	}

	// Set default version if not provided
	version := req.Version
	if version == "" {
		version = "1.0.0"
	}

	// Validate agent data
	if err := agent.ValidateAgentID(req.AgentID); err != nil {
		return nil, goerr.Wrap(err, "invalid agent ID")
	}

	if err := agent.ValidateVersion(version); err != nil {
		return nil, goerr.Wrap(err, "invalid version")
	}

	// Check if agent ID already exists
	exists, err := u.agentRepo.AgentIDExists(ctx, req.AgentID)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to check agent ID existence")
	}
	if exists {
		return nil, goerr.New("agent ID already exists", goerr.V("agent_id", req.AgentID))
	}

	// Create agent
	now := time.Now()

	// Handle optional fields
	description := ""
	if req.Description != nil {
		description = *req.Description
	}

	agentObj := &agent.Agent{
		ID:          types.NewUUID(ctx),
		AgentID:     req.AgentID,
		Name:        req.Name,
		Description: description,
		Author:      "anonymous",        // As per requirement
		Status:      agent.StatusActive, // Default to active
		Latest:      version,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Validate the complete agent
	if err := agent.ValidateAgent(agentObj); err != nil {
		return nil, goerr.Wrap(err, "agent validation failed")
	}

	// Create agent in repository
	if err := u.agentRepo.CreateAgent(ctx, agentObj); err != nil {
		return nil, goerr.Wrap(err, "failed to create agent")
	}

	// Create initial version
	// Handle optional system prompt
	systemPrompt := ""
	if req.SystemPrompt != nil {
		systemPrompt = *req.SystemPrompt
	}

	agentVersion := &agent.AgentVersion{
		AgentUUID:    agentObj.ID,
		Version:      version,
		SystemPrompt: systemPrompt,
		LLMProvider:  req.LLMProvider,
		LLMModel:     req.LLMModel,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Validate the agent version
	if err := agent.ValidateAgentVersion(agentVersion); err != nil {
		return nil, goerr.Wrap(err, "agent version validation failed")
	}

	// Create version in repository
	if err := u.agentRepo.CreateAgentVersion(ctx, agentVersion); err != nil {
		// Clean up: delete the created agent if version creation fails
		_ = u.agentRepo.DeleteAgent(ctx, agentObj.ID)
		return nil, goerr.Wrap(err, "failed to create agent version")
	}

	return agentObj, nil
}

// GetAgent retrieves an agent with its latest version
func (u *agentUseCaseImpl) GetAgent(ctx context.Context, id types.UUID) (*interfaces.AgentWithVersion, error) {
	if !id.IsValid() {
		return nil, goerr.New("invalid agent ID")
	}

	// Get agent
	agentObj, err := u.agentRepo.GetAgent(ctx, id)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get agent")
	}

	// Get latest version
	latestVersion, err := u.agentRepo.GetLatestAgentVersion(ctx, id)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get latest agent version")
	}

	return &interfaces.AgentWithVersion{
		Agent:         agentObj,
		LatestVersion: latestVersion,
	}, nil
}

// UpdateAgent updates an existing agent
func (u *agentUseCaseImpl) UpdateAgent(ctx context.Context, id types.UUID, req *interfaces.UpdateAgentRequest) (*agent.Agent, error) {
	if req == nil {
		return nil, goerr.New("update agent request cannot be nil")
	}

	// Get existing agent
	agentObj, err := u.agentRepo.GetAgent(ctx, id)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get agent for update")
	}

	// Update fields if provided
	if req.AgentID != nil {
		if err := agent.ValidateAgentID(*req.AgentID); err != nil {
			return nil, goerr.Wrap(err, "invalid agent ID")
		}

		// Check if new agent ID conflicts with existing ones
		if *req.AgentID != agentObj.AgentID {
			exists, err := u.agentRepo.AgentIDExists(ctx, *req.AgentID)
			if err != nil {
				return nil, goerr.Wrap(err, "failed to check agent ID existence")
			}
			if exists {
				return nil, goerr.New("agent ID already exists", goerr.V("agent_id", *req.AgentID))
			}
		}

		agentObj.AgentID = *req.AgentID
	}

	if req.Name != nil {
		agentObj.Name = *req.Name
	}

	if req.Description != nil {
		agentObj.Description = *req.Description
	}

	// Check if version-related fields are being updated
	needsNewVersion := req.SystemPrompt != nil || req.LLMProvider != nil || req.LLMModel != nil

	if needsNewVersion {
		// Get current latest version to increment
		latestVersion, err := u.agentRepo.GetLatestAgentVersion(ctx, id)
		if err != nil {
			return nil, goerr.Wrap(err, "failed to get latest version for update")
		}

		// Create new version with updated fields
		newVersionReq := &interfaces.CreateVersionRequest{
			AgentUUID:   id,
			Version:     incrementVersion(latestVersion.Version), // Simple increment
			LLMProvider: latestVersion.LLMProvider,
			LLMModel:    latestVersion.LLMModel,
		}

		// Use existing system prompt by default
		systemPrompt := latestVersion.SystemPrompt
		if req.SystemPrompt != nil {
			systemPrompt = *req.SystemPrompt
		}
		newVersionReq.SystemPrompt = &systemPrompt

		// Override with new values if provided
		if req.LLMProvider != nil {
			newVersionReq.LLMProvider = *req.LLMProvider
		}
		if req.LLMModel != nil {
			newVersionReq.LLMModel = *req.LLMModel
		}

		// Create the new version
		_, err = u.CreateAgentVersion(ctx, newVersionReq)
		if err != nil {
			return nil, goerr.Wrap(err, "failed to create new version with updated configuration")
		}

		// Re-fetch the agent to get updated Latest version
		agentObj, err = u.agentRepo.GetAgent(ctx, id)
		if err != nil {
			return nil, goerr.Wrap(err, "failed to get updated agent after version creation")
		}
	}

	// Validate the updated agent
	if err := agent.ValidateAgent(agentObj); err != nil {
		return nil, goerr.Wrap(err, "agent validation failed")
	}

	// Update in repository (only if non-version fields were updated)
	if req.AgentID != nil || req.Name != nil || req.Description != nil {
		if err := u.agentRepo.UpdateAgent(ctx, agentObj); err != nil {
			return nil, goerr.Wrap(err, "failed to update agent")
		}
	}

	return agentObj, nil
}

// DeleteAgent deletes an agent and all its versions
func (u *agentUseCaseImpl) DeleteAgent(ctx context.Context, id types.UUID) error {
	if !id.IsValid() {
		return goerr.New("invalid agent ID")
	}

	// Check if agent exists
	_, err := u.agentRepo.GetAgent(ctx, id)
	if err != nil {
		return goerr.Wrap(err, "failed to get agent for deletion")
	}

	// Delete agent (this should also delete all versions)
	if err := u.agentRepo.DeleteAgent(ctx, id); err != nil {
		return goerr.Wrap(err, "failed to delete agent")
	}

	return nil
}

// ListAgents retrieves a list of active agents with their latest versions
func (u *agentUseCaseImpl) ListAgents(ctx context.Context, offset, limit int) (*interfaces.AgentListResponse, error) {
	if offset < 0 || limit < 0 {
		return nil, goerr.New("offset and limit must be non-negative")
	}

	// Get only active agents
	agents, totalCount, err := u.agentRepo.ListActiveAgents(ctx, offset, limit)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to list active agents")
	}

	// Get latest versions for the agents
	agentsWithVersions := make([]*interfaces.AgentWithVersion, 0, len(agents))
	for _, agentObj := range agents {
		// Get latest version for this agent
		latestVersion, err := u.agentRepo.GetLatestAgentVersion(ctx, agentObj.ID)
		if err != nil {
			// If version not found, create AgentWithVersion with nil version
			latestVersion = nil
		}

		agentsWithVersions = append(agentsWithVersions, &interfaces.AgentWithVersion{
			Agent:         agentObj,
			LatestVersion: latestVersion,
		})
	}

	return &interfaces.AgentListResponse{
		Agents:     agentsWithVersions,
		TotalCount: totalCount,
	}, nil
}

// ListAllAgents retrieves a list of all agents (both active and archived) with their latest versions
func (u *agentUseCaseImpl) ListAllAgents(ctx context.Context, offset, limit int) (*interfaces.AgentListResponse, error) {
	if offset < 0 || limit < 0 {
		return nil, goerr.New("offset and limit must be non-negative")
	}

	// Get all agents regardless of status
	agents, totalCount, err := u.agentRepo.ListAgents(ctx, offset, limit)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to list all agents")
	}

	// Get latest versions for the agents
	agentsWithVersions := make([]*interfaces.AgentWithVersion, 0, len(agents))
	for _, agentObj := range agents {
		// Get latest version for this agent
		latestVersion, err := u.agentRepo.GetLatestAgentVersion(ctx, agentObj.ID)
		if err != nil {
			// If version not found, create AgentWithVersion with nil version
			latestVersion = nil
		}

		agentsWithVersions = append(agentsWithVersions, &interfaces.AgentWithVersion{
			Agent:         agentObj,
			LatestVersion: latestVersion,
		})
	}

	return &interfaces.AgentListResponse{
		Agents:     agentsWithVersions,
		TotalCount: totalCount,
	}, nil
}

// CreateAgentVersion creates a new version for an existing agent
func (u *agentUseCaseImpl) CreateAgentVersion(ctx context.Context, req *interfaces.CreateVersionRequest) (*agent.AgentVersion, error) {
	if req == nil {
		return nil, goerr.New("create version request cannot be nil")
	}

	// Validate version
	if err := agent.ValidateVersion(req.Version); err != nil {
		return nil, goerr.Wrap(err, "invalid version")
	}

	// Check if agent exists
	agentObj, err := u.agentRepo.GetAgent(ctx, req.AgentUUID)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get agent for version creation")
	}

	// Create agent version
	now := time.Now()

	// Handle optional system prompt
	systemPrompt := ""
	if req.SystemPrompt != nil {
		systemPrompt = *req.SystemPrompt
	}

	agentVersion := &agent.AgentVersion{
		AgentUUID:    req.AgentUUID,
		Version:      req.Version,
		SystemPrompt: systemPrompt,
		LLMProvider:  req.LLMProvider,
		LLMModel:     req.LLMModel,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Validate the agent version
	if err := agent.ValidateAgentVersion(agentVersion); err != nil {
		return nil, goerr.Wrap(err, "agent version validation failed")
	}

	// Create version in repository
	if err := u.agentRepo.CreateAgentVersion(ctx, agentVersion); err != nil {
		return nil, goerr.Wrap(err, "failed to create agent version")
	}

	// Update agent's latest version if this is a newer version
	// For simplicity, we'll just update the latest field
	// In a real implementation, you might want to use semantic version comparison
	agentObj.Latest = req.Version
	agentObj.UpdatedAt = now
	if err := u.agentRepo.UpdateAgent(ctx, agentObj); err != nil {
		// Log the error but don't fail the version creation since the version was already created
		// This creates data inconsistency that needs to be addressed
		return agentVersion, goerr.Wrap(err, "version created but failed to update agent's latest version tag",
			goerr.V("agent_uuid", req.AgentUUID),
			goerr.V("version", req.Version),
			goerr.V("agent_id", agentObj.AgentID))
	}

	return agentVersion, nil
}

// GetAgentVersions retrieves all versions of an agent
func (u *agentUseCaseImpl) GetAgentVersions(ctx context.Context, agentUUID types.UUID) ([]*agent.AgentVersion, error) {
	if !agentUUID.IsValid() {
		return nil, goerr.New("invalid agent UUID")
	}

	// Check if agent exists
	_, err := u.agentRepo.GetAgent(ctx, agentUUID)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get agent")
	}

	// Get versions
	versions, err := u.agentRepo.ListAgentVersions(ctx, agentUUID)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to list agent versions")
	}

	return versions, nil
}

// CheckAgentIDAvailability checks if an agent ID is available
func (u *agentUseCaseImpl) CheckAgentIDAvailability(ctx context.Context, agentID string) (*interfaces.AgentIDAvailability, error) {
	// Validate format first
	if err := agent.ValidateAgentID(agentID); err != nil {
		return &interfaces.AgentIDAvailability{
			Available: false,
			Message:   err.Error(),
		}, nil
	}

	// Check if it exists
	exists, err := u.agentRepo.AgentIDExists(ctx, agentID)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to check agent ID existence")
	}

	if exists {
		return &interfaces.AgentIDAvailability{
			Available: false,
			Message:   "Agent ID is already taken",
		}, nil
	}

	return &interfaces.AgentIDAvailability{
		Available: true,
		Message:   "Agent ID is available",
	}, nil
}

// ValidateAgentID validates an agent ID format
func (u *agentUseCaseImpl) ValidateAgentID(agentID string) error {
	return agent.ValidateAgentID(agentID)
}

// ValidateVersion validates a version format
func (u *agentUseCaseImpl) ValidateVersion(version string) error {
	return agent.ValidateVersion(version)
}

// incrementVersion increments the patch version of a semantic version string
func incrementVersion(version string) string {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		// Fallback to simple increment
		return version + ".1"
	}

	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		// Fallback to simple increment
		return version + ".1"
	}

	return fmt.Sprintf("%s.%s.%d", parts[0], parts[1], patch+1)
}

// ArchiveAgent archives an existing agent
func (u *agentUseCaseImpl) ArchiveAgent(ctx context.Context, id types.UUID) (*agent.Agent, error) {
	if !id.IsValid() {
		return nil, goerr.New("invalid agent ID")
	}

	// Get existing agent
	agentObj, err := u.agentRepo.GetAgent(ctx, id)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get agent for archiving")
	}

	// Check if already archived
	if agentObj.Status == agent.StatusArchived {
		return nil, goerr.New("agent is already archived", goerr.V("agent_id", agentObj.AgentID))
	}

	// Update status to archived
	if err := u.agentRepo.UpdateAgentStatus(ctx, id, agent.StatusArchived); err != nil {
		return nil, goerr.Wrap(err, "failed to archive agent")
	}

	// Return updated agent
	agentObj.Status = agent.StatusArchived
	agentObj.UpdatedAt = time.Now()

	return agentObj, nil
}

// UnarchiveAgent unarchives an existing agent (sets to active)
func (u *agentUseCaseImpl) UnarchiveAgent(ctx context.Context, id types.UUID) (*agent.Agent, error) {
	if !id.IsValid() {
		return nil, goerr.New("invalid agent ID")
	}

	// Get existing agent
	agentObj, err := u.agentRepo.GetAgent(ctx, id)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get agent for unarchiving")
	}

	// Check if already active
	if agentObj.Status == agent.StatusActive {
		return nil, goerr.New("agent is already active", goerr.V("agent_id", agentObj.AgentID))
	}

	// Update status to active
	if err := u.agentRepo.UpdateAgentStatus(ctx, id, agent.StatusActive); err != nil {
		return nil, goerr.Wrap(err, "failed to unarchive agent")
	}

	// Return updated agent
	agentObj.Status = agent.StatusActive
	agentObj.UpdatedAt = time.Now()

	return agentObj, nil
}

// ListAgentsByStatus retrieves a list of agents with a specific status
func (u *agentUseCaseImpl) ListAgentsByStatus(ctx context.Context, status agent.Status, offset, limit int) (*interfaces.AgentListResponse, error) {
	if offset < 0 || limit < 0 {
		return nil, goerr.New("offset and limit must be non-negative")
	}

	// Get agents by status
	agents, totalCount, err := u.agentRepo.ListAgentsByStatus(ctx, status, offset, limit)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to list agents by status",
			goerr.V("status", status),
			goerr.V("offset", offset),
			goerr.V("limit", limit))
	}

	// Get latest versions for the agents
	agentsWithVersions := make([]*interfaces.AgentWithVersion, 0, len(agents))
	for _, agentObj := range agents {
		// Get latest version for this agent
		latestVersion, err := u.agentRepo.GetLatestAgentVersion(ctx, agentObj.ID)
		if err != nil {
			// If version not found, create AgentWithVersion with nil version
			latestVersion = nil
		}

		agentsWithVersions = append(agentsWithVersions, &interfaces.AgentWithVersion{
			Agent:         agentObj,
			LatestVersion: latestVersion,
		})
	}

	return &interfaces.AgentListResponse{
		Agents:     agentsWithVersions,
		TotalCount: totalCount,
	}, nil
}
