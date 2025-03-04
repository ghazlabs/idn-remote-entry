package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	shcore "github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"github.com/ghazlabs/idn-remote-entry/internal/shared/rmq"
	"github.com/ghazlabs/idn-remote-entry/internal/testutil"
	"github.com/riandyrn/go-env"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockService struct {
	mock.Mock
}

func (m *mockService) Handle(ctx context.Context, req shcore.SubmitRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

type inValidMsg struct{}

func (i inValidMsg) ToJSON() []byte {
	return []byte("invalid")
}
func TestWorkerRun(t *testing.T) {
	tests := []struct {
		name               string
		setupMocks         func(*mockService)
		message            rmq.Message
		handlerCalledCount int
	}{
		{
			name:               "empty msg",
			setupMocks:         func(s *mockService) {},
			message:            inValidMsg{},
			handlerCalledCount: 0,
		},
		{
			name: "successful message handling",
			setupMocks: func(s *mockService) {
				// this means handle will be called 1 time
				s.On("Handle", mock.Anything, mock.AnythingOfType("SubmitRequest")).Return(nil)
			},
			message: shcore.SubmitRequest{
				Vacancy: shcore.Vacancy{JobTitle: "Software Engineer"},
			},
			handlerCalledCount: 1,
		},
		{
			name: "retry on error",
			setupMocks: func(s *mockService) {
				// this means handle will be called 2 times
				s.On("Handle", mock.Anything, mock.AnythingOfType("SubmitRequest")).Return(errors.New("test error"))
				s.On("Handle", mock.Anything, mock.AnythingOfType("SubmitRequest")).Return(nil)
			},
			message: shcore.SubmitRequest{
				Vacancy: shcore.Vacancy{JobTitle: "Software Engineer"},
			},
			handlerCalledCount: 2,
		},
		{
			name: "discard after max retries",
			setupMocks: func(s *mockService) {
				// this means handle will be called 3 times
				s.On("Handle", mock.Anything, mock.AnythingOfType("SubmitRequest")).Return(errors.New("test error"))
				s.On("Handle", mock.Anything, mock.AnythingOfType("SubmitRequest")).Return(errors.New("test error"))
				s.On("Handle", mock.Anything, mock.AnythingOfType("SubmitRequest")).Return(errors.New("test error"))
			},
			message: shcore.SubmitRequest{
				Vacancy: shcore.Vacancy{JobTitle: "Software Engineer"},
			},
			handlerCalledCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create context with timeout
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Create mocks
			mockService := &mockService{}

			// Setup mocks
			tt.setupMocks(mockService)

			mqConn := env.GetString(testutil.EnvKeyRabbitMQConn)
			queueName := env.GetString(testutil.EnvKeyRabbitMQVacancyQueueName)
			consumer, err := rmq.NewConsumer(rmq.ConsumerConfig{
				RabbitMQConnString: mqConn,
				QueueName:          queueName,
			})
			require.NoError(t, err)
			defer consumer.Close()

			publisher, err := rmq.NewPublisher(rmq.PublisherConfig{
				QueueName:          queueName,
				RabbitMQConnString: mqConn,
			})
			require.NoError(t, err)
			defer publisher.Close()

			// Create worker
			w := &Worker{
				Config: Config{
					Service:      mockService,
					RmqConsumer:  consumer,
					RmqPublisher: publisher,
				},
			}

			// Channel to signal worker completion
			done := make(chan error)

			// Start worker in goroutine
			go func() {
				done <- w.Run()
			}()

			// Publish test message
			err = publisher.Publish(ctx, tt.message)
			assert.NoError(t, err)

			// Wait for either context timeout or expected calls
			for {
				select {
				case err := <-done:
					t.Fatalf("worker stopped unexpectedly: %v", err)
				case <-ctx.Done():
					mockService.AssertExpectations(t)
					return
				case <-time.After(100 * time.Millisecond):
					continue
				}
			}
		})
	}
}
