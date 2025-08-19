package types

import (
	"context"

	"github.com/google/uuid"
	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/utils/errors"
)

type UUID string

func NewUUID(ctx context.Context) UUID {
	return UUID(newUUID(ctx))
}

func (id UUID) String() string {
	return string(id)
}

// IsValid checks if the UUID is valid
func (id UUID) IsValid() bool {
	if id == "" {
		return false
	}
	_, err := uuid.Parse(string(id))
	return err == nil
}

func newUUID(ctx context.Context) string {
	id, err := uuid.NewV7()
	if err != nil {
		errors.Handle(ctx, goerr.Wrap(err, "failed to generate uuid V7, fallback to V4"))
		return uuid.New().String()
	}

	return id.String()
}
