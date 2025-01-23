package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	shcore "github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"github.com/ghazlabs/idn-remote-entry/internal/shared/rmq"
	"github.com/ghazlabs/idn-remote-entry/internal/wa-worker/core"
	"github.com/wagslane/go-rabbitmq"
	"gopkg.in/validator.v2"
)

type Worker struct {
	Config
}

type Config struct {
	Service     core.Service  `validate:"nonnil"`
	RmqConsumer *rmq.Consumer `validate:"nonnil"`
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
		// parse the message
		var n shcore.WaNotification
		err := json.Unmarshal(d.Body, &n)
		if err != nil {
			// discard the message if failed to parse
			return rabbitmq.NackDiscard
		}

		// handle the message
		log.Printf("handling notification %s\n", d.Body)
		err = w.Config.Service.Handle(context.Background(), n)
		if err != nil {
			// output error and sleep for 1 second
			log.Printf("failed to handle notification %+v: %v, sleeping for 1 second...\n", n, err)
			time.Sleep(1 * time.Second)

			// requeue the message if failed to handle
			return rabbitmq.NackRequeue
		}

		return rabbitmq.Ack
	})
	if err != nil {
		return fmt.Errorf("failed to run rabbitmq consumer: %w", err)
	}

	return nil
}
