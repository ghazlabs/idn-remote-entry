package worker

import (
	"context"
	"encoding/json"
	"fmt"

	shcore "github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"github.com/ghazlabs/idn-remote-entry/internal/shared/rmq"
	shworker "github.com/ghazlabs/idn-remote-entry/internal/shared/worker"
	"github.com/ghazlabs/idn-remote-entry/internal/vacancy-worker/core"
	"github.com/wagslane/go-rabbitmq"
	"gopkg.in/validator.v2"
)

type Worker struct {
	Config
}

type Config struct {
	Service      core.Service   `validate:"nonnil"`
	RmqConsumer  *rmq.Consumer  `validate:"nonnil"`
	RmqPublisher *rmq.Publisher `validate:"nonnil"`
}

func New(cfg Config) (*Worker, error) {
	// validate config
	err := validator.Validate(cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &Worker{Config: cfg}, nil
}

// Run starts the worker and block until it's done.
func (w *Worker) Run() error {
	// run the consumer
	err := w.RmqConsumer.Run(func(d rabbitmq.Delivery) rabbitmq.Action {
		ctx := context.Background()
		// parse the message
		var req shcore.SubmitRequest
		err := json.Unmarshal(d.Body, &req)
		if err != nil {
			// discard the message if failed to parse
			return rabbitmq.NackDiscard
		}

		return shworker.HandleWithRetry(ctx, &req, w.RmqPublisher, func(ctx context.Context, msg shworker.RetryableMessage) error {
			return w.Service.Handle(ctx, *msg.(*shcore.SubmitRequest))
		})
	})
	if err != nil {
		return fmt.Errorf("failed to run rabbitmq consumer: %w", err)
	}

	return nil
}
