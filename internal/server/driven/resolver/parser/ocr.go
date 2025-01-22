package parser

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/ghazlabs/idn-remote-entry/internal/server/core"
	"github.com/ghazlabs/idn-remote-entry/internal/server/driven/resolver"
	"github.com/ghazlabs/idn-remote-entry/internal/server/driven/resolver/util"
	"github.com/openai/openai-go"
	"gopkg.in/validator.v2"
)

type OCRParser struct {
	OCRParserConfig
}

type OCRParserConfig struct {
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
	buf, err := util.TakeScreenshot(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to take the screenshot: %w", err)
	}

	// do OCR on the screenshot
	vac, err := p.doOCR(ctx, buf, url)
	if err != nil {
		return nil, fmt.Errorf("failed to do OCR: %w", err)
	}

	return vac, nil
}

func (p *OCRParser) doOCR(ctx context.Context, buf []byte, url string) (*core.Vacancy, error) {
	// call the OpenAI API to parse the vacancy information
	vacInfo, err := util.CallOpenAI[resolver.VacancyInfo](ctx, p.OpenaAiClient, []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage("You will be given vacancy description from the image and you need to parse the information from it."),
		openai.UserMessageParts(openai.ImagePart(fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(buf)))),
	})
	if err != nil {
		return nil, fmt.Errorf("unable to parse the vacancy information: %w", err)
	}

	return vacInfo.ToVacancy(url), nil
}
