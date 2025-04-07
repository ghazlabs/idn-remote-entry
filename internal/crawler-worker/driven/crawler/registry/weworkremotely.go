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
	url             = "https://weworkremotely.com/100-percent-remote-jobs"
	concurrentJobs  = 10
	unnecessaryTags = "head, script, style, footer, iframe, #nav-header, #mobile-header"
)

type WeWorkRemotelyCrawler struct {
	WeWorkRemotelyCrawlerConfig
}

type WeWorkRemotelyCrawlerConfig struct {
	HttpClient *resty.Client `validate:"nonnil"`
}

func NewWeWorkRemotelyCrawler(cfg WeWorkRemotelyCrawlerConfig) (*WeWorkRemotelyCrawler, error) {
	if err := validator.Validate(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &WeWorkRemotelyCrawler{
		WeWorkRemotelyCrawlerConfig: cfg,
	}, nil
}

func (p *WeWorkRemotelyCrawler) Crawl(ctx context.Context) ([]core.Vacancy, error) {
	vacanciesLinks, err := p.getLinks(ctx, url)
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
	semaphore := make(chan struct{}, concurrentJobs)

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

func (p *WeWorkRemotelyCrawler) getLinks(ctx context.Context, url string) ([]string, error) {
	// get the html content of the URL
	resp, err := p.HttpClient.R().SetContext(ctx).Get(url)
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
	doc.Find(unnecessaryTags).Each(func(i int, s *goquery.Selection) {
		s.Remove()
	})

	var jobLinks []string
	doc.Find(".jobs li a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists && strings.HasPrefix(href, "/remote-jobs/") {
			fullURL := "https://weworkremotely.com" + href
			jobLinks = append(jobLinks, fullURL)
		}
	})

	return jobLinks, nil
}

func (p *WeWorkRemotelyCrawler) getInfo(ctx context.Context, url string) (core.Vacancy, error) {
	vacancy := core.Vacancy{}
	// get the html content of the URL
	resp, err := p.HttpClient.R().SetContext(ctx).Get(url)
	if err != nil {
		return vacancy, fmt.Errorf("failed to open the URL: %w", err)
	}

	// parse the HTML content
	respBody := bytes.NewReader(resp.Body())
	doc, err := goquery.NewDocumentFromReader(respBody)
	if err != nil {
		return vacancy, fmt.Errorf("failed to parse the HTML content: %w", err)
	}

	// Remove unnecessary tags from the document
	doc.Find(unnecessaryTags).Each(func(i int, s *goquery.Selection) {
		s.Remove()
	})

	// Extract job details
	vacancy.JobTitle = strings.TrimSpace(doc.Find("h2.lis-container__header__hero__company-info__title").Text())
	vacancy.CompanyName = strings.TrimSpace(doc.Find("div.lis-container__job__sidebar__companyDetails__info__title h3").Text())
	vacancy.ApplyURL, _ = doc.Find("a#job-cta-alt").Attr("href")

	var descBuilder strings.Builder
	doc.Find("div.lis-container__job__content__description").Children().Each(func(i int, s *goquery.Selection) {
		text := strings.TrimSpace(s.Text())
		if text != "" {
			descBuilder.WriteString(text + "\n")
		}
	})
	vacancy.ShortDescription = strings.TrimSpace(descBuilder.String())
	// skip location and tags for now

	return vacancy, nil
}
