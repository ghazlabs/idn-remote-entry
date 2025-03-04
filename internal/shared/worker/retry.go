package worker

import (
	"context"
	"log"
	"time"

	"github.com/ghazlabs/idn-remote-entry/internal/shared/rmq"
	"github.com/wagslane/go-rabbitmq"
)

// RetryableMessage represents a message that can be retried
// Note: All implementations must use pointer receivers
type RetryableMessage interface {
	GetRetries() int
	IncreaseRetries()
	ToJSON() []byte
}

// HandleWithRetry handles a message with retry mechanism
func HandleWithRetry(
	ctx context.Context,
	msg RetryableMessage,
	publisher *rmq.Publisher,
	handler func(context.Context, RetryableMessage) error,
) rabbitmq.Action {
	// handle the message
	oldRetries := msg.GetRetries()
	log.Printf("handling vacancy (attempt %d) %s\n", oldRetries+1, msg.ToJSON())
	startTime := time.Now()
	defer func() {
		log.Printf("handled vacancy %s in %v\n", msg.ToJSON(), time.Since(startTime))
	}()

	err := handler(ctx, msg)
	if err != nil {
		// output error and sleep for 1 second
		log.Printf("failed to handle message (attempt %d): %v, sleeping for 1 second...\n", oldRetries+1, err)
		time.Sleep(1 * time.Second)

		// increment retry count
		msg.IncreaseRetries()
		if msg.GetRetries() >= 3 {
			log.Printf("discarding message after %d retries\n", msg.GetRetries())
			return rabbitmq.NackDiscard
		}

		// publish new message
		publisher.Publish(ctx, msg)

		// discard the original message since we've published new one
		return rabbitmq.NackDiscard
	}

	return rabbitmq.Ack
}
