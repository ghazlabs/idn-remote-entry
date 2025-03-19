package token

import (
	"encoding/base64"
	"fmt"

	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	jwt "github.com/golang-jwt/jwt/v5"
	"gopkg.in/validator.v2"
)

type TokenizerConfig struct {
	SecretKey string `validate:"nonzero"`
}

type Tokenizer struct {
	TokenizerConfig
}

func NewTokenizer(cfg TokenizerConfig) (*Tokenizer, error) {
	err := validator.Validate(cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &Tokenizer{
		TokenizerConfig: cfg,
	}, nil
}

func (t *Tokenizer) EncodeRequest(req core.SubmitRequest) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"submission_type":  req.SubmissionType,
		"submission_email": req.SubmissionEmail,
		"vacancy":          req.Vacancy,
	})
	tokenString, err := token.SignedString([]byte(t.SecretKey))
	if err != nil {
		return "", fmt.Errorf("failed to sign token due to: %w", err)
	}

	encodedToken := base64.URLEncoding.EncodeToString([]byte(tokenString))
	return encodedToken, nil
}

func (t *Tokenizer) DecodeToken(tokenStr string) (core.SubmitRequest, error) {
	claims, err := t.parseJWT(tokenStr)
	if err != nil {
		return core.SubmitRequest{}, fmt.Errorf("failed to parse token due to: %w", err)
	}

	return t.mapReqVacancy(claims), nil
}

func (t *Tokenizer) parseJWT(tokenStr string) (jwt.MapClaims, error) {
	decoded, err := base64.URLEncoding.DecodeString(tokenStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 token: %w", err)
	}

	token, err := jwt.Parse(string(decoded), func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(t.SecretKey), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	return token.Claims.(jwt.MapClaims), nil
}

func (t *Tokenizer) mapReqVacancy(claims jwt.MapClaims) core.SubmitRequest {
	submissionType := core.SubmitType(claims["submission_type"].(string))

	var vacancy core.Vacancy
	if vacancyMap, ok := claims["vacancy"].(map[string]interface{}); ok {
		submissionType := core.SubmitType(getString(claims, "submission_type"))

		vacancy = core.Vacancy{
			ApplyURL: getString(vacancyMap, "apply_url"),
		}

		if submissionType == core.SubmitTypeManual {
			vacancy.JobTitle = getString(vacancyMap, "job_title")
			vacancy.CompanyName = getString(vacancyMap, "company_name")
			vacancy.CompanyLocation = getString(vacancyMap, "company_location")
			vacancy.ShortDescription = getString(vacancyMap, "short_description")
			vacancy.RelevantTags = getStringSlice(vacancyMap, "relevant_tags")
		}
	}

	return core.SubmitRequest{
		SubmissionType:  submissionType,
		SubmissionEmail: getString(claims, "submission_email"),
		Vacancy:         vacancy,
	}
}

func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func getStringSlice(m map[string]interface{}, key string) []string {
	if tags, ok := m[key].([]interface{}); ok {
		result := make([]string, 0, len(tags))
		for _, tag := range tags {
			if strTag, ok := tag.(string); ok {
				result = append(result, strTag)
			}
		}
		return result
	}
	return nil
}
