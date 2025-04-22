package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ghazlabs/idn-remote-entry/internal/server/core"
	"github.com/ghazlabs/idn-remote-entry/internal/server/driven/approval"
	"github.com/ghazlabs/idn-remote-entry/internal/server/driven/email"
	"github.com/ghazlabs/idn-remote-entry/internal/server/driven/queue"
	"github.com/ghazlabs/idn-remote-entry/internal/server/driven/storage/mysql"
	"github.com/ghazlabs/idn-remote-entry/internal/server/driven/token"
	"github.com/ghazlabs/idn-remote-entry/internal/server/driver"
	"github.com/ghazlabs/idn-remote-entry/internal/shared/rmq"
	vwcore "github.com/ghazlabs/idn-remote-entry/internal/vacancy-worker/core"
	"github.com/ghazlabs/idn-remote-entry/internal/vacancy-worker/driven/resolver"
	"github.com/ghazlabs/idn-remote-entry/internal/vacancy-worker/driven/resolver/hqloc"
	"github.com/ghazlabs/idn-remote-entry/internal/vacancy-worker/driven/resolver/parser"
	"github.com/ghazlabs/idn-remote-entry/internal/vacancy-worker/driven/storage/jsonl"
	shnotion "github.com/ghazlabs/idn-remote-entry/internal/vacancy-worker/driven/storage/notion"
	"github.com/go-resty/resty/v2"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/riandyrn/go-env"

	_ "github.com/go-sql-driver/mysql"
)

const (
	envKeyStorageType              = "STORAGE_TYPE"
	envKeyNotionDatabaseID         = "NOTION_DATABASE_ID"
	envKeyNotionToken              = "NOTION_TOKEN"
	envKeyOpenAiKey                = "OPENAI_KEY"
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
	envKeyMysqlDSN                 = "MYSQL_DSN"
)

func initStorage() (vwcore.Storage, error) {
	switch env.GetString(envKeyStorageType) {
	case "jsonl":
		return jsonl.NewJSONLStorage(jsonl.JSONLStorageConfig{
			FilePath: "vacancies.jsonl",
		})

	default: // notion
		httpClient := resty.New()
		return shnotion.NewNotionStorage(shnotion.NotionStorageConfig{
			DatabaseID:  env.GetString(envKeyNotionDatabaseID),
			NotionToken: env.GetString(envKeyNotionToken),
			HttpClient:  httpClient,
		})
	}
}

func main() {
	// initialize storage
	strg, err := initStorage()
	if err != nil {
		log.Fatalf("failed to initialize storage: %v", err)
	}

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

	mysqlClient, err := sql.Open("mysql", env.GetString(envKeyMysqlDSN))
	if err != nil {
		log.Fatalf("failed to initialize mysql client: %v", err)
	}

	approvalStorage, err := mysql.NewMySQLStorage(mysql.MySQLStorageConfig{
		DB: mysqlClient,
	})
	if err != nil {
		log.Fatalf("failed to initialize approval storage: %v", err)
	}

	// initialize parser
	httpClient := resty.New()
	openAiClient := openai.NewClient(option.WithAPIKey(env.GetString(envKeyOpenAiKey)))
	textParser, err := parser.NewGreenhouseParser(parser.GreenhouseParserConfig{
		HttpClient:    httpClient,
		OpenaAiClient: openAiClient,
	})
	if err != nil {
		log.Fatalf("failed to initialize text parser: %v", err)
	}
	ocrParser, err := parser.NewOCRParser(parser.OCRParserConfig{
		OpenaAiClient: openAiClient,
	})
	if err != nil {
		log.Fatalf("failed to initialize OCR parser: %v", err)
	}

	// initialize locator
	locator, err := hqloc.NewLocator(hqloc.LocatorConfig{
		OpenaAiClient: openAiClient,
		Storage:       strg,
	})
	if err != nil {
		log.Fatalf("failed to initialize locator: %v", err)
	}

	// initialize resolver
	rslvr, err := resolver.NewVacancyResolver(resolver.VacancyResolverConfig{
		DefaultParser: ocrParser,
		ParserRegistries: []resolver.ParserRegistry{
			{
				ApexDomains: []string{"greenhouse.io"},
				Parser:      textParser,
			},
		},
		HQLocator: locator,
	})
	if err != nil {
		log.Fatalf("failed to initialize resolver: %v", err)
	}

	// initialize service
	svc, err := core.NewService(core.ServiceConfig{
		VacancyResolver: rslvr,
		Queue:           queue.NewQueue(rmqPub),
		Email:           emailClient,
		Tokenizer:       tokenizer,
		Approval:        approval,
		ApprovalStorage: approvalStorage,
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
