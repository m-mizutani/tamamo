package slack

import (
	"context"
	"sync"
	"time"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/domain/interfaces"
	"github.com/m-mizutani/tamamo/pkg/domain/model/slack"
)

// ChannelCache provides cached access to channel information with TTL
type ChannelCache struct {
	mu     sync.RWMutex
	cache  map[string]*cachedChannelInfo
	client interfaces.SlackClient
	ttl    time.Duration
}

// cachedChannelInfo holds channel information with timestamp
type cachedChannelInfo struct {
	info      *slack.ChannelInfo
	timestamp time.Time
}

// NewChannelCache creates a new channel cache with the specified TTL
func NewChannelCache(client interfaces.SlackClient, ttl time.Duration) *ChannelCache {
	if ttl <= 0 {
		ttl = time.Hour // Default to 1 hour TTL
	}

	return &ChannelCache{
		cache:  make(map[string]*cachedChannelInfo),
		client: client,
		ttl:    ttl,
	}
}

// GetChannelInfo retrieves channel information with caching
func (c *ChannelCache) GetChannelInfo(ctx context.Context, channelID string) (*slack.ChannelInfo, error) {
	if channelID == "" {
		return nil, goerr.New("channelID cannot be empty")
	}

	// Try to get from cache first
	if info := c.getCachedInfo(channelID); info != nil {
		return info, nil
	}

	// Cache miss - fetch from Slack API
	info, err := c.client.GetChannelInfo(ctx, channelID)
	if err != nil {
		return nil, goerr.Wrap(err, "failed to get channel info from Slack client",
			goerr.V("channel_id", channelID))
	}

	// Store in cache
	c.setCachedInfo(channelID, info)

	return info, nil
}

// getCachedInfo retrieves channel info from cache if not expired
func (c *ChannelCache) getCachedInfo(channelID string) *slack.ChannelInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cached, exists := c.cache[channelID]
	if !exists {
		return nil
	}

	// Check if cached entry has expired
	if time.Since(cached.timestamp) > c.ttl {
		// Entry expired, will be cleaned up later
		return nil
	}

	// Return a copy to avoid external modifications
	infoCopy := *cached.info
	return &infoCopy
}

// setCachedInfo stores channel info in cache
func (c *ChannelCache) setCachedInfo(channelID string, info *slack.ChannelInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Store a copy to avoid external modifications
	infoCopy := *info
	infoCopy.UpdatedAt = time.Now()

	c.cache[channelID] = &cachedChannelInfo{
		info:      &infoCopy,
		timestamp: time.Now(),
	}
}

// InvalidateChannel removes a specific channel from cache
func (c *ChannelCache) InvalidateChannel(channelID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.cache, channelID)
}

// InvalidateAll clears all cached entries
func (c *ChannelCache) InvalidateAll() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*cachedChannelInfo)
}

// CleanExpiredEntries removes expired entries from cache
func (c *ChannelCache) CleanExpiredEntries() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for channelID, cached := range c.cache {
		if now.Sub(cached.timestamp) > c.ttl {
			delete(c.cache, channelID)
		}
	}
}

// GetCacheStats returns cache statistics for monitoring
func (c *ChannelCache) GetCacheStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now()
	var validEntries, expiredEntries int

	for _, cached := range c.cache {
		if now.Sub(cached.timestamp) > c.ttl {
			expiredEntries++
		} else {
			validEntries++
		}
	}

	return CacheStats{
		TotalEntries:   len(c.cache),
		ValidEntries:   validEntries,
		ExpiredEntries: expiredEntries,
		TTL:            c.ttl,
	}
}

// CacheStats holds cache statistics
type CacheStats struct {
	TotalEntries   int
	ValidEntries   int
	ExpiredEntries int
	TTL            time.Duration
}

// StartCleanupWorker starts a background goroutine to periodically clean up expired entries
func (c *ChannelCache) StartCleanupWorker(ctx context.Context, interval time.Duration) {
	if interval <= 0 {
		interval = 30 * time.Minute // Default cleanup interval
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				c.CleanExpiredEntries()
			}
		}
	}()
}
