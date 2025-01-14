package parser

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/ghazlabs/idn-remote-entry/internal/core"
	"github.com/ghazlabs/idn-remote-entry/internal/driven/resolver"
	"github.com/go-resty/resty/v2"
	"github.com/openai/openai-go"
	"gopkg.in/validator.v2"
)

type OCRParser struct {
	OCRParserConfig
}

type OCRParserConfig struct {
	HttpClient    *resty.Client  `validate:"nonnil"`
	OpenaAiClient *openai.Client `validate:"nonnil"`
}

func NewOCRParser(cfg OCRParserConfig) (*OCRParser, error) {
	if err := validator.Validate(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	return &OCRParser{
		OCRParserConfig: cfg,
	}, nil
}

func (p *OCRParser) Parse(ctx context.Context, url string) (*core.Vacancy, error) {
	// take a screenshot of the URL
	buf, err := p.takeScreenshot(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to take the screenshot: %w", err)
	}

	// do OCR on the screenshot
	vac, err := p.doOCR(ctx, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to do OCR: %w", err)
	}

	return vac, nil
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

func (p *OCRParser) doOCR(ctx context.Context, buf []byte) (*core.Vacancy, error) {
	// call the OpenAI API to parse the vacancy information
	vacInfo, err := resolver.CallOpenAI[resolver.VacancyInfo](ctx, p.OpenaAiClient, []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage("You will be given vacancy description from the image and you need to parse the information from it."),
		openai.UserMessageParts(openai.ImagePart(fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(buf)))),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to parse the vacancy information: %w", err)
	}

	return vacInfo.ToVacancy(), nil
}
