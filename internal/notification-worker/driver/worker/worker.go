package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/ghazlabs/idn-remote-entry/internal/notification-worker/core"
	shcore "github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"github.com/ghazlabs/idn-remote-entry/internal/shared/rmq"
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
		var n shcore.Notification
		err := json.Unmarshal(d.Body, &n)
		if err != nil {
			// discard the message if failed to parse
			return rabbitmq.NackDiscard
		}

		// handle the message
		log.Printf("handling notification (attempt %d) %s\n", n.Retries+1, d.Body)
		startTime := time.Now()
		err = w.Config.Service.Handle(ctx, n)
		if err != nil {
			// output error and sleep for 1 second
			log.Printf("failed to handle notification (attempt %d) %s: %s, sleeping for 1 second...\n", n.Retries+1, n.ToJSON(), err)
			time.Sleep(1 * time.Second)

			// increment retry count
			n.Retries++
			if n.Retries >= 3 {
				log.Printf("discarding notification after %d retries: %s\n", n.Retries, n.ToJSON())
				return rabbitmq.NackDiscard
			}

			// publish new request
			w.Config.RmqPublisher.Publish(ctx, n)

			// discard the original message since we've published new one
			return rabbitmq.NackDiscard
		}

		defer func() {
			log.Printf("handled notification %s in %s\n", d.Body, time.Since(startTime))
		}()

		return rabbitmq.Ack
	})
	if err != nil {
		return fmt.Errorf("failed to run rabbitmq consumer: %w", err)
	}

	return nil
}
