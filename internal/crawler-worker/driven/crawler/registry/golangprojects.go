package registry

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"github.com/go-resty/resty/v2"
	"gopkg.in/validator.v2"
)

const (
	golangProjectsURL             = "https://www.golangprojects.com/golang-rest-of-world-jobs.html"
	golangProjectsConcurrentJobs  = 10
	golangProjectsUnnecessaryTags = "head, nav, iframe, script, noscript, footer"
)

type GolangProjectsCrawler struct {
	GolangProjectsCrawlerConfig
}

type GolangProjectsCrawlerConfig struct {
	HttpClient *resty.Client `validate:"nonnil"`
}

func NewGolangProjectsCrawler(cfg GolangProjectsCrawlerConfig) (*GolangProjectsCrawler, error) {
	if err := validator.Validate(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &GolangProjectsCrawler{
		GolangProjectsCrawlerConfig: cfg,
	}, nil
}

func (p *GolangProjectsCrawler) Crawl(ctx context.Context) ([]core.Vacancy, error) {
	vacanciesLinks, err := p.getLinks(ctx, golangProjectsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get vacancies links: %w", err)
	}

	vacancies := make([]core.Vacancy, 0)

	type result struct {
		vacancy core.Vacancy
		link    string
		err     error
	}
	resultCh := make(chan result, len(vacanciesLinks))

	// Limit to 10 concurrent requests
	semaphore := make(chan struct{}, golangProjectsConcurrentJobs)

	// Process links in parallel
	for _, link := range vacanciesLinks {
		go func(link string) {
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore when done

			vacancy, err := p.getInfo(ctx, link)
			resultCh <- result{vacancy: vacancy, link: link, err: err}
		}(link)
	}

	// Collect results
	for i := 0; i < len(vacanciesLinks); i++ {
		res := <-resultCh
		if res.err != nil {
			log.Printf("failed to get vacancy info from link %s due to: %v", res.link, res.err)
			continue
		}
		vacancies = append(vacancies, res.vacancy)
	}

	return vacancies, nil
}

func (p *GolangProjectsCrawler) getLinks(ctx context.Context, url string) ([]string, error) {
	// Configure request with a browser-like User-Agent to avoid basic scraping protection
	resp, err := p.HttpClient.R().
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36").
		SetContext(ctx).
		Get(url)

	if err != nil {
		return nil, fmt.Errorf("failed to open the URL: %w", err)
	}

	// parse the HTML content
	respBody := bytes.NewReader(resp.Body())
	doc, err := goquery.NewDocumentFromReader(respBody)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the HTML content: %w", err)
	}

	// Remove unnecessary tags from the document
	doc.Find(golangProjectsUnnecessaryTags).Each(func(i int, s *goquery.Selection) {
		s.Remove()
	})

	var jobLinks []string

	// Find job links after each hr.clear-both but before the next hr.clear-both
	doc.Find("hr.clear-both").Each(func(i int, hr *goquery.Selection) {
		nextUntilHr := hr.NextUntil("hr.clear-both")

		// Check if the job posting contains "Worldwide, 100% Remote"
		jobText := nextUntilHr.Text()
		if !strings.Contains(jobText, "Worldwide, 100% Remote") {
			return
		}

		nextUntilHr.Find("a[href]").First().Each(func(j int, a *goquery.Selection) {
			href, exists := a.Attr("href")
			if exists {
				if strings.HasPrefix(href, "/") {
					href = "https://www.golangprojects.com" + href
				}
				jobLinks = append(jobLinks, href)
			}
		})
	})

	return jobLinks, nil
}

func (p *GolangProjectsCrawler) getInfo(ctx context.Context, url string) (core.Vacancy, error) {
	vacancy := core.Vacancy{}

	// Configure request with a browser-like User-Agent
	resp, err := p.HttpClient.R().
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36").
		SetContext(ctx).
		Get(url)

	if err != nil {
		return vacancy, fmt.Errorf("failed to open the URL: %w", err)
	}

	// parse the HTML content
	respBody := bytes.NewReader(resp.Body())
	doc, err := goquery.NewDocumentFromReader(respBody)
	if err != nil {
		return vacancy, fmt.Errorf("failed to parse the HTML content: %w", err)
	}

	// Find the "Apply now!" button URL first
	applyURL := ""
	doc.Find("a.cs-button[role='button']").Each(func(i int, s *goquery.Selection) {
		buttonText := strings.TrimSpace(s.Text())
		if strings.Contains(strings.ToLower(buttonText), "apply now") {
			href, exists := s.Attr("href")
			if exists {
				applyURL = href
				return
			}
		}
	})

	if applyURL == "" {
		return vacancy, fmt.Errorf("no apply URL found, skipping vacancy")
	}

	if applyURL != "" {
		vacancy.ApplyURL = applyURL
	}

	// Find the job title
	jobTitle := doc.Find("h1").First().Text()
	if jobTitle != "" {
		vacancy.JobTitle = strings.TrimSpace(jobTitle)
	}

	// Find company name
	doc.Find("h3").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if strings.Contains(text, "Company: ") {
			companyName := strings.TrimSpace(strings.TrimPrefix(text, "Company: "))
			if companyName != "" {
				vacancy.CompanyName = companyName
			}
		}
	})

	// Get the job description
	var descBuilder strings.Builder
	jobDescSection := doc.Find("b:contains('Job description')").Parent()
	if jobDescSection.Length() > 0 {
		var currentNode *goquery.Selection = jobDescSection.Next()
		for currentNode.Length() > 0 {
			// Stop if we hit the "Other Golang jobs" heading
			if currentNode.Find("h3:contains('Other Golang jobs')").Length() > 0 {
				break
			}

			text := strings.TrimSpace(currentNode.Text())
			if text != "" {
				descBuilder.WriteString(text + "\n")
			}

			currentNode = currentNode.Next()
		}
		vacancy.ShortDescription = strings.TrimSpace(descBuilder.String())
	}

	return vacancy, nil
}
