package safe

import (
	"context"
	"io"

	"github.com/m-mizutani/goerr/v2"
	"github.com/m-mizutani/tamamo/pkg/utils/errors"
)

func Close(ctx context.Context, c io.Closer) {
	if err := c.Close(); err != nil {
		errors.Handle(ctx, goerr.Wrap(err, "failed to close by safe.Close"))
	}
}

func Write(ctx context.Context, w io.Writer, data []byte) {
	if _, err := w.Write(data); err != nil {
		errors.Handle(ctx, goerr.Wrap(err, "failed to write by safe.Write"))
	}
}
