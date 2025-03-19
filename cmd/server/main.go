package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ghazlabs/idn-remote-entry/internal/server/core"
	"github.com/ghazlabs/idn-remote-entry/internal/server/driven/approval"
	"github.com/ghazlabs/idn-remote-entry/internal/server/driven/email"
	"github.com/ghazlabs/idn-remote-entry/internal/server/driven/queue"
	"github.com/ghazlabs/idn-remote-entry/internal/server/driven/token"
	"github.com/ghazlabs/idn-remote-entry/internal/server/driver"
	"github.com/ghazlabs/idn-remote-entry/internal/shared/rmq"
	"github.com/riandyrn/go-env"
)

const (
	envKeyServerDomain             = "SERVER_DOMAIN"
	envKeyListenPort               = "LISTEN_PORT"
	envKeyClientApiKey             = "CLIENT_API_KEY"
	envKeyRabbitMQConn             = "RABBITMQ_CONN"
	envKeyRabbitMQVacancyQueueName = "RABBITMQ_VACANCY_QUEUE_NAME"
	envKeyApprovedSubmitterEmails  = "APPROVED_SUBMITTER_EMAILS"
	envKeyAdminEmails              = "ADMIN_EMAILS"
	envKeySmtpHost                 = "SMTP_HOST"
	envKeySmtpPort                 = "SMTP_PORT"
	envKeySmtpFrom                 = "SMTP_FROM"
	envKeySmtpPassword             = "SMTP_PASS"
	envKeyApprovalJwtSecretKey     = "APPROVAL_JWT_SECRET_KEY"
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

	emailClient, err := email.NewEmail(email.EmailConfig{
		Host:         env.GetString(envKeySmtpHost),
		Port:         env.GetInt(envKeySmtpPort),
		From:         env.GetString(envKeySmtpFrom),
		Password:     env.GetString(envKeySmtpPassword),
		ServerDomain: env.GetString(envKeyServerDomain),
		AdminEmails:  env.GetString(envKeyAdminEmails),
	})
	if err != nil {
		log.Fatalf("failed to initialize email client: %v", err)
	}

	tokenizer, err := token.NewTokenizer(token.TokenizerConfig{
		SecretKey: env.GetString(envKeyApprovalJwtSecretKey),
	})
	if err != nil {
		log.Fatalf("failed to initialize tokenizer: %v", err)
	}

	approval, err := approval.NewApproval(approval.ApprovalConfig{
		ApprovedSubmitterEmails: env.GetString(envKeyApprovedSubmitterEmails),
	})
	if err != nil {
		log.Fatalf("failed to initialize approval: %v", err)
	}
	// initialize service
	svc, err := core.NewService(core.ServiceConfig{
		Queue:     queue.NewQueue(rmqPub),
		Email:     emailClient,
		Tokenizer: tokenizer,
		Approval:  approval,
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
