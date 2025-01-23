package core

import (
	"context"

	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
)

type Queue interface {
	Put(ctx context.Context, req core.SubmitRequest) error
}
