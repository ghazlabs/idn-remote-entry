package rmqutil

import (
	"fmt"

	"github.com/wagslane/go-rabbitmq"
	"gopkg.in/validator.v2"
)

type ConsumerConfig PublisherConfig

func NewConsumer(cfg ConsumerConfig) (*rabbitmq.Consumer, error) {
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

	return rmqConsumer, nil
}
