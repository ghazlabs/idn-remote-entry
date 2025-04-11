package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/ghazlabs/idn-remote-entry/internal/crawler-worker/core"
	contentchecker "github.com/ghazlabs/idn-remote-entry/internal/crawler-worker/driven/content-checker"
	"github.com/ghazlabs/idn-remote-entry/internal/crawler-worker/driven/crawler"
	"github.com/ghazlabs/idn-remote-entry/internal/crawler-worker/driven/crawler/registry"
	"github.com/ghazlabs/idn-remote-entry/internal/crawler-worker/driven/server"
	"github.com/ghazlabs/idn-remote-entry/internal/crawler-worker/driven/storage/mysql"
	"github.com/ghazlabs/idn-remote-entry/internal/crawler-worker/driven/storage/notion"
	"github.com/go-resty/resty/v2"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/riandyrn/go-env"
	"github.com/robfig/cron/v3"

	_ "github.com/go-sql-driver/mysql"
)

const (
	envKeyNotionDatabaseID             = "NOTION_DATABASE_ID"
	envKeyNotionToken                  = "NOTION_TOKEN"
	envKeyOpenAiKey                    = "OPENAI_KEY"
	envKeyServerAPIKey                 = "SERVER_API_KEY"
	envKeyServerBaseURL                = "SERVER_BASE_URL"
	envKeyCronSchedule                 = "CRON_SCHEDULE"
	envKeyEnabledApplicableChecker     = "ENABLED_APPLICABLE_CHECKER_LLM"
	envKeyEnabledWeWorkRemotelyCrawler = "ENABLED_WEWORKREMOTELY_CRAWLER"
	envKeyEnabledGolangProjectsCrawler = "ENABLED_GOLANGPROJECTS_CRAWLER"
	envKeyMysqlDSN                     = "MYSQL_DSN"
)

func main() {
	// initialize crawler
	crawlRegistryList := []crawler.CrawlRegistry{}
	if env.GetBool(envKeyEnabledWeWorkRemotelyCrawler) {
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

		crawlRegistryList = append(crawlRegistryList, wwrRegister)

	}

	if env.GetBool(envKeyEnabledGolangProjectsCrawler) {
		golangProjectsCrawler, err := registry.NewGolangProjectsCrawler(registry.GolangProjectsCrawlerConfig{
			HttpClient: resty.New(),
		})
		if err != nil {
			log.Fatalf("failed to initialize golangprojects crawler: %v", err)
		}

		golangProjectsRegister := crawler.CrawlRegistry{
			Name:    "golangprojects",
			Crawler: golangProjectsCrawler,
		}
		
		crawlRegistryList = append(crawlRegistryList, golangProjectsRegister)
	}

	// initialize resolver
	crawlers, err := crawler.NewVacancyCrawler(crawler.VacancyResolverConfig{
		CrawlerRegistries: crawlRegistryList,
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

	// initialize approval storage
	mysqlClient, err := sql.Open("mysql", env.GetString(envKeyMysqlDSN))
	if err != nil {
		log.Fatalf("failed to initialize mysql client: %v", err)
	}

	if err := mysqlClient.Ping(); err != nil {
		log.Fatalf("failed to ping mysql client: %v", err)
	}

	approvalStorage, err := mysql.NewMySQLStorage(mysql.MySQLStorageConfig{
		DB: mysqlClient,
	})
	if err != nil {
		log.Fatalf("failed to initialize approval storage: %v", err)
	}

	// initialize service
	svc, err := core.NewService(core.ServiceConfig{
		Crawler:                  crawlers,
		VacancyStorage:           storage,
		ContentChecker:           contentChecker,
		ApprovalStorage:          approvalStorage,
		Server:                   client,
		EnabledApplicableChecker: env.GetBool(envKeyEnabledApplicableChecker),
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
		fmt.Println("Scheduled task completed")
	})

	// Start the cron scheduler
	c.Start()
	fmt.Println("Cron scheduler initialized")

	// Defer the stop of the cron scheduler to ensure it stops when main function exits
	defer c.Stop()

	// Keep the main program running
	select {}
}
