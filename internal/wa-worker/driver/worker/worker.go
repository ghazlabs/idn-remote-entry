package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	shcore "github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"github.com/ghazlabs/idn-remote-entry/internal/wa-worker/core"
	"github.com/wagslane/go-rabbitmq"
	"gopkg.in/validator.v2"
)

type Worker struct {
	Config
}

type Config struct {
	Service     core.Service       `validate:"nonnil"`
	RmqConsumer *rabbitmq.Consumer `validate:"nonnil"`
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
	// define channel to receive shutdown signal
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)
	done := make(chan bool)

	go func() {
		<-shutdownCh
		done <- true
	}()

	// start consuming messages
	err := w.RmqConsumer.Run(func(d rabbitmq.Delivery) rabbitmq.Action {
		// parse the message
		var n shcore.WhatsappNotification
		err := json.Unmarshal(d.Body, &n)
		if err != nil {
			// discard the message if failed to parse
			return rabbitmq.NackDiscard
		}

		// handle the message
		err = w.Service.Handle(context.Background(), n)
		if err != nil {
			// requeue the message if failed to handle
			return rabbitmq.NackRequeue
		}

		return rabbitmq.Ack
	})
	if err != nil {
		return fmt.Errorf("failed to start consuming messages: %w", err)
	}

	// wait for cleanup to finish
	<-done

	return nil
}
