package notifier

import (
	"context"
	"fmt"

	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"github.com/ghazlabs/idn-remote-entry/internal/shared/rmq"
	"gopkg.in/validator.v2"
)

type Notifier struct {
	NotifierConfig
}

func NewNotifier(cfg NotifierConfig) (*Notifier, error) {
	// validate config
	err := validator.Validate(cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return &Notifier{
		NotifierConfig: cfg,
	}, nil
}

type NotifierConfig struct {
	RmqPublisher   *rmq.Publisher `validate:"nonnil"`
	WaRecipientIDs []string       `validate:"nonzero"`
}

func (n *Notifier) Notify(ctx context.Context, v core.VacancyRecord) error {
	for _, waID := range n.WaRecipientIDs {
		ntf := core.Notification{
			RecipientID:   waID,
			VacancyRecord: v,
		}
		err := n.RmqPublisher.Publish(ctx, ntf)
		if err != nil {
			return fmt.Errorf("failed to publish notification %+v: %w", ntf, err)
		}
	}

	return nil
}
