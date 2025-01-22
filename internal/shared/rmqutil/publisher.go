package rmqutil

import (
	"fmt"

	"github.com/wagslane/go-rabbitmq"
	"gopkg.in/validator.v2"
)

type PublisherConfig struct {
	QueueName          string `validate:"nonzero"`
	RabbitMQConnString string `validate:"nonzero"`
}

func NewPublisher(cfg PublisherConfig) (*rabbitmq.Publisher, error) {
	// validate config
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

	// initialize rabbitmq publisher
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

	return rmqPub, nil
}
