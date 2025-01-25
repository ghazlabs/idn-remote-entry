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
	rmqConsumer  *rabbitmq.Consumer
	rmqPublisher *rabbitmq.Publisher
	queueName    string
	maxRetry     int
}

type ConsumerConfig struct {
	QueueName          string `validate:"nonzero"`
	RabbitMQConnString string `validate:"nonzero"`
	MaxMessageRetry    int    `validate:"min=0"`
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

	// initialize rabbitmq publisher for retry mechanism
	rmqPub, err := rabbitmq.NewPublisher(
		rmqConn,
		rabbitmq.WithPublisherOptionsLogging,
		rabbitmq.WithPublisherOptionsExchangeName(cfg.QueueName),
		rabbitmq.WithPublisherOptionsExchangeDeclare,
		rabbitmq.WithPublisherOptionsExchangeDurable,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize rabbitmq publisher: %w", err)
	}

	return &Consumer{
		rmqConsumer:  rmqConsumer,
		queueName:    cfg.QueueName,
		maxRetry:     cfg.MaxMessageRetry,
		rmqPublisher: rmqPub,
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
	err := c.rmqConsumer.Run(func(d rabbitmq.Delivery) (action rabbitmq.Action) {
		// execute handler
		act := h(d)

		// handle retry mechanism
		if act == rabbitmq.NackRequeue {
			// check if message has been retried too many times
			retryCount := d.Headers[retryHeaderKey].(int)
			if retryCount >= c.maxRetry {
				log.Printf("max retry reached for message %s\n", d.Body)
				return rabbitmq.NackDiscard
			}

			// increment retry count
			retryCount++
			headers := map[string]interface{}(d.Headers)
			headers[retryHeaderKey] = retryCount

			// requeue the message
			err := c.rmqPublisher.Publish(
				d.Body,
				[]string{c.queueName},
				rabbitmq.WithPublishOptionsContentType(d.ContentType),
				rabbitmq.WithPublishOptionsExchange(c.queueName),
				rabbitmq.WithPublishOptionsHeaders(headers),
				rabbitmq.WithPublishOptionsPersistentDelivery,
			)
			if err != nil {
				log.Printf("failed to requeue message %s: %v\n", d.Body, err)
				return rabbitmq.NackRequeue
			}

			// acknowledge the current failed message to remove it from queue
			// since the new message has been requeued
			return rabbitmq.Ack
		}

		return act
	})
	if err != nil {
		return fmt.Errorf("failed to start consuming messages: %w", err)
	}

	// wait for cleanup to finish
	<-done

	return nil
}

func (c *Consumer) Close() {
	c.rmqPublisher.Close()
	c.rmqConsumer.Close()
}

const retryHeaderKey = "x-retry-count"
