package token_test

import (
	"testing"

	"github.com/ghazlabs/idn-remote-entry/internal/server/driven/token"
	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeDecodeToken(t *testing.T) {
	secretKey := "secret"
	tokenizer, err := token.NewTokenizer(token.TokenizerConfig{SecretKey: secretKey})
	require.NoError(t, err)

	req := core.SubmitRequest{
		SubmissionType:  core.SubmitTypeManual,
		SubmissionEmail: "admin@gmail.com",
		Vacancy: core.Vacancy{
			JobTitle:         "Test Job",
			CompanyName:      "Test Company",
			CompanyLocation:  "Test Location",
			ShortDescription: "Test Description",
			RelevantTags:     []string{"test", "job"},
			ApplyURL:         "https://example.com/apply",
		},
	}

	tokenStr, err := tokenizer.EncodeRequest(req)
	require.NoError(t, err)

	decodedReq, err := tokenizer.DecodeToken(tokenStr)
	require.NoError(t, err)

	assert.Equal(t, req, decodedReq)
}
