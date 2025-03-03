package core

import (
	"context"

	shcore "github.com/ghazlabs/idn-remote-entry/internal/shared/core"
)

type Publisher interface {
	Publish(ctx context.Context, n shcore.Notification) error
}
