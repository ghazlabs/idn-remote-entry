package rmq

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/wagslane/go-rabbitmq"
	"gopkg.in/validator.v2"
)

type Consumer struct {
	rmqConsumer *rabbitmq.Consumer
	queueName   string
}

type ConsumerConfig struct {
	QueueName          string `validate:"nonzero"`
	RabbitMQConnString string `validate:"nonzero"`
}

func NewConsumer(cfg ConsumerConfig) (*Consumer, error) {
	err := validator.Validate(cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	// initialize rabbitmq connection
	rmqConn, err := rabbitmq.NewConn(
		cfg.RabbitMQConnString,
		rabbitmq.WithConnectionOptionsLogging,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize rabbitmq connection: %w", err)
	}

	// initialize rabbitmq consumer
	rmqConsumer, err := rabbitmq.NewConsumer(
		rmqConn,
		cfg.QueueName,
		rabbitmq.WithConsumerOptionsRoutingKey(cfg.QueueName),
		rabbitmq.WithConsumerOptionsExchangeName(cfg.QueueName),
		rabbitmq.WithConsumerOptionsExchangeDeclare,
		rabbitmq.WithConsumerOptionsQueueDurable,
		rabbitmq.WithConsumerOptionsExchangeDurable,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize rabbitmq consumer: %w", err)
	}

	return &Consumer{
		rmqConsumer: rmqConsumer,
		queueName:   cfg.QueueName,
	}, nil
}

// Run starts the consumer and block until it's done.
func (c *Consumer) Run(h func(rabbitmq.Delivery) rabbitmq.Action) error {
	// define channel to receive shutdown signal
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGINT)
	done := make(chan bool)

	go func() {
		<-shutdownCh
		log.Println("shutting down worker...")
		done <- true
	}()

	// start consuming messages
	err := c.rmqConsumer.Run(h)
	if err != nil {
		return fmt.Errorf("failed to start consuming messages: %w", err)
	}

	// wait for cleanup to finish
	<-done

	return nil
}

func (c *Consumer) Close() {
	c.rmqConsumer.Close()
}
