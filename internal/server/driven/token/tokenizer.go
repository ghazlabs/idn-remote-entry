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
		"retries":          req.Retries,
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

	vacancy := t.mapVacancy(claims)

	req := core.SubmitRequest{
		SubmissionType:  core.SubmitType(claims["submission_type"].(string)),
		SubmissionEmail: claims["submission_email"].(string),
		Retries:         int(claims["retries"].(float64)),
		Vacancy:         vacancy,
	}

	return req, nil
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

func (t *Tokenizer) mapVacancy(claims jwt.MapClaims) core.Vacancy {
	var submissionType core.SubmitType = core.SubmitType(claims["submission_type"].(string))
	var vacancy core.Vacancy
	vacancyMap := claims["vacancy"].(map[string]interface{})

	if submissionType == core.SubmitTypeManual {
		var relevantTags []string
		if tags, ok := vacancyMap["relevant_tags"].([]interface{}); ok {
			relevantTags = make([]string, len(tags))
			for i, tag := range tags {
				if strTag, ok := tag.(string); ok {
					relevantTags[i] = strTag
				}
			}
		}

		vacancy = core.Vacancy{
			JobTitle:         vacancyMap["job_title"].(string),
			CompanyName:      vacancyMap["company_name"].(string),
			CompanyLocation:  vacancyMap["company_location"].(string),
			ShortDescription: vacancyMap["short_description"].(string),
			RelevantTags:     relevantTags,
			ApplyURL:         vacancyMap["apply_url"].(string),
		}
	}

	if submissionType == core.SubmitTypeURL {
		vacancy = core.Vacancy{
			ApplyURL: vacancyMap["apply_url"].(string),
		}
	}

	return vacancy
}
