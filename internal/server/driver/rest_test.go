package driver_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ghazlabs/idn-remote-entry/internal/server/core"
	"github.com/ghazlabs/idn-remote-entry/internal/server/driver"
	shcore "github.com/ghazlabs/idn-remote-entry/internal/shared/core"
)

func TestAPIServeSubmitVacancy(t *testing.T) {
	tests := []struct {
		name              string
		apiKey            string
		requestBody       interface{}
		serviceResponse   error
		expectedStatus    int
		expectedErrorCode string
		validateParams    func(t *testing.T, req *shcore.SubmitRequest)
	}{
		{
			name:   "success with manual submission",
			apiKey: "test-api-key",
			requestBody: shcore.SubmitRequest{
				SubmissionType:  shcore.SubmitTypeManual,
				SubmissionEmail: "test@example.com",
				Vacancy: shcore.Vacancy{
					JobTitle:         "Software Engineer",
					CompanyName:      "Test Company",
					CompanyLocation:  "Test Location",
					ShortDescription: "Test Description",
					RelevantTags:     []string{"tag1", "tag2"},
					ApplyURL:         "https://example.com/apply",
				},
			},
			serviceResponse: nil,
			expectedStatus:  http.StatusOK,
			validateParams: func(t *testing.T, req *shcore.SubmitRequest) {
				require.NotNil(t, req)
				assert.Equal(t, shcore.SubmitTypeManual, req.SubmissionType)
				assert.Equal(t, "test@example.com", req.SubmissionEmail)
				assert.Equal(t, "Software Engineer", req.Vacancy.JobTitle)
				assert.Equal(t, "Test Company", req.Vacancy.CompanyName)
				assert.Equal(t, "Test Location", req.Vacancy.CompanyLocation)
				assert.Equal(t, "Test Description", req.Vacancy.ShortDescription)
				assert.Equal(t, []string{"tag1", "tag2"}, req.Vacancy.RelevantTags)
				assert.Equal(t, "https://example.com/apply", req.Vacancy.ApplyURL)
			},
		},
		{
			name:   "success with manual submission - with invalid tags",
			apiKey: "test-api-key",
			requestBody: shcore.SubmitRequest{
				SubmissionType:  shcore.SubmitTypeManual,
				SubmissionEmail: "test@example.com",
				Vacancy: shcore.Vacancy{
					JobTitle:         "Software Engineer",
					CompanyName:      "Test Company",
					CompanyLocation:  "Test Location",
					ShortDescription: "Test Description",
					RelevantTags:     []string{"tag1", "tag2", "", " "},
					ApplyURL:         "https://example.com/apply",
				},
			},
			serviceResponse: nil,
			expectedStatus:  http.StatusOK,
			validateParams: func(t *testing.T, req *shcore.SubmitRequest) {
				require.NotNil(t, req)
				assert.Equal(t, shcore.SubmitTypeManual, req.SubmissionType)
				assert.Equal(t, "test@example.com", req.SubmissionEmail)
				assert.Equal(t, "Software Engineer", req.Vacancy.JobTitle)
				assert.Equal(t, "Test Company", req.Vacancy.CompanyName)
				assert.Equal(t, "Test Location", req.Vacancy.CompanyLocation)
				assert.Equal(t, "Test Description", req.Vacancy.ShortDescription)
				assert.Equal(t, []string{"tag1", "tag2"}, req.Vacancy.RelevantTags)
				assert.Equal(t, "https://example.com/apply", req.Vacancy.ApplyURL)
			},
		},
		{
			name:   "success with URL submission",
			apiKey: "test-api-key",
			requestBody: shcore.SubmitRequest{
				SubmissionType:  shcore.SubmitTypeURL,
				SubmissionEmail: "recruiter@company.com",
				Vacancy: shcore.Vacancy{
					ApplyURL: "https://jobs.company.com/123",
				},
			},
			serviceResponse: nil,
			expectedStatus:  http.StatusOK,
			validateParams: func(t *testing.T, req *shcore.SubmitRequest) {
				require.NotNil(t, req)
				assert.Equal(t, shcore.SubmitTypeURL, req.SubmissionType)
				assert.Equal(t, "recruiter@company.com", req.SubmissionEmail)
				assert.Equal(t, "https://jobs.company.com/123", req.Vacancy.ApplyURL)
			},
		},
		{
			name:   "success with bulk submission",
			apiKey: "test-api-key",
			requestBody: shcore.SubmitRequest{
				SubmissionType:  shcore.SubmitTypeBulk,
				SubmissionEmail: "crawler@system.com",
				BulkVacancies: []shcore.Vacancy{
					{
						JobTitle:    "Frontend Developer",
						CompanyName: "Company A",
						ApplyURL:    "https://example.com/jobs/1",
					},
					{
						JobTitle:    "Backend Engineer",
						CompanyName: "Company B",
						ApplyURL:    "https://example.com/jobs/2",
					},
				},
			},
			serviceResponse: nil,
			expectedStatus:  http.StatusOK,
			validateParams: func(t *testing.T, req *shcore.SubmitRequest) {
				require.NotNil(t, req)
				assert.Equal(t, shcore.SubmitTypeBulk, req.SubmissionType)
				assert.Equal(t, "crawler@system.com", req.SubmissionEmail)
				assert.Len(t, req.BulkVacancies, 2)
				assert.Equal(t, "Frontend Developer", req.BulkVacancies[0].JobTitle)
				assert.Equal(t, "Company A", req.BulkVacancies[0].CompanyName)
				assert.Equal(t, "https://example.com/jobs/1", req.BulkVacancies[0].ApplyURL)
				assert.Equal(t, "Backend Engineer", req.BulkVacancies[1].JobTitle)
			},
		},
		{
			name:              "invalid api key",
			apiKey:            "invalid-api-key",
			requestBody:       shcore.SubmitRequest{},
			serviceResponse:   nil,
			expectedStatus:    http.StatusUnauthorized,
			expectedErrorCode: "ERR_INVALID_API_KEY",
			validateParams:    nil, // No validation needed as service won't be called
		},
		{
			name:              "invalid request body",
			apiKey:            "test-api-key",
			requestBody:       "invalid json",
			serviceResponse:   nil,
			expectedStatus:    http.StatusBadRequest,
			expectedErrorCode: "ERR_BAD_REQUEST",
			validateParams:    nil, // No validation needed as service won't be called
		},
		{
			name:   "service returns error",
			apiKey: "test-api-key",
			requestBody: shcore.SubmitRequest{
				SubmissionType: shcore.SubmitTypeManual,
				Vacancy:        shcore.Vacancy{ApplyURL: "https://example.com/apply"},
			},
			serviceResponse:   shcore.NewBadRequestError("validation failed"),
			expectedStatus:    http.StatusBadRequest,
			expectedErrorCode: "ERR_BAD_REQUEST",
			validateParams: func(t *testing.T, req *shcore.SubmitRequest) {
				require.NotNil(t, req)
				assert.Equal(t, shcore.SubmitTypeManual, req.SubmissionType)
				assert.Equal(t, "https://example.com/apply", req.Vacancy.ApplyURL)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create stub service with desired response
			service := newStubService()
			service.handleRequestFunc = func(ctx context.Context, req shcore.SubmitRequest) error {
				return tt.serviceResponse
			}

			api, err := driver.NewAPI(driver.APIConfig{
				Service:      service,
				ClientApiKey: "test-api-key",
			})
			require.NoError(t, err)

			// Create request
			var body bytes.Buffer
			if s, ok := tt.requestBody.(string); ok {
				body.WriteString(s)
			} else {
				err := json.NewEncoder(&body).Encode(tt.requestBody)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/vacancies", &body)
			req.Header.Set("X-Api-Key", tt.apiKey)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Call the handler
			handler := api.GetHandler()
			handler.ServeHTTP(w, req)

			// Check the status code
			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			// Parse the response
			var respBody driver.RespBody
			err = json.NewDecoder(resp.Body).Decode(&respBody)
			require.NoError(t, err)

			if tt.expectedErrorCode != "" {
				assert.Equal(t, tt.expectedErrorCode, respBody.Err)
				assert.False(t, respBody.OK)
			} else {
				assert.True(t, respBody.OK)
			}

			// Validate the parameters passed to the service
			if tt.validateParams != nil && service.lastSubmitRequest != nil {
				tt.validateParams(t, service.lastSubmitRequest)
			}
		})
	}
}

func TestAPIServeApproveVacancy(t *testing.T) {
	tests := []struct {
		name              string
		token             string
		messageID         string
		serviceResponse   error
		expectedStatus    int
		expectedErrorCode string
		validateParams    func(t *testing.T, req *core.ApprovalRequest)
	}{
		{
			name:            "success",
			token:           "valid-token",
			messageID:       "message-123",
			serviceResponse: nil,
			expectedStatus:  http.StatusOK,
			validateParams: func(t *testing.T, req *core.ApprovalRequest) {
				require.NotNil(t, req)
				assert.Equal(t, "valid-token", req.TokenRequest)
				assert.Equal(t, "message-123", req.MessageID)
			},
		},
		{
			name:              "missing token",
			token:             "",
			messageID:         "message-123",
			serviceResponse:   nil,
			expectedStatus:    http.StatusBadRequest,
			expectedErrorCode: "ERR_BAD_REQUEST",
			validateParams:    nil, // No validation needed as service won't be called
		},
		{
			name:              "service returns error",
			token:             "invalid-token",
			messageID:         "message-123",
			serviceResponse:   createBadRequestError("invalid token"),
			expectedStatus:    http.StatusBadRequest,
			expectedErrorCode: "ERR_BAD_REQUEST",
			validateParams: func(t *testing.T, req *core.ApprovalRequest) {
				require.NotNil(t, req)
				assert.Equal(t, "invalid-token", req.TokenRequest)
				assert.Equal(t, "message-123", req.MessageID)
			},
		},
		{
			name:            "token only, no message ID",
			token:           "token-only",
			messageID:       "",
			serviceResponse: nil,
			expectedStatus:  http.StatusOK,
			validateParams: func(t *testing.T, req *core.ApprovalRequest) {
				require.NotNil(t, req)
				assert.Equal(t, "token-only", req.TokenRequest)
				assert.Equal(t, "", req.MessageID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create stub service with desired response
			service := newStubService()
			service.handleApproveFunc = func(ctx context.Context, req core.ApprovalRequest) error {
				return tt.serviceResponse
			}

			api, err := driver.NewAPI(driver.APIConfig{
				Service:      service,
				ClientApiKey: "test-api-key",
			})
			require.NoError(t, err)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/vacancies/approve?data="+tt.token+"&message_id="+tt.messageID, nil)
			w := httptest.NewRecorder()

			// Call the handler
			handler := api.GetHandler()
			handler.ServeHTTP(w, req)

			// Check the status code
			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			// Parse the response
			var respBody driver.RespBody
			err = json.NewDecoder(resp.Body).Decode(&respBody)
			require.NoError(t, err)

			if tt.expectedErrorCode != "" {
				assert.Equal(t, tt.expectedErrorCode, respBody.Err)
				assert.False(t, respBody.OK)
			} else {
				assert.True(t, respBody.OK)
			}

			// Validate the parameters passed to the service
			if tt.validateParams != nil && service.lastApprovalRequest != nil {
				tt.validateParams(t, service.lastApprovalRequest)
			}
		})
	}
}

func TestAPIServeRejectVacancy(t *testing.T) {
	tests := []struct {
		name              string
		token             string
		messageID         string
		serviceResponse   error
		expectedStatus    int
		expectedErrorCode string
		validateParams    func(t *testing.T, req *core.ApprovalRequest)
	}{
		{
			name:            "success",
			token:           "valid-token",
			messageID:       "message-123",
			serviceResponse: nil,
			expectedStatus:  http.StatusOK,
			validateParams: func(t *testing.T, req *core.ApprovalRequest) {
				require.NotNil(t, req)
				assert.Equal(t, "valid-token", req.TokenRequest)
				assert.Equal(t, "message-123", req.MessageID)
			},
		},
		{
			name:              "missing token",
			token:             "",
			messageID:         "message-123",
			serviceResponse:   nil,
			expectedStatus:    http.StatusBadRequest,
			expectedErrorCode: "ERR_BAD_REQUEST",
			validateParams:    nil, // No validation needed as service won't be called
		},
		{
			name:              "service returns error",
			token:             "invalid-token",
			messageID:         "message-123",
			serviceResponse:   createBadRequestError("invalid token"),
			expectedStatus:    http.StatusBadRequest,
			expectedErrorCode: "ERR_BAD_REQUEST",
			validateParams: func(t *testing.T, req *core.ApprovalRequest) {
				require.NotNil(t, req)
				assert.Equal(t, "invalid-token", req.TokenRequest)
				assert.Equal(t, "message-123", req.MessageID)
			},
		},
		{
			name:            "token only, no message ID",
			token:           "token-only",
			messageID:       "",
			serviceResponse: nil,
			expectedStatus:  http.StatusOK,
			validateParams: func(t *testing.T, req *core.ApprovalRequest) {
				require.NotNil(t, req)
				assert.Equal(t, "token-only", req.TokenRequest)
				assert.Equal(t, "", req.MessageID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create stub service with desired response
			service := newStubService()
			service.handleRejectFunc = func(ctx context.Context, req core.ApprovalRequest) error {
				return tt.serviceResponse
			}

			api, err := driver.NewAPI(driver.APIConfig{
				Service:      service,
				ClientApiKey: "test-api-key",
			})
			require.NoError(t, err)

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/vacancies/reject?data="+tt.token+"&message_id="+tt.messageID, nil)
			w := httptest.NewRecorder()

			// Call the handler
			handler := api.GetHandler()
			handler.ServeHTTP(w, req)

			// Check the status code
			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			// Parse the response
			var respBody driver.RespBody
			err = json.NewDecoder(resp.Body).Decode(&respBody)
			require.NoError(t, err)

			if tt.expectedErrorCode != "" {
				assert.Equal(t, tt.expectedErrorCode, respBody.Err)
				assert.False(t, respBody.OK)
			} else {
				assert.True(t, respBody.OK)
			}

			// Validate the parameters passed to the service
			if tt.validateParams != nil && service.lastRejectRequest != nil {
				tt.validateParams(t, service.lastRejectRequest)
			}
		})
	}
}

// stubService is a simple implementation of core.Service for testing
type stubService struct {
	handleRequestFunc func(ctx context.Context, req shcore.SubmitRequest) error
	handleApproveFunc func(ctx context.Context, req core.ApprovalRequest) error
	handleRejectFunc  func(ctx context.Context, req core.ApprovalRequest) error
	// Capture last received parameters for validation
	lastSubmitRequest   *shcore.SubmitRequest
	lastApprovalRequest *core.ApprovalRequest
	lastRejectRequest   *core.ApprovalRequest
}

func newStubService() *stubService {
	return &stubService{
		handleRequestFunc: func(ctx context.Context, req shcore.SubmitRequest) error {
			return nil
		},
		handleApproveFunc: func(ctx context.Context, req core.ApprovalRequest) error {
			return nil
		},
		handleRejectFunc: func(ctx context.Context, req core.ApprovalRequest) error {
			return nil
		},
	}
}

func (s *stubService) HandleRequest(ctx context.Context, req shcore.SubmitRequest) error {
	// Capture request for later inspection
	s.lastSubmitRequest = &req
	return s.handleRequestFunc(ctx, req)
}

func (s *stubService) HandleApprove(ctx context.Context, req core.ApprovalRequest) error {
	// Capture request for later inspection
	s.lastApprovalRequest = &req
	return s.handleApproveFunc(ctx, req)
}

func (s *stubService) HandleReject(ctx context.Context, req core.ApprovalRequest) error {
	// Capture request for later inspection
	s.lastRejectRequest = &req
	return s.handleRejectFunc(ctx, req)
}

// CreateBadRequestError is a helper to create error similar to core.NewBadRequestError
func createBadRequestError(message string) error {
	return shcore.NewBadRequestError(message)
}
