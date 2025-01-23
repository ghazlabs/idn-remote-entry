package queue

import (
	"context"
	"fmt"

	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"github.com/ghazlabs/idn-remote-entry/internal/shared/rmq"
)

type Queue struct {
	rmqPublisher *rmq.Publisher
}

func NewQueue(rmqPublisher *rmq.Publisher) *Queue {
	return &Queue{
		rmqPublisher: rmqPublisher,
	}
}

func (q *Queue) Put(ctx context.Context, req core.SubmitRequest) error {
	err := q.rmqPublisher.Publish(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to publish request to queue: %w", err)
	}

	return nil
}
