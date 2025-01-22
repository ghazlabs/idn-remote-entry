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
	svc      core.Service
	consumer *rmq.Consumer
}

type Config struct {
	Service            core.Service `validate:"nonnil"`
	QueueName          string       `validate:"nonzero"`
	RabbitMQConnString string       `validate:"nonzero"`
}

func New(cfg Config) (*Worker, error) {
	// validate config
	err := validator.Validate(cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// initialize consumer
	rmqConsumer, err := rmq.NewConsumer(rmq.ConsumerConfig{
		QueueName:          cfg.QueueName,
		RabbitMQConnString: cfg.RabbitMQConnString,
		Handler: func(d rabbitmq.Delivery) rabbitmq.Action {
			// parse the message
			var n shcore.WhatsappNotification
			err := json.Unmarshal(d.Body, &n)
			if err != nil {
				// discard the message if failed to parse
				return rabbitmq.NackDiscard
			}

			// handle the message
			log.Printf("handling notification %s\n", d.Body)
			err = cfg.Service.Handle(context.Background(), n)
			if err != nil {
				// output error and sleep for 1 second
				log.Printf("failed to handle notification %+v: %v, sleeping for 1 second...\n", n, err)
				time.Sleep(1 * time.Second)

				// requeue the message if failed to handle
				return rabbitmq.NackRequeue
			}

			return rabbitmq.Ack
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize rabbitmq consumer: %w", err)
	}

	return &Worker{
		svc:      cfg.Service,
		consumer: rmqConsumer,
	}, nil
}

// Run starts the worker and block until it's done.
func (w *Worker) Run() error {
	err := w.consumer.Run()
	if err != nil {
		return fmt.Errorf("failed to run consumer: %w", err)
	}
	w.consumer.Close()

	return nil
}
