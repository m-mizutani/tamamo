package usecase

import (
	"context"
	"time"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/user"
	"github.com/m-mizutani/tamamo/pkg/domain/types"
)

type UserUseCase struct {
	userRepo      interfaces.UserRepository
	avatarService interfaces.UserAvatarService
	slackClient   interfaces.SlackClient
}

func NewUserUseCase(userRepo interfaces.UserRepository, avatarService interfaces.UserAvatarService, slackClient interfaces.SlackClient) *UserUseCase {
	return &UserUseCase{
		userRepo:      userRepo,
		avatarService: avatarService,
		slackClient:   slackClient,
	}
}

// GetOrCreateUser gets an existing user or creates a new one for Slack OAuth
func (uc *UserUseCase) GetOrCreateUser(ctx context.Context, slackID, slackName, email, teamID string) (*user.User, error) {
	// Get complete profile information from Slack
	profile, err := uc.slackClient.GetUserProfile(ctx, slackID)
	if err != nil {
		// If we can't get the profile, fall back to the provided slackName as display name
		// This is a warning-level log since we can continue with fallback
		_ = goerr.Wrap(err, "failed to get user profile from Slack, using fallback", goerr.V("slack_id", slackID))
	}

	displayName := slackName // fallback
	if profile != nil && profile.DisplayName != "" {
		displayName = profile.DisplayName
	}

	// Try to get existing user first
	existingUser, err := uc.userRepo.GetBySlackIDAndTeamID(ctx, slackID, teamID)
	if err == nil {
		// User exists, update their information if needed
		needsUpdate := existingUser.SlackName != slackName ||
			existingUser.DisplayName != displayName ||
			existingUser.Email != email

		if needsUpdate {
			existingUser.UpdateSlackInfo(slackName, displayName, email)
			if err := uc.userRepo.Update(ctx, existingUser); err != nil {
				return nil, goerr.Wrap(err, "failed to update existing user", goerr.V("slack_id", slackID), goerr.V("team_id", teamID))
			}
		}
		return existingUser, nil
	}

	// User doesn't exist, create a new one
	newUser := &user.User{
		ID:          types.NewUserID(ctx),
		SlackID:     slackID,
		SlackName:   slackName,
		DisplayName: displayName,
		Email:       email,
		TeamID:      teamID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := uc.userRepo.Create(ctx, newUser); err != nil {
		return nil, goerr.Wrap(err, "failed to create new user", goerr.V("slack_id", slackID), goerr.V("team_id", teamID))
	}

	return newUser, nil
}

// GetUserByID retrieves a user by their ID
func (uc *UserUseCase) GetUserByID(ctx context.Context, userID types.UserID) (*user.User, error) {
	u, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get user", goerr.V("user_id", userID))
	}
	return u, nil
}

// UpdateUser updates user information
func (uc *UserUseCase) UpdateUser(ctx context.Context, u *user.User) error {
	u.UpdatedAt = time.Now()
	if err := uc.userRepo.Update(ctx, u); err != nil {
		return goerr.Wrap(err, "failed to update user", goerr.V("user_id", u.ID))
	}
	return nil
}

// GetUserAvatar retrieves avatar data for a user
func (uc *UserUseCase) GetUserAvatar(ctx context.Context, userID types.UserID, size int) ([]byte, error) {
	u, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get user", goerr.V("user_id", userID))
	}

	avatarData, err := uc.avatarService.GetAvatarData(ctx, u.SlackID, size)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get user avatar", goerr.V("user_id", userID), goerr.V("slack_id", u.SlackID))
	}

	return avatarData, nil
}

// InvalidateUserAvatarCache invalidates the avatar cache for a user
func (uc *UserUseCase) InvalidateUserAvatarCache(ctx context.Context, userID types.UserID) error {
	u, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		return goerr.Wrap(err, "failed to get user", goerr.V("user_id", userID))
	}

	if err := uc.avatarService.InvalidateCache(ctx, u.SlackID); err != nil {
		return goerr.Wrap(err, "failed to invalidate avatar cache", goerr.V("user_id", userID), goerr.V("slack_id", u.SlackID))
	}

	return nil
}
