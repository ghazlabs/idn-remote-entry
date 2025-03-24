package core_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ghazlabs/idn-remote-entry/internal/server/core"
	shcore "github.com/ghazlabs/idn-remote-entry/internal/shared/core"
)

func TestServiceHandleRequest(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(context.Context, *mockQueue, *mockEmailClient, *mockTokenizer, *mockApproval, *mockApprovalStorage)
		request    shcore.SubmitRequest
	}{
		{
			name: "submission type manual - submission email not whitelisted - approval needed",
			setupMocks: func(ctx context.Context, q *mockQueue, e *mockEmailClient, tok *mockTokenizer, a *mockApproval, s *mockApprovalStorage) {
				tok.On("EncodeRequest", mock.MatchedBy(func(req shcore.SubmitRequest) bool {
					return req.SubmissionEmail == "submitter@example.com"
				})).Return("mock-token", nil)
				a.On("NeedsApproval", "submitter@example.com").Return(true)
				e.On("SendApprovalRequest", ctx, mock.MatchedBy(func(req shcore.SubmitRequest) bool {
					return req.SubmissionEmail == "submitter@example.com"
				}), "mock-token").Return("mock-message-id", nil)
				s.On("SaveApprovalRequest", "mock-message-id", mock.MatchedBy(func(req shcore.SubmitRequest) bool {
					return req.SubmissionEmail == "submitter@example.com"
				})).Return(nil)
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
		},
		{
			name: "submission type url - submission email not whitelisted - approval needed",
			setupMocks: func(ctx context.Context, q *mockQueue, e *mockEmailClient, tok *mockTokenizer, a *mockApproval, s *mockApprovalStorage) {
				tok.On("EncodeRequest", mock.MatchedBy(func(req shcore.SubmitRequest) bool {
					return req.SubmissionEmail == "submitter@example.com"
				})).Return("mock-token", nil)
				a.On("NeedsApproval", "submitter@example.com").Return(true)
				e.On("SendApprovalRequest", ctx, mock.MatchedBy(func(req shcore.SubmitRequest) bool {
					return req.SubmissionEmail == "submitter@example.com"
				}), "mock-token").Return("mock-message-id", nil)
				s.On("SaveApprovalRequest", "mock-message-id", mock.MatchedBy(func(req shcore.SubmitRequest) bool {
					return req.SubmissionEmail == "submitter@example.com"
				})).Return(nil)
			},
			request: shcore.SubmitRequest{
				SubmissionType:  shcore.SubmitTypeManual,
				SubmissionEmail: "submitter@example.com",
				Vacancy: shcore.Vacancy{
					ApplyURL: "https://example.com/apply",
				},
			},
		},
		{
			name: "submission type manual - submission email is empty - approval needed",
			setupMocks: func(ctx context.Context, q *mockQueue, e *mockEmailClient, tok *mockTokenizer, a *mockApproval, s *mockApprovalStorage) {
				tok.On("EncodeRequest", mock.MatchedBy(func(req shcore.SubmitRequest) bool {
					return req.SubmissionEmail == ""
				})).Return("mock-token", nil)
				a.On("NeedsApproval", "").Return(true)
				e.On("SendApprovalRequest", ctx, mock.MatchedBy(func(req shcore.SubmitRequest) bool {
					return req.SubmissionEmail == ""
				}), "mock-token").Return("mock-message-id", nil)
				s.On("SaveApprovalRequest", "mock-message-id", mock.MatchedBy(func(req shcore.SubmitRequest) bool {
					return req.SubmissionEmail == ""
				})).Return(nil)
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
		},
		{
			name: "submission type url - submission email is empty - approval needed",
			setupMocks: func(ctx context.Context, q *mockQueue, e *mockEmailClient, tok *mockTokenizer, a *mockApproval, s *mockApprovalStorage) {
				tok.On("EncodeRequest", mock.MatchedBy(func(req shcore.SubmitRequest) bool {
					return req.SubmissionEmail == ""
				})).Return("mock-token", nil)
				a.On("NeedsApproval", "").Return(true)
				e.On("SendApprovalRequest", ctx, mock.MatchedBy(func(req shcore.SubmitRequest) bool {
					return req.SubmissionEmail == ""
				}), "mock-token").Return("mock-message-id", nil)
				s.On("SaveApprovalRequest", "mock-message-id", mock.MatchedBy(func(req shcore.SubmitRequest) bool {
					return req.SubmissionEmail == ""
				})).Return(nil)
			},
			request: shcore.SubmitRequest{
				SubmissionType:  shcore.SubmitTypeURL,
				SubmissionEmail: "",
				Vacancy: shcore.Vacancy{
					ApplyURL: "https://example.com/apply",
				},
			},
		},
		{
			name: "submission type manual - submission email is whitelisted - no approval needed",
			setupMocks: func(ctx context.Context, q *mockQueue, e *mockEmailClient, tok *mockTokenizer, a *mockApproval, s *mockApprovalStorage) {
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
		},
		{
			name: "submission type url - submission email is whitelisted - no approval needed",
			setupMocks: func(ctx context.Context, q *mockQueue, e *mockEmailClient, tok *mockTokenizer, a *mockApproval, s *mockApprovalStorage) {
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
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create context
			ctx := context.Background()

			// Initialize mocks
			mockQueue := &mockQueue{}
			mockEmail := &mockEmailClient{}
			mockTokenizer := &mockTokenizer{}
			mockApproval := &mockApproval{}
			mockApprovalStorage := &mockApprovalStorage{}

			service, err := core.NewService(core.ServiceConfig{
				Queue:           mockQueue,
				Email:           mockEmail,
				Tokenizer:       mockTokenizer,
				Approval:        mockApproval,
				ApprovalStorage: mockApprovalStorage,
			})
			require.NoError(t, err)

			tt.setupMocks(ctx, mockQueue, mockEmail, mockTokenizer, mockApproval, mockApprovalStorage)

			// Execute test
			err = service.HandleRequest(ctx, tt.request)
			require.NoError(t, err)

			// Verify mock expectations
			mockQueue.AssertExpectations(t)
			mockEmail.AssertExpectations(t)
			mockTokenizer.AssertExpectations(t)
			mockApproval.AssertExpectations(t)
		})
	}
}

type mockQueue struct {
	mock.Mock
}

func (m *mockQueue) Put(ctx context.Context, req shcore.SubmitRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

type mockEmailClient struct {
	mock.Mock
}

func (m *mockEmailClient) SendApprovalRequest(ctx context.Context, req shcore.SubmitRequest, token string) (string, error) {
	args := m.Called(ctx, req, token)
	return args.String(0), args.Error(1)
}

func (m *mockEmailClient) ApproveRequest(ctx context.Context, messageID string) error {
	args := m.Called(ctx, messageID)
	return args.Error(0)
}

func (m *mockEmailClient) RejectRequest(ctx context.Context, messageID string) error {
	args := m.Called(ctx, messageID)
	return args.Error(0)
}

type mockTokenizer struct {
	mock.Mock
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
}

func (m *mockApproval) NeedsApproval(email string) bool {
	args := m.Called(email)
	return args.Bool(0)
}

type mockApprovalStorage struct {
	mock.Mock
}

func (m *mockApprovalStorage) GetApprovalState(messageID string) (core.ApprovalState, error) {
	args := m.Called(messageID)
	return args.Get(0).(core.ApprovalState), args.Error(1)
}

func (m *mockApprovalStorage) UpdateApprovalState(messageID string, state core.ApprovalState) error {
	args := m.Called(messageID, state)
	return args.Error(0)
}

func (m *mockApprovalStorage) SaveApprovalRequest(messageID string, req shcore.SubmitRequest) error {
	args := m.Called(messageID, req)
	return args.Error(0)
}
