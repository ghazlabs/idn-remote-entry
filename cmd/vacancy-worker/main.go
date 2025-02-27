package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/ghazlabs/idn-remote-entry/internal/shared/rmq"
	"github.com/ghazlabs/idn-remote-entry/internal/vacancy-worker/core"
	"github.com/ghazlabs/idn-remote-entry/internal/vacancy-worker/driven/notifier"
	"github.com/ghazlabs/idn-remote-entry/internal/vacancy-worker/driven/resolver"
	"github.com/ghazlabs/idn-remote-entry/internal/vacancy-worker/driven/resolver/hqloc"
	"github.com/ghazlabs/idn-remote-entry/internal/vacancy-worker/driven/resolver/parser"
	"github.com/ghazlabs/idn-remote-entry/internal/vacancy-worker/driven/storage/mysql"
	"github.com/ghazlabs/idn-remote-entry/internal/vacancy-worker/driven/storage/notion"
	"github.com/ghazlabs/idn-remote-entry/internal/vacancy-worker/driver/worker"
	"github.com/go-resty/resty/v2"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/riandyrn/go-env"

	_ "github.com/go-sql-driver/mysql"
)

const (
	envKeyStorageType              = "STORAGE_TYPE"
	envKeyMysqlDSN                 = "MYSQL_DSN"
	envKeyNotionDatabaseID         = "NOTION_DATABASE_ID"
	envKeyNotionToken              = "NOTION_TOKEN"
	envKeyOpenAiKey                = "OPENAI_KEY"
	envKeyRabbitMQConn             = "RABBITMQ_CONN"
	envKeyRabbitMQWaQueueName      = "RABBITMQ_WA_QUEUE_NAME"
	envKeyRabbitMQVacancyQueueName = "RABBITMQ_VACANCY_QUEUE_NAME"
)

func initStorage() (core.Storage, error) {
	switch env.GetString(envKeyStorageType) {
	case "notion":
		httpClient := resty.New()
		return notion.NewNotionStorage(notion.NotionStorageConfig{
			DatabaseID:  env.GetString(envKeyNotionDatabaseID),
			NotionToken: env.GetString(envKeyNotionToken),
			HttpClient:  httpClient,
		})
	case "mysql":
		mysqlClient, err := sql.Open("mysql", env.GetString(envKeyMysqlDSN))
		if err != nil {
			return nil, err
		}
		return mysql.NewMySQLStorage(mysql.MySQLStorageConfig{
			DB: mysqlClient,
		})
	default:
		return nil, fmt.Errorf("invalid storage type: %s", env.GetString(envKeyStorageType))
	}
}

func main() {
	// initialize storage
	strg, err := initStorage()
	if err != nil {
		log.Fatalf("failed to initialize storage: %v", err)
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
		HttpClient:    httpClient,
		OpenaAiClient: openAiClient,
		DatabaseID:    env.GetString(envKeyNotionDatabaseID),
		NotionToken:   env.GetString(envKeyNotionToken),
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

	// initialize rabbitmq publisher
	waRmqPub, err := rmq.NewPublisher(rmq.PublisherConfig{
		QueueName:          env.GetString(envKeyRabbitMQWaQueueName),
		RabbitMQConnString: env.GetString(envKeyRabbitMQConn),
	})
	if err != nil {
		log.Fatalf("failed to initialize rabbitmq publisher: %v", err)
	}
	defer waRmqPub.Close()

	// initialize notifier
	waNotf, err := notifier.NewNotifier(notifier.NotifierConfig{
		RmqPublisher: waRmqPub,
	})
	if err != nil {
		log.Fatalf("failed to initialize notifier: %v", err)
	}

	// initialize service
	svc, err := core.NewService(core.ServiceConfig{
		Storage:         strg,
		VacancyResolver: rslvr,
		Notifier:        waNotf,
	})
	if err != nil {
		log.Fatalf("failed to initialize service: %v", err)
	}

	// initialize consumer
	rmqConsumer, err := rmq.NewConsumer(rmq.ConsumerConfig{
		QueueName:          env.GetString(envKeyRabbitMQVacancyQueueName),
		RabbitMQConnString: env.GetString(envKeyRabbitMQConn),
	})
	if err != nil {
		log.Fatalf("failed to initialize consumer: %v", err)
	}
	defer rmqConsumer.Close()

	// initialize worker
	w, err := worker.New(worker.Config{
		Service:     svc,
		RmqConsumer: rmqConsumer,
	})
	if err != nil {
		log.Fatalf("failed to initialize worker: %v", err)
	}

	// run worker
	log.Printf("vacancy-worker is running...")
	err = w.Run()
	if err != nil {
		log.Fatalf("failed to run worker: %v", err)
	}
}
