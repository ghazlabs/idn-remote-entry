package main

import (
	"log"

	"github.com/ghazlabs/idn-remote-entry/internal/wa-worker/core"
	"github.com/ghazlabs/idn-remote-entry/internal/wa-worker/driven/publisher"
	"github.com/ghazlabs/idn-remote-entry/internal/wa-worker/driver/worker"
	"github.com/go-resty/resty/v2"
	"github.com/riandyrn/go-env"
	"github.com/wagslane/go-rabbitmq"
)

const (
	envKeyWhatsappApiUser     = "WHATSAPP_API_USER"
	envKeyWhatsappApiPass     = "WHATSAPP_API_PASS"
	envKeyWhatsappApiBaseUrl  = "WHATSAPP_API_BASE_URL"
	envKeyRabbitMQConn        = "RABBITMQ_CONN"
	envKeyRabbitMQWaQueueName = "RABBITMQ_WA_QUEUE_NAME"
)

func main() {
	// initialize publisher
	pub, err := publisher.NewWaPublisher(publisher.WaPublisherConfig{
		HttpClient:   resty.New(),
		Username:     env.GetString(envKeyWhatsappApiUser),
		Password:     env.GetString(envKeyWhatsappApiPass),
		WaApiBaseUrl: env.GetString(envKeyWhatsappApiBaseUrl),
	})
	if err != nil {
		log.Fatalf("failed to initialize publisher: %v", err)
	}

	// initialize service
	svc, err := core.NewService(core.ServiceConfig{
		Publisher: pub,
	})
	if err != nil {
		log.Fatalf("failed to initialize service: %v", err)
	}

	// initialize rabbitmq connection
	rmqConn, err := rabbitmq.NewConn(
		env.GetString(envKeyRabbitMQConn),
		rabbitmq.WithConnectionOptionsLogging,
	)
	if err != nil {
		log.Fatalf("failed to initialize rabbitmq connection: %v", err)
	}

	// initialize rabbitmq consumer
	queueName := env.GetString(envKeyRabbitMQWaQueueName)
	rmqConsumer, err := rabbitmq.NewConsumer(
		rmqConn,
		queueName,
		rabbitmq.WithConsumerOptionsRoutingKey(queueName),
		rabbitmq.WithConsumerOptionsExchangeName(queueName),
		rabbitmq.WithConsumerOptionsExchangeDeclare,
	)
	if err != nil {
		log.Fatalf("failed to initialize rabbitmq consumer: %v", err)
	}

	// initialize worker
	w, err := worker.New(worker.Config{
		Service:     svc,
		RmqConsumer: rmqConsumer,
	})
	if err != nil {
		log.Fatalf("failed to initialize worker: %v", err)
	}

	// run worker
	err = w.Run()
	if err != nil {
		log.Fatalf("failed to run worker: %v", err)
	}
}
