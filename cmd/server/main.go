package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ghazlabs/idn-remote-entry/internal/server/core"
	"github.com/ghazlabs/idn-remote-entry/internal/server/driven/queue"
	"github.com/ghazlabs/idn-remote-entry/internal/server/driver"
	"github.com/ghazlabs/idn-remote-entry/internal/shared/rmq"
	"github.com/riandyrn/go-env"
)

const (
	envKeyListenPort               = "LISTEN_PORT"
	envKeyClientApiKey             = "CLIENT_API_KEY"
	envKeyRabbitMQConn             = "RABBITMQ_CONN"
	envKeyRabbitMQVacancyQueueName = "RABBITMQ_VACANCY_QUEUE_NAME"
)

func main() {
	// initialize rmq publisher
	rmqPub, err := rmq.NewPublisher(rmq.PublisherConfig{
		QueueName:          env.GetString(envKeyRabbitMQVacancyQueueName),
		RabbitMQConnString: env.GetString(envKeyRabbitMQConn),
	})
	if err != nil {
		log.Fatalf("failed to initialize rmq publisher: %v", err)
	}

	// initialize service
	svc, err := core.NewService(core.ServiceConfig{
		Queue: queue.NewQueue(rmqPub),
	})
	if err != nil {
		log.Fatalf("failed to initialize service: %v", err)
	}

	// initialize handler
	api, err := driver.NewAPI(driver.APIConfig{
		Service:      svc,
		ClientApiKey: env.GetString(envKeyClientApiKey),
	})
	if err != nil {
		log.Fatalf("failed to initialize API: %v", err)
	}

	// initialize server
	listenAddr := fmt.Sprintf(":%s", env.GetString(envKeyListenPort))
	s := &http.Server{
		Addr:        listenAddr,
		Handler:     api.GetHandler(),
		ReadTimeout: 5 * time.Second,
	}

	// start server
	log.Printf("server is listening on %s", listenAddr)
	if err := s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("failed to start server: %v", err)
	}
}
