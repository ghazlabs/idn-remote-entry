package hqloc

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/url"

	"github.com/ghazlabs/idn-remote-entry/internal/driven/resolver/util"
	"github.com/go-resty/resty/v2"
	"github.com/openai/openai-go"
	"gopkg.in/validator.v2"
)

type Locator struct {
	LocatorConfig
}

func NewLocator(cfg LocatorConfig) (*Locator, error) {
	if err := validator.Validate(cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	return &Locator{
		LocatorConfig: cfg,
	}, nil
}

type LocatorConfig struct {
	HttpClient    *resty.Client  `validate:"nonnil"`
	DatabaseID    string         `validate:"nonzero"`
	NotionToken   string         `validate:"nonzero"`
	OpenaAiClient *openai.Client `validate:"nonnil"`
}

func (l *Locator) Locate(ctx context.Context, companyName string) (string, error) {
	companyLoc, err := l.lookupToNotion(ctx, companyName)
	if err != nil {
		return "", fmt.Errorf("failed to lookup company location from Notion: %w", err)
	}
	if companyLoc != "" {
		return companyLoc, nil
	}

	return l.lookupToWeb(ctx, companyName)
}

func (l *Locator) lookupToNotion(ctx context.Context, companyName string) (string, error) {
	var respBody NotionResponse
	resp, err := l.HttpClient.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+l.NotionToken).
		SetHeader("Content-Type", "application/json").
		SetHeader("Notion-Version", "2022-06-28").
		SetBody(map[string]interface{}{
			"filter": map[string]interface{}{
				"property": "Company Name",
				"rich_text": map[string]string{
					"equals": companyName,
				},
			},
			"sorts": []map[string]string{
				{
					"property":  "Last edited time",
					"direction": "descending",
				},
			},
			"page_size": 1,
		}).
		SetResult(&respBody).
		Post(fmt.Sprintf("https://api.notion.com/v1/databases/%s/query", l.DatabaseID))
	if err != nil {
		return "", fmt.Errorf("failed to call notion api to lookup company location: %w", err)
	}
	if resp.IsError() {
		return "", fmt.Errorf("failed to lookup company location from Notion: %s", resp.String())
	}

	return respBody.GetCompanyLocation(), nil
}

func (l *Locator) lookupToWeb(ctx context.Context, companyName string) (string, error) {
	queryParams := url.Values{}
	queryParams.Add("q", fmt.Sprintf("Where is %s hq located?", companyName))
	queryParams.Add("t", "h_")
	queryParams.Add("ia", "web")

	url := fmt.Sprintf(`https://duckduckgo.com/?%v`, queryParams.Encode())
	data, err := util.TakeScreenshot(ctx, url)
	if err != nil {
		return "", fmt.Errorf("failed to take the screenshot: %w", err)
	}

	loc, err := l.doOCR(ctx, data)
	if err != nil {
		return "", fmt.Errorf("failed to do OCR: %w", err)
	}

	return loc, nil
}

func (l *Locator) doOCR(ctx context.Context, buf []byte) (string, error) {
	// call the OpenAI API to parse the vacancy information
	compLoc, err := util.CallOpenAI[companyLocation](ctx, l.OpenaAiClient, []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage("You will be given screenshot of search result for company HQ location parse the company HQ location from it."),
		openai.UserMessageParts(openai.ImagePart(fmt.Sprintf("data:image/png;base64,%s", base64.StdEncoding.EncodeToString(buf)))),
	})
	if err != nil {
		return "", fmt.Errorf("unable to parse the vacancy information: %w", err)
	}

	return compLoc.Location, nil
}

type companyLocation struct {
	Location string `json:"location" jsonschema_description:"The company HQ location, the format must in the form of 'City, Country' for example 'Riyadh, Saudi Arabia'. If the location is not found just return empty string."`
}
