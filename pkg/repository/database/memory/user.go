package memory

import (
	"context"
	"sync"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/user"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
)

// userStorage holds user data in memory
type userStorage struct {
	users        map[types.UserID]*user.User
	slackIDIndex map[string]map[string]types.UserID // slackID -> teamID -> userID
	mu           sync.RWMutex
}

// NewUserRepository creates a new user repository
func NewUserRepository() interfaces.UserRepository {
	return &userStorage{
		users:        make(map[types.UserID]*user.User),
		slackIDIndex: make(map[string]map[string]types.UserID),
	}
}

// newUserStorage creates a new user storage
func newUserStorage() *userStorage {
	return &userStorage{
		users:        make(map[types.UserID]*user.User),
		slackIDIndex: make(map[string]map[string]types.UserID),
	}
}

// GetByID retrieves a user by their UUID
func (s *userStorage) GetByID(ctx context.Context, id types.UserID) (*user.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	u, exists := s.users[id]
	if !exists {
		return nil, goerr.New("user not found", goerr.V("user_id", id))
	}

	// Return a copy to prevent external modification
	userCopy := *u
	return &userCopy, nil
}

// GetBySlackIDAndTeamID retrieves a user by their Slack ID and Team ID
func (s *userStorage) GetBySlackIDAndTeamID(ctx context.Context, slackID, teamID string) (*user.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	teamMap, exists := s.slackIDIndex[slackID]
	if !exists {
		return nil, goerr.New("user not found", goerr.V("slack_id", slackID), goerr.V("team_id", teamID))
	}

	userID, exists := teamMap[teamID]
	if !exists {
		return nil, goerr.New("user not found", goerr.V("slack_id", slackID), goerr.V("team_id", teamID))
	}

	u, exists := s.users[userID]
	if !exists {
		return nil, goerr.New("user not found in storage", goerr.V("user_id", userID))
	}

	// Return a copy to prevent external modification
	userCopy := *u
	return &userCopy, nil
}

// Create creates a new user
func (s *userStorage) Create(ctx context.Context, u *user.User) error {
	if err := u.Validate(); err != nil {
		return goerr.Wrap(err, "invalid user")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user already exists
	if _, exists := s.users[u.ID]; exists {
		return goerr.New("user already exists", goerr.V("user_id", u.ID))
	}

	// Check if SlackID + TeamID combination already exists
	if teamMap, exists := s.slackIDIndex[u.SlackID]; exists {
		if _, exists := teamMap[u.TeamID]; exists {
			return goerr.New("user with same slack_id and team_id already exists", goerr.V("slack_id", u.SlackID), goerr.V("team_id", u.TeamID))
		}
	}

	// Create a copy to prevent external modification
	userCopy := *u
	s.users[u.ID] = &userCopy

	// Update index
	if s.slackIDIndex[u.SlackID] == nil {
		s.slackIDIndex[u.SlackID] = make(map[string]types.UserID)
	}
	s.slackIDIndex[u.SlackID][u.TeamID] = u.ID

	return nil
}

// Update updates an existing user
func (s *userStorage) Update(ctx context.Context, u *user.User) error {
	if err := u.Validate(); err != nil {
		return goerr.Wrap(err, "invalid user")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user exists
	existingUser, exists := s.users[u.ID]
	if !exists {
		return goerr.New("user not found", goerr.V("user_id", u.ID))
	}

	// If SlackID or TeamID changed, update index
	if existingUser.SlackID != u.SlackID || existingUser.TeamID != u.TeamID {
		// Remove old index
		if teamMap, exists := s.slackIDIndex[existingUser.SlackID]; exists {
			delete(teamMap, existingUser.TeamID)
			if len(teamMap) == 0 {
				delete(s.slackIDIndex, existingUser.SlackID)
			}
		}

		// Add new index
		if s.slackIDIndex[u.SlackID] == nil {
			s.slackIDIndex[u.SlackID] = make(map[string]types.UserID)
		}
		s.slackIDIndex[u.SlackID][u.TeamID] = u.ID
	}

	// Create a copy to prevent external modification
	userCopy := *u
	s.users[u.ID] = &userCopy

	return nil
}
