package parser

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/go-resty/resty/v2"
	"gopkg.in/validator.v2"
)

type OCRParser struct {
	OCRParserConfig
}

type OCRParserConfig struct {
	HttpClient *resty.Client `validate:"nonnil"`
}

func NewOCRParser(cfg OCRParserConfig) (*OCRParser, error) {
	if err := validator.Validate(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	return &OCRParser{
		OCRParserConfig: cfg,
	}, nil
}

func (p *OCRParser) GetText(ctx context.Context, url string) (string, error) {
	// take a screenshot of the URL
	buf, err := p.takeScreenshot(ctx, url)
	if err != nil {
		return "", fmt.Errorf("failed to take the screenshot: %w", err)
	}

	// do OCR on the screenshot
	text, err := p.doOCR(ctx, buf)
	if err != nil {
		return "", fmt.Errorf("failed to do OCR: %w", err)
	}

	return text, nil
}

func (p *OCRParser) takeScreenshot(ctx context.Context, url string) ([]byte, error) {
	// create context for chrome
	ctx, cancel := chromedp.NewContext(ctx)
	defer cancel()

	// allocate a buffer to store the screenshot
	var buf []byte

	// capture the screenshot
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(3*time.Second),
		chromedp.FullScreenshot(&buf, 90),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to take the screenshot: %w", err)
	}

	return buf, nil
}

type tesseractServerResponse struct {
	Data struct {
		Stdout string `json:"stdout"`
	} `json:"data"`
}

func (p *OCRParser) doOCR(ctx context.Context, buf []byte) (string, error) {
	// call tesseract server to do OCR
	var serverResp tesseractServerResponse
	_, err := p.HttpClient.R().
		SetContext(ctx).
		SetFileReader("file", "file", bytes.NewReader(buf)).
		SetFormData(map[string]string{
			"options": `{"languages": ["eng"]}`, // Add the "options" field
		}).
		SetResult(&serverResp).
		Post("http://127.0.0.1:8884/tesseract")
	if err != nil {
		return "", fmt.Errorf("failed to call the OCR server: %w", err)
	}
	return serverResp.Data.Stdout, nil
}
