package main

import (
	"log"

	"github.com/ghazlabs/idn-remote-entry/internal/shared/rmq"
	"github.com/ghazlabs/idn-remote-entry/internal/vacancy-worker/core"
	"github.com/ghazlabs/idn-remote-entry/internal/vacancy-worker/driven/notifier"
	"github.com/ghazlabs/idn-remote-entry/internal/vacancy-worker/driven/resolver"
	"github.com/ghazlabs/idn-remote-entry/internal/vacancy-worker/driven/resolver/hqloc"
	"github.com/ghazlabs/idn-remote-entry/internal/vacancy-worker/driven/resolver/parser"
	"github.com/ghazlabs/idn-remote-entry/internal/vacancy-worker/driven/storage"
	"github.com/go-resty/resty/v2"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/riandyrn/go-env"
)

const (
	envKeyNotionDatabaseID     = "NOTION_DATABASE_ID"
	envKeyNotionToken          = "NOTION_TOKEN"
	envKeyOpenAiKey            = "OPENAI_KEY"
	envKeyWhatsappRecipientIDs = "WHATSAPP_RECIPIENT_IDS"
	envKeyRabbitMQConn         = "RABBITMQ_CONN"
	envKeyRabbitMQWaQueueName  = "RABBITMQ_WA_QUEUE_NAME"
)

func main() {
	// initialize storage
	httpClient := resty.New()
	strg, err := storage.NewNotionStorage(storage.NotionStorageConfig{
		DatabaseID:  env.GetString(envKeyNotionDatabaseID),
		NotionToken: env.GetString(envKeyNotionToken),
		HttpClient:  httpClient,
	})
	if err != nil {
		log.Fatalf("failed to initialize storage: %v", err)
	}

	// initialize parser
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
	waNotf, err := notifier.NewWaNotifier(notifier.WaNotifierConfig{
		RmqPublisher:   waRmqPub,
		WaRecipientIDs: env.GetStrings(envKeyWhatsappRecipientIDs, ","),
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
}
