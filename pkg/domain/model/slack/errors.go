package slack

import "errors"

var (
	// Thread errors
	ErrInvalidThreadID = errors.New("invalid thread ID")
	ErrEmptyTeamID     = errors.New("team ID is empty")
	ErrEmptyChannelID  = errors.New("channel ID is empty")
	ErrEmptyThreadTS   = errors.New("thread timestamp is empty")

	// Message errors
	ErrInvalidMessageID = errors.New("invalid message ID")
	ErrEmptyMessageID   = errors.New("message ID is empty")
	ErrEmptyUserID      = errors.New("user ID is empty")
	ErrEmptyText        = errors.New("message text is empty")
	ErrEmptyTimestamp   = errors.New("message timestamp is empty")

	// Repository errors
	ErrThreadNotFound = errors.New("thread not found")
)
