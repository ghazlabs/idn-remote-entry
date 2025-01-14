package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ghazlabs/idn-remote-entry/internal/core"
	"github.com/ghazlabs/idn-remote-entry/internal/driven/resolver"
	"github.com/ghazlabs/idn-remote-entry/internal/driven/resolver/parser"
	"github.com/ghazlabs/idn-remote-entry/internal/driven/storage"
	"github.com/ghazlabs/idn-remote-entry/internal/driver"
	"github.com/go-resty/resty/v2"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/riandyrn/go-env"
)

const (
	envKeyNotionDatabaseID = "NOTION_DATABASE_ID"
	envKeyNotionToken      = "NOTION_TOKEN"
	envKeyOpenAiKey        = "OPENAI_KEY"
	envKeyListenPort       = "LISTEN_PORT"
	envKeyClientApiKey     = "CLIENT_API_KEY"
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
		HttpClient:    httpClient,
		OpenaAiClient: openAiClient,
	})
	if err != nil {
		log.Fatalf("failed to initialize OCR parser: %v", err)
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
	})
	if err != nil {
		log.Fatalf("failed to initialize resolver: %v", err)
	}

	// initialize service
	svc, err := core.NewService(core.ServiceConfig{
		Storage:         strg,
		VacancyResolver: rslvr,
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
