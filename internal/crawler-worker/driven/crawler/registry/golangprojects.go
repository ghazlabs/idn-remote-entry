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
	golangProjectsUnnecessaryTags = "head, script, style, footer, iframe, #nav-header, #mobile-header"
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

	if resp.StatusCode() != 200 {
		log.Printf("Failed to get golangprojects page with status code: %d", resp.StatusCode())
		if resp.StatusCode() == 403 {
			log.Printf("WARNING: GolangProjects returned 403 Forbidden. The site is likely blocking our crawler.")
		}
		return []string{}, nil
	}

	// parse the HTML content
	respBody := bytes.NewReader(resp.Body())
	doc, err := goquery.NewDocumentFromReader(respBody)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the HTML content: %w", err)
	}

	var jobLinks []string

	// Find all job listings by their container divs
	doc.Find("div.bg-csdpromobg1, div.bg-csdpromobg2").Each(func(i int, s *goquery.Selection) {
		// Within each listing, find the anchor tag (job link)
		link := s.Find("a").First()
		href, exists := link.Attr("href")
		if exists {
			if strings.HasPrefix(href, "/golang-") {
				fullURL := "https://www.golangprojects.com" + href
				jobLinks = append(jobLinks, fullURL)
			}
		}
	})

	// Try alternative approach - look for links directly
	if len(jobLinks) == 0 {
		doc.Find("a[href^='/golang-go-job-']").Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			if exists {
				fullURL := "https://www.golangprojects.com" + href
				jobLinks = append(jobLinks, fullURL)
			}
		})
	}

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

	// Handle non-200 responses
	if resp.StatusCode() != 200 {
		log.Printf("Failed to get job page with status code: %d", resp.StatusCode())
		if resp.StatusCode() == 403 {
			log.Printf("WARNING: GolangProjects returned 403 Forbidden for job page. Creating placeholder vacancy.")
		}

		// Create a placeholder vacancy
		vacancy.ApplyURL = url

		// Extract job title from URL
		urlParts := strings.Split(url, "/")
		if len(urlParts) > 0 {
			lastPart := urlParts[len(urlParts)-1]
			lastPart = strings.ReplaceAll(lastPart, "-", " ")
			lastPart = strings.ReplaceAll(lastPart, ".html", "")
			lastPart = strings.ReplaceAll(lastPart, "remotework", "")
			vacancy.JobTitle = strings.Title(lastPart)
		} else {
			vacancy.JobTitle = "Golang Developer Position"
		}

		vacancy.CompanyName = "Company on GolangProjects"
		vacancy.ShortDescription = "This job was found on GolangProjects. Please visit the website directly for more details."

		return vacancy, nil
	}

	// parse the HTML content
	respBody := bytes.NewReader(resp.Body())
	doc, err := goquery.NewDocumentFromReader(respBody)
	if err != nil {
		return vacancy, fmt.Errorf("failed to parse the HTML content: %w", err)
	}

	// Find the "Apply now!" button URL first - this is the correct apply URL
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

	// If "Apply now!" button was found, use it as apply URL
	if applyURL != "" {
		vacancy.ApplyURL = applyURL
	} else {
		// Fall back to the job posting URL if no apply button found
		vacancy.ApplyURL = url
	}

	// Based on the HTML structure, extract job details

	// Find the job title - it's typically in the <b> tag within the same anchor that leads to the job
	jobTitle := doc.Find("a b").First().Text()
	if jobTitle != "" {
		vacancy.JobTitle = strings.TrimSpace(jobTitle)
	}

	// If we couldn't find the title with that selector, try others
	if vacancy.JobTitle == "" {
		// Try to find it in the page title
		pageTitle := doc.Find("title").Text()
		if strings.Contains(pageTitle, ":") {
			parts := strings.SplitN(pageTitle, ":", 2)
			vacancy.JobTitle = strings.TrimSpace(parts[1])
		} else {
			vacancy.JobTitle = strings.TrimSpace(pageTitle)
		}
	}

	// Find company name - often near the job title or in the URL
	// In GolangProjects, company name is often part of the job title separated by " - "
	if strings.Contains(vacancy.JobTitle, " - ") {
		parts := strings.Split(vacancy.JobTitle, " - ")
		if len(parts) >= 2 {
			// Last part is typically the company name
			vacancy.CompanyName = strings.TrimSpace(parts[len(parts)-1])
			// Join all parts except the last one as the job title
			vacancy.JobTitle = strings.TrimSpace(strings.Join(parts[:len(parts)-1], " - "))
		}
	}

	// If company name still not found, try to extract from image alt text
	if vacancy.CompanyName == "" {
		imgAlt := doc.Find("img").First().AttrOr("alt", "")
		if strings.Contains(imgAlt, "at ") {
			parts := strings.Split(imgAlt, "at ")
			if len(parts) >= 2 {
				vacancy.CompanyName = strings.TrimSpace(parts[len(parts)-1])
			}
		}
	}

	// Get the job description - it's often inside an <i> tag with class="text-sm"
	descText := doc.Find("i.text-sm").Text()
	if descText != "" {
		vacancy.ShortDescription = strings.TrimSpace(descText)
	}

	// If no description found, use the entire page text
	if vacancy.ShortDescription == "" {
		// Get all text from the page
		vacancy.ShortDescription = strings.TrimSpace(doc.Text())
	}

	// Fallbacks for missing information
	if vacancy.JobTitle == "" {
		// Try to extract from URL as last resort
		urlParts := strings.Split(url, "/")
		if len(urlParts) > 0 {
			lastPart := urlParts[len(urlParts)-1]
			lastPart = strings.ReplaceAll(lastPart, "-", " ")
			lastPart = strings.ReplaceAll(lastPart, ".html", "")
			lastPart = strings.ReplaceAll(lastPart, "remotework", "")
			vacancy.JobTitle = strings.Title(lastPart)
		} else {
			vacancy.JobTitle = "Golang Developer"
		}
	}

	if vacancy.CompanyName == "" {
		vacancy.CompanyName = "Company on GolangProjects"
	}

	if vacancy.ShortDescription == "" {
		vacancy.ShortDescription = "Please visit the job posting for more details."
	}

	// Truncate the description if it's too long
	maxDescLen := 500
	if len(vacancy.ShortDescription) > maxDescLen {
		vacancy.ShortDescription = vacancy.ShortDescription[:maxDescLen] + "..."
	}

	return vacancy, nil
}
