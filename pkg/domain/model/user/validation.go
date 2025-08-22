package user

import (
	"github.com/m-mizutani/goerr/v2"
)

// Validate validates the user entity
func (u *User) Validate() error {
	if !u.ID.IsValid() {
		return goerr.New("invalid user ID")
	}
	if u.SlackID == "" {
		return goerr.New("slack ID is required")
	}
	if u.TeamID == "" {
		return goerr.New("team ID is required")
	}
	return nil
}
