package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ghazlabs/idn-remote-entry/internal/crawler-worker/core"
	contentchecker "github.com/ghazlabs/idn-remote-entry/internal/crawler-worker/driven/content-checker"
	"github.com/ghazlabs/idn-remote-entry/internal/crawler-worker/driven/crawler"
	"github.com/ghazlabs/idn-remote-entry/internal/crawler-worker/driven/crawler/registry"
	"github.com/ghazlabs/idn-remote-entry/internal/crawler-worker/driven/server"
	"github.com/ghazlabs/idn-remote-entry/internal/crawler-worker/driven/storage/notion"
	"github.com/go-resty/resty/v2"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/riandyrn/go-env"
	"github.com/robfig/cron/v3"
)

const (
	envKeyNotionDatabaseID = "NOTION_DATABASE_ID"
	envKeyNotionToken      = "NOTION_TOKEN"
	envKeyOpenAiKey        = "OPENAI_KEY"
	envKeyServerAPIKey     = "SERVER_API_KEY"
	envKeyServerBaseURL    = "SERVER_BASE_URL"
	envKeyCronSchedule     = "CRON_SCHEDULE"
)

func main() {
	// initialize crawler
	wwrCrawler, err := registry.NewWeWorkRemotelyCrawler(registry.WeWorkRemotelyCrawlerConfig{
		HttpClient: resty.New(),
	})
	if err != nil {
		log.Fatalf("failed to initialize wwr crawler: %v", err)
	}
	wwrRegister := crawler.CrawlRegistry{
		Name:    "weworkremotely",
		Crawler: wwrCrawler,
	}

	// initialize resolver
	crawlers, err := crawler.NewVacancyCrawler(crawler.VacancyResolverConfig{
		CrawlerRegistries: []crawler.CrawlRegistry{
			wwrRegister,
		},
	})
	if err != nil {
		log.Fatalf("failed to initialize resolver: %v", err)
	}

	// initialize storage
	storage, err := notion.NewNotionStorage(notion.NotionStorageConfig{
		DatabaseID:  env.GetString(envKeyNotionDatabaseID),
		NotionToken: env.GetString(envKeyNotionToken),
		HttpClient:  resty.New(),
	})
	if err != nil {
		log.Fatalf("failed to initialize storage: %v", err)
	}

	// initialize content checker
	openAiClient := openai.NewClient(option.WithAPIKey(env.GetString(envKeyOpenAiKey)))
	contentChecker, err := contentchecker.NewContentChecker(contentchecker.CheckerConfig{
		OpenaAiClient: openAiClient,
	})
	if err != nil {
		log.Fatalf("failed to initialize content checker: %v", err)
	}

	// initialize server
	client, err := server.NewClientServer(server.ServerConfig{
		HttpClient: resty.New().SetBaseURL(env.GetString(envKeyServerBaseURL)),
		ApiKey:     env.GetString(envKeyServerAPIKey),
	})
	if err != nil {
		log.Fatalf("failed to initialize server: %v", err)
	}

	// initialize service
	svc, err := core.NewService(core.ServiceConfig{
		Crawler:        crawlers,
		VacancyStorage: storage,
		ContentChecker: contentChecker,
		Server:         client,
	})
	if err != nil {
		log.Fatalf("failed to initialize service: %v", err)
	}

	// Create a new cron instance
	c := cron.New()
	c.AddFunc(env.GetString(envKeyCronSchedule), func() {
		fmt.Println("Running scheduled task...")
		err := svc.Run(context.Background())
		if err != nil {
			log.Printf("failed to run service: %v", err)
		}
	})

	// Start the cron scheduler
	c.Start()
	fmt.Println("Cron scheduler initialized")

	// Defer the stop of the cron scheduler to ensure it stops when main function exits
	defer c.Stop()

	// Keep the main program running
	select {}
}
