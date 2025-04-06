package contentchecker

import (
	"context"
	"fmt"

	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"github.com/ghazlabs/idn-remote-entry/internal/vacancy-worker/driven/resolver/util"
	"github.com/openai/openai-go"
	"gopkg.in/validator.v2"
)

type applicable struct {
	IsApplicableForIndonesia bool `json:"is_applicable" jsonschema_description:"Is the vacancy applicable for Indonesian."`
}

type Checker struct {
	CheckerConfig
}

type CheckerConfig struct {
	OpenaAiClient *openai.Client `validate:"nonnil"`
}

func NewContentChecker(cfg CheckerConfig) (*Checker, error) {
	err := validator.Validate(cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	return &Checker{
		CheckerConfig: cfg,
	}, nil
}

func (c *Checker) IsApplicableForIndonesian(ctx context.Context, v core.Vacancy) (bool, error) {
	// fmt.Printf("desc: %s\n", v.ShortDescription)

	applicable, err := util.CallOpenAI[applicable](ctx, c.OpenaAiClient, []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage("I will give you unstructured text content of a remote vacancy, and you need to determine whether the vacancy is applicable for Indonesian (GMT +7) or not. Please answer with true or false."),
		openai.UserMessage(v.ShortDescription),
	})
	if err != nil {
		return false, fmt.Errorf("unable to parse the vacancy information: %w", err)
	}

	fmt.Printf("applicable: %v\n", applicable)

	return applicable.IsApplicableForIndonesia, nil
}
