package notifier

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"github.com/ghazlabs/idn-remote-entry/internal/shared/rmq"
	"gopkg.in/validator.v2"
)

type WaNotifier struct {
	WaNotifierConfig
}

func NewWaNotifier(cfg WaNotifierConfig) (*WaNotifier, error) {
	// validate config
	err := validator.Validate(cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return &WaNotifier{
		WaNotifierConfig: cfg,
	}, nil
}

type WaNotifierConfig struct {
	RmqPublisher   *rmq.Publisher `validate:"nonnil"`
	WaRecipientIDs []string       `validate:"nonzero"`
}

func (n *WaNotifier) Notify(ctx context.Context, v core.VacancyRecord) error {
	for _, waID := range n.WaRecipientIDs {
		ntf := core.WhatsappNotification{
			RecipientID:   waID,
			VacancyRecord: v,
		}
		data, _ := json.Marshal(ntf)
		err := n.RmqPublisher.Publish(ctx, rmq.PublishParams{
			ContentType: "application/json",
			Data:        data,
		})
		if err != nil {
			return fmt.Errorf("failed to publish notification %+v: %w", ntf, err)
		}
	}

	return nil
}
