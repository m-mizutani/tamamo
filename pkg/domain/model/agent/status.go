package agent

import "github.com/m-mizutani/goerr/v2"

// Status represents the status of an agent
type Status string

const (
	StatusActive   Status = "active"
	StatusArchived Status = "archived"
)

// IsValid checks if the status is valid
func (s Status) IsValid() bool {
	switch s {
	case StatusActive, StatusArchived:
		return true
	default:
		return false
	}
}

// String returns the string representation of the status
func (s Status) String() string {
	return string(s)
}

// ValidateStatus validates an agent status
func ValidateStatus(status Status) error {
	if !status.IsValid() {
		return goerr.New("invalid agent status", goerr.V("status", status))
	}
	return nil
}
