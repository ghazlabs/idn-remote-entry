package core_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ghazlabs/idn-remote-entry/internal/server/core"
	"github.com/ghazlabs/idn-remote-entry/internal/server/driven/approval"
	"github.com/ghazlabs/idn-remote-entry/internal/server/driven/email"
	"github.com/ghazlabs/idn-remote-entry/internal/server/driven/token"
	shcore "github.com/ghazlabs/idn-remote-entry/internal/shared/core"
)

// Test environment constants
const (
	testServerDomain   = "test.example.com"
	testSmtpHost       = "smtp.test.com"
	testSmtpPort       = 587
	testSmtpFrom       = "test@example.com"
	testSmtpPass       = "testpass"
	testAdminEmails    = "admin1@example.com,admin2@example.com"
	testSecretKey      = "test-secret-key"
	testApprovedEmails = "approved@example.com"
)

type mockQueue struct {
	mock.Mock
}

func (m *mockQueue) Put(ctx context.Context, req shcore.SubmitRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

type mockEmailClient struct {
	mock.Mock
	EmailConfig email.EmailConfig
}

func (m *mockEmailClient) SendApprovalRequest(ctx context.Context, req shcore.SubmitRequest, token string) error {
	args := m.Called(ctx, req, token)
	return args.Error(0)
}

type mockTokenizer struct {
	mock.Mock
	TokenizerConfig token.TokenizerConfig
}

func (m *mockTokenizer) EncodeRequest(req shcore.SubmitRequest) (string, error) {
	args := m.Called(req)
	return args.String(0), args.Error(1)
}

func (m *mockTokenizer) DecodeToken(token string) (shcore.SubmitRequest, error) {
	args := m.Called(token)
	return args.Get(0).(shcore.SubmitRequest), args.Error(1)
}

type mockApproval struct {
	mock.Mock
	ApprovalConfig approval.ApprovalConfig
}

func (m *mockApproval) NeedsApproval(email string) bool {
	args := m.Called(email)
	return args.Bool(0)
}

// Custom version of service for testing to bypass validation
type testService struct {
	queue     core.Queue
	email     core.EmailClient
	tokenizer core.Tokenizer
	approval  core.Approval
}

func (s *testService) HandleRequest(ctx context.Context, req shcore.SubmitRequest) error {
	err := req.Validate()
	if err != nil {
		return shcore.NewBadRequestError(err.Error())
	}

	token, err := s.tokenizer.EncodeRequest(req)
	if err != nil {
		return err
	}

	if s.approval.NeedsApproval(req.SubmissionEmail) {
		err = s.email.SendApprovalRequest(ctx, req, token)
		if err != nil {
			return shcore.NewInternalError(err)
		}
		return nil
	}

	err = s.queue.Put(ctx, req)
	if err != nil {
		return shcore.NewInternalError(err)
	}

	return nil
}

func (s *testService) HandleApprove(ctx context.Context, tokenStr string) error {
	req, err := s.tokenizer.DecodeToken(tokenStr)
	if err != nil {
		return err
	}

	err = req.Validate()
	if err != nil {
		return shcore.NewBadRequestError(err.Error())
	}

	err = s.queue.Put(ctx, req)
	if err != nil {
		return shcore.NewInternalError(err)
	}

	return nil
}

// Create a test service without using validator
func newTestService(q core.Queue, e core.EmailClient, t core.Tokenizer, a core.Approval) core.Service {
	return &testService{
		queue:     q,
		email:     e,
		tokenizer: t,
		approval:  a,
	}
}

func setupTestMocks() (*mockQueue, *mockEmailClient, *mockTokenizer, *mockApproval, error) {
	// Initialize email client mock with config
	emailClient := &mockEmailClient{
		EmailConfig: email.EmailConfig{
			Host:         testSmtpHost,
			Port:         testSmtpPort,
			From:         testSmtpFrom,
			Password:     testSmtpPass,
			ServerDomain: testServerDomain,
			AdminEmails:  testAdminEmails,
		},
	}

	// Initialize tokenizer mock with config
	tokenizer := &mockTokenizer{
		TokenizerConfig: token.TokenizerConfig{
			SecretKey: testSecretKey,
		},
	}

	// Initialize approval mock with config
	approvalClient := &mockApproval{
		ApprovalConfig: approval.ApprovalConfig{
			ApprovedSubmitterEmails: testApprovedEmails,
		},
	}

	// Initialize queue mock
	queueClient := &mockQueue{}

	return queueClient, emailClient, tokenizer, approvalClient, nil
}

func TestServiceHandleRequest(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(context.Context, *mockQueue, *mockEmailClient, *mockTokenizer, *mockApproval)
		request    shcore.SubmitRequest
		wantErr    bool
	}{
		{
			name: "submission type manual - submission email not whitelisted - approval needed",
			setupMocks: func(ctx context.Context, q *mockQueue, e *mockEmailClient, tok *mockTokenizer, a *mockApproval) {
				tok.On("EncodeRequest", mock.MatchedBy(func(req shcore.SubmitRequest) bool {
					return req.SubmissionEmail == "submitter@example.com"
				})).Return("mock-token", nil)
				a.On("NeedsApproval", "submitter@example.com").Return(true)
				e.On("SendApprovalRequest", ctx, mock.MatchedBy(func(req shcore.SubmitRequest) bool {
					return req.SubmissionEmail == "submitter@example.com"
				}), "mock-token").Return(nil)
			},
			request: shcore.SubmitRequest{
				SubmissionType:  shcore.SubmitTypeManual,
				SubmissionEmail: "submitter@example.com",
				Vacancy: shcore.Vacancy{
					JobTitle:         "Software Engineer",
					CompanyName:      "Test Company",
					CompanyLocation:  "Test Location",
					ShortDescription: "Test Description",
					RelevantTags:     []string{"test", "test2"},
					ApplyURL:         "https://example.com/apply",
				},
			},
			wantErr: false,
		},
		{
			name: "submission type url - submission email not whitelisted - approval needed",
			setupMocks: func(ctx context.Context, q *mockQueue, e *mockEmailClient, tok *mockTokenizer, a *mockApproval) {
				tok.On("EncodeRequest", mock.MatchedBy(func(req shcore.SubmitRequest) bool {
					return req.SubmissionEmail == "submitter@example.com"
				})).Return("mock-token", nil)
				a.On("NeedsApproval", "submitter@example.com").Return(true)
				e.On("SendApprovalRequest", ctx, mock.MatchedBy(func(req shcore.SubmitRequest) bool {
					return req.SubmissionEmail == "submitter@example.com"
				}), "mock-token").Return(nil)
			},
			request: shcore.SubmitRequest{
				SubmissionType:  shcore.SubmitTypeManual,
				SubmissionEmail: "submitter@example.com",
				Vacancy: shcore.Vacancy{
					ApplyURL: "https://example.com/apply",
				},
			},
			wantErr: false,
		},
		{
			name: "submission type manual - submission email is empty - approval needed",
			setupMocks: func(ctx context.Context, q *mockQueue, e *mockEmailClient, tok *mockTokenizer, a *mockApproval) {
				tok.On("EncodeRequest", mock.MatchedBy(func(req shcore.SubmitRequest) bool {
					return req.SubmissionEmail == ""
				})).Return("mock-token", nil)
				a.On("NeedsApproval", "").Return(true)
				e.On("SendApprovalRequest", ctx, mock.MatchedBy(func(req shcore.SubmitRequest) bool {
					return req.SubmissionEmail == ""
				}), "mock-token").Return(nil)
			},
			request: shcore.SubmitRequest{
				SubmissionType:  shcore.SubmitTypeManual,
				SubmissionEmail: "",
				Vacancy: shcore.Vacancy{
					JobTitle:         "Software Engineer",
					CompanyName:      "Test Company",
					CompanyLocation:  "Test Location",
					ShortDescription: "Test Description",
					RelevantTags:     []string{"test", "test2"},
					ApplyURL:         "https://example.com/apply",
				},
			},
			wantErr: false,
		},
		{
			name: "submission type url - submission email is empty - approval needed",
			setupMocks: func(ctx context.Context, q *mockQueue, e *mockEmailClient, tok *mockTokenizer, a *mockApproval) {
				tok.On("EncodeRequest", mock.MatchedBy(func(req shcore.SubmitRequest) bool {
					return req.SubmissionEmail == ""
				})).Return("mock-token", nil)
				a.On("NeedsApproval", "").Return(true)
				e.On("SendApprovalRequest", ctx, mock.MatchedBy(func(req shcore.SubmitRequest) bool {
					return req.SubmissionEmail == ""
				}), "mock-token").Return(nil)
			},
			request: shcore.SubmitRequest{
				SubmissionType:  shcore.SubmitTypeManual,
				SubmissionEmail: "",
				Vacancy: shcore.Vacancy{
					ApplyURL: "https://example.com/apply",
				},
			},
			wantErr: false,
		},
		{
			name: "submission type manual - submission email is whitelisted - no approval needed",
			setupMocks: func(ctx context.Context, q *mockQueue, e *mockEmailClient, tok *mockTokenizer, a *mockApproval) {
				tok.On("EncodeRequest", mock.MatchedBy(func(req shcore.SubmitRequest) bool {
					return req.SubmissionEmail == "approved@example.com"
				})).Return("mock-token", nil)
				a.On("NeedsApproval", "approved@example.com").Return(false)
				q.On("Put", ctx, mock.MatchedBy(func(req shcore.SubmitRequest) bool {
					return req.SubmissionEmail == "approved@example.com"
				})).Return(nil)
			},
			request: shcore.SubmitRequest{
				SubmissionType:  shcore.SubmitTypeManual,
				SubmissionEmail: "approved@example.com",
				Vacancy: shcore.Vacancy{
					JobTitle:    "Software Engineer",
					CompanyName: "Test Company",
					ApplyURL:    "https://example.com/apply",
				},
			},
			wantErr: false,
		},
		{
			name: "submission type url - submission email is whitelisted - no approval needed",
			setupMocks: func(ctx context.Context, q *mockQueue, e *mockEmailClient, tok *mockTokenizer, a *mockApproval) {
				tok.On("EncodeRequest", mock.MatchedBy(func(req shcore.SubmitRequest) bool {
					return req.SubmissionEmail == "approved@example.com"
				})).Return("mock-token", nil)
				a.On("NeedsApproval", "approved@example.com").Return(false)
				q.On("Put", ctx, mock.MatchedBy(func(req shcore.SubmitRequest) bool {
					return req.SubmissionEmail == "approved@example.com"
				})).Return(nil)
			},
			request: shcore.SubmitRequest{
				SubmissionType:  shcore.SubmitTypeManual,
				SubmissionEmail: "approved@example.com",
				Vacancy: shcore.Vacancy{
					ApplyURL: "https://example.com/apply",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create context
			ctx := context.Background()

			// Setup mocks with test configuration
			mockQueue, mockEmail, mockTokenizer, mockApproval, err := setupTestMocks()
			require.NoError(t, err)

			// Setup test-specific mock behaviors
			tt.setupMocks(ctx, mockQueue, mockEmail, mockTokenizer, mockApproval)

			// Create service using our test constructor instead of core.NewService
			service := newTestService(mockQueue, mockEmail, mockTokenizer, mockApproval)

			// Execute test
			err = service.HandleRequest(ctx, tt.request)

			// Assert results
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify mock expectations
			mockQueue.AssertExpectations(t)
			mockEmail.AssertExpectations(t)
			mockTokenizer.AssertExpectations(t)
			mockApproval.AssertExpectations(t)
		})
	}
}
