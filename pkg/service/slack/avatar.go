package slack

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
)

// UserInfo represents user information from Slack API
type UserInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Profile struct {
		Image24   string `json:"image_24"`
		Image32   string `json:"image_32"`
		Image48   string `json:"image_48"`
		Image72   string `json:"image_72"`
		Image192  string `json:"image_192"`
		Image512  string `json:"image_512"`
		ImageOrig string `json:"image_original"`
	} `json:"profile"`
}

// avatarCacheEntry represents a cached avatar entry
type avatarCacheEntry struct {
	data      []byte
	timestamp time.Time
}

// AvatarService implements UserAvatarService interface
type AvatarService struct {
	slackClient interfaces.SlackClient
	httpClient  *http.Client
	cache       map[string]*avatarCacheEntry
	cacheMutex  sync.RWMutex
	cacheTTL    time.Duration
}

// NewAvatarService creates a new avatar service
func NewAvatarService(slackClient interfaces.SlackClient) *AvatarService {
	return &AvatarService{
		slackClient: slackClient,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		cache:       make(map[string]*avatarCacheEntry),
		cacheTTL:    1 * time.Hour,
	}
}

// GetAvatarData retrieves avatar data for the specified user and size
func (s *AvatarService) GetAvatarData(ctx context.Context, slackID string, size int) ([]byte, error) {
	cacheKey := fmt.Sprintf("%s:%d", slackID, size)

	// Check cache first
	s.cacheMutex.RLock()
	if entry, exists := s.cache[cacheKey]; exists {
		if time.Since(entry.timestamp) < s.cacheTTL {
			s.cacheMutex.RUnlock()
			return entry.data, nil
		}
	}
	s.cacheMutex.RUnlock()

	// Get user info from Slack API
	userInfo, err := s.getUserInfo(ctx, slackID)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get user info from Slack")
	}

	// Select appropriate avatar URL based on size
	avatarURL := s.selectAvatarURL(userInfo, size)
	if avatarURL == "" {
		return nil, goerr.New("no avatar URL available")
	}

	// Fetch image data
	avatarData, err := s.fetchImage(ctx, avatarURL)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to fetch avatar image")
	}

	// Cache the result
	s.cacheMutex.Lock()
	s.cache[cacheKey] = &avatarCacheEntry{
		data:      avatarData,
		timestamp: time.Now(),
	}
	s.cacheMutex.Unlock()

	return avatarData, nil
}

// InvalidateCache removes all cached entries for the specified user
func (s *AvatarService) InvalidateCache(ctx context.Context, slackID string) error {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	// Remove all cache entries for this user
	keysToDelete := make([]string, 0)
	for key := range s.cache {
		if len(key) > len(slackID) && key[:len(slackID)] == slackID && key[len(slackID)] == ':' {
			keysToDelete = append(keysToDelete, key)
		}
	}

	for _, key := range keysToDelete {
		delete(s.cache, key)
	}

	return nil
}

// selectAvatarURL selects the most appropriate avatar URL based on requested size
func (s *AvatarService) selectAvatarURL(userInfo *UserInfo, size int) string {
	switch {
	case size <= 24:
		return userInfo.Profile.Image24
	case size <= 32:
		return userInfo.Profile.Image32
	case size <= 48:
		return userInfo.Profile.Image48
	case size <= 72:
		return userInfo.Profile.Image72
	case size <= 192:
		return userInfo.Profile.Image192
	case size <= 512:
		return userInfo.Profile.Image512
	default:
		if userInfo.Profile.ImageOrig != "" {
			return userInfo.Profile.ImageOrig
		}
		return userInfo.Profile.Image512
	}
}

// fetchImage downloads image data from the given URL
func (s *AvatarService) fetchImage(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to create HTTP request")
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to fetch image")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, goerr.New("unexpected status code", goerr.V("status_code", resp.StatusCode), goerr.V("url", url))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to read response body")
	}

	return data, nil
}

// getUserInfo retrieves user information from Slack API
func (s *AvatarService) getUserInfo(ctx context.Context, slackID string) (*UserInfo, error) {
	userProfile, err := s.slackClient.GetUserProfile(ctx, slackID)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get user profile from Slack client")
	}

	return &UserInfo{
		ID:    userProfile.ID,
		Name:  userProfile.Name,
		Email: userProfile.Email,
		Profile: struct {
			Image24   string `json:"image_24"`
			Image32   string `json:"image_32"`
			Image48   string `json:"image_48"`
			Image72   string `json:"image_72"`
			Image192  string `json:"image_192"`
			Image512  string `json:"image_512"`
			ImageOrig string `json:"image_original"`
		}{
			Image24:   userProfile.Profile.Image24,
			Image32:   userProfile.Profile.Image32,
			Image48:   userProfile.Profile.Image48,
			Image72:   userProfile.Profile.Image72,
			Image192:  userProfile.Profile.Image192,
			Image512:  userProfile.Profile.Image512,
			ImageOrig: userProfile.Profile.ImageOrig,
		},
	}, nil
}
