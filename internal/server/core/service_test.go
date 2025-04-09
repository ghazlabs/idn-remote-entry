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
			name: "submission type bulk",
			setupMocks: func(ctx context.Context, q *mockQueue, e *mockEmailClient, tok *mockTokenizer, a *mockApproval, s *mockApprovalStorage) {
				tok.On("EncodeRequest", mock.MatchedBy(func(req shcore.SubmitRequest) bool {
					return req.JobTitle == "Test 1"
				})).Return("mock-token1", nil)
				tok.On("EncodeRequest", mock.MatchedBy(func(req shcore.SubmitRequest) bool {
					return req.JobTitle == "Test 2"
				})).Return("mock-token2", nil)
				e.On("SendBulkApprovalRequest", ctx, mock.MatchedBy(func(req shcore.SubmitRequest) bool {
					return req.SubmissionEmail == "crawler"
				}), []string{"mock-token1", "mock-token2"}).Return([]string{"mock-message-id1", "mock-message-id2"}, nil)
				s.On("SaveBulkApprovalRequest", ctx, mock.MatchedBy(func(req shcore.SubmitRequest) bool {
					return req.SubmissionEmail == "crawler"
				}), []string{"mock-message-id1", "mock-message-id2"}).Return(nil)
			},
			request: shcore.SubmitRequest{
				SubmissionType:  shcore.SubmitTypeBulk,
				SubmissionEmail: "crawler",
				BulkVacancy: []shcore.Vacancy{
					{
						JobTitle:    "Test 1",
						CompanyName: "Test 1 Company",
						ApplyURL:    "https://example.com/apply",
					},
					{
						JobTitle:    "Test 2",
						CompanyName: "Test 2 Company",
						ApplyURL:    "https://example.com/apply",
					},
				},
			},
		},
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
				s.On("SaveApprovalRequest", ctx, "mock-message-id", mock.MatchedBy(func(req shcore.SubmitRequest) bool {
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
				s.On("SaveApprovalRequest", ctx, "mock-message-id", mock.MatchedBy(func(req shcore.SubmitRequest) bool {
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
				s.On("SaveApprovalRequest", ctx, "mock-message-id", mock.MatchedBy(func(req shcore.SubmitRequest) bool {
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
				s.On("SaveApprovalRequest", ctx, "mock-message-id", mock.MatchedBy(func(req shcore.SubmitRequest) bool {
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

func TestServiceHandleApprove(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(context.Context, *mockQueue, *mockEmailClient, *mockTokenizer, *mockApprovalStorage)
		request    core.ApprovalRequest
		wantErr    bool
		errMsg     string
	}{
		{
			name: "successful approval with message ID",
			setupMocks: func(ctx context.Context, q *mockQueue, e *mockEmailClient, tok *mockTokenizer, s *mockApprovalStorage) {
				req := shcore.SubmitRequest{
					SubmissionEmail: "test@example.com",
					Vacancy: shcore.Vacancy{
						JobTitle: "Test Job",
					},
				}
				tok.On("DecodeToken", "test-token").Return(req, nil)
				s.On("GetApprovalState", ctx, "test-message").Return(core.ApprovalStatePending, nil)
				e.On("ApproveRequest", ctx, "test-message").Return(nil)
				s.On("UpdateApprovalState", ctx, "test-message", core.ApprovalStateApproved).Return(nil)
				q.On("Put", ctx, req).Return(nil)
			},
			request: core.ApprovalRequest{
				MessageID:    "test-message",
				TokenRequest: "test-token",
			},
		},
		{
			name: "successful approval without message ID",
			setupMocks: func(ctx context.Context, q *mockQueue, e *mockEmailClient, tok *mockTokenizer, s *mockApprovalStorage) {
				req := shcore.SubmitRequest{
					SubmissionEmail: "test@example.com",
					Vacancy: shcore.Vacancy{
						JobTitle: "Test Job",
					},
				}
				tok.On("DecodeToken", "test-token").Return(req, nil)
				q.On("Put", ctx, req).Return(nil)
			},
			request: core.ApprovalRequest{
				TokenRequest: "test-token",
			},
		},
		{
			name: "approval already processed",
			setupMocks: func(ctx context.Context, q *mockQueue, e *mockEmailClient, tok *mockTokenizer, s *mockApprovalStorage) {
				req := shcore.SubmitRequest{
					SubmissionEmail: "test@example.com",
				}
				tok.On("DecodeToken", "test-token").Return(req, nil)
				s.On("GetApprovalState", ctx, "test-message").Return(core.ApprovalStateApproved, nil)
			},
			request: core.ApprovalRequest{
				MessageID:    "test-message",
				TokenRequest: "test-token",
			},
			wantErr: true,
			errMsg:  "approval already processed",
		},
		{
			name: "invalid token",
			setupMocks: func(ctx context.Context, q *mockQueue, e *mockEmailClient, tok *mockTokenizer, s *mockApprovalStorage) {
				tok.On("DecodeToken", "invalid-token").Return(shcore.SubmitRequest{}, shcore.NewBadRequestError("invalid token"))
			},
			request: core.ApprovalRequest{
				MessageID:    "test-message",
				TokenRequest: "invalid-token",
			},
			wantErr: true,
			errMsg:  "invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			queue := &mockQueue{}
			email := &mockEmailClient{}
			tokenizer := &mockTokenizer{}
			approvalStorage := &mockApprovalStorage{}

			svc, err := core.NewService(core.ServiceConfig{
				Queue:           queue,
				Email:           email,
				Tokenizer:       tokenizer,
				Approval:        &mockApproval{}, // not used
				ApprovalStorage: approvalStorage,
			})
			require.NoError(t, err)

			tt.setupMocks(ctx, queue, email, tokenizer, approvalStorage)

			err = svc.HandleApprove(ctx, tt.request)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
				return
			}

			require.NoError(t, err)
			mock.AssertExpectationsForObjects(t, queue, email, tokenizer, approvalStorage)
		})
	}
}

func TestServiceHandleReject(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(context.Context, *mockEmailClient, *mockTokenizer, *mockApprovalStorage)
		request    core.ApprovalRequest
		wantErr    bool
		errMsg     string
	}{
		{
			name: "successful rejection with message ID",
			setupMocks: func(ctx context.Context, e *mockEmailClient, tok *mockTokenizer, s *mockApprovalStorage) {
				req := shcore.SubmitRequest{
					SubmissionEmail: "test@example.com",
					Vacancy: shcore.Vacancy{
						JobTitle: "Test Job",
					},
				}
				tok.On("DecodeToken", "test-token").Return(req, nil)
				s.On("GetApprovalState", ctx, "test-message").Return(core.ApprovalStatePending, nil)
				e.On("RejectRequest", ctx, "test-message").Return(nil)
				s.On("UpdateApprovalState", ctx, "test-message", core.ApprovalStateRejected).Return(nil)
			},
			request: core.ApprovalRequest{
				MessageID:    "test-message",
				TokenRequest: "test-token",
			},
		},
		{
			name: "successful rejection without message ID",
			setupMocks: func(ctx context.Context, e *mockEmailClient, tok *mockTokenizer, s *mockApprovalStorage) {
				req := shcore.SubmitRequest{
					SubmissionEmail: "test@example.com",
					Vacancy: shcore.Vacancy{
						JobTitle: "Test Job",
					},
				}
				tok.On("DecodeToken", "test-token").Return(req, nil)
			},
			request: core.ApprovalRequest{
				TokenRequest: "test-token",
			},
		},
		{
			name: "rejection already processed",
			setupMocks: func(ctx context.Context, e *mockEmailClient, tok *mockTokenizer, s *mockApprovalStorage) {
				req := shcore.SubmitRequest{
					SubmissionEmail: "test@example.com",
				}
				tok.On("DecodeToken", "test-token").Return(req, nil)
				s.On("GetApprovalState", ctx, "test-message").Return(core.ApprovalStateRejected, nil)
			},
			request: core.ApprovalRequest{
				MessageID:    "test-message",
				TokenRequest: "test-token",
			},
			wantErr: true,
			errMsg:  "approval already processed",
		},
		{
			name: "invalid token",
			setupMocks: func(ctx context.Context, e *mockEmailClient, tok *mockTokenizer, s *mockApprovalStorage) {
				tok.On("DecodeToken", "invalid-token").Return(shcore.SubmitRequest{}, shcore.NewBadRequestError("invalid token"))
			},
			request: core.ApprovalRequest{
				MessageID:    "test-message",
				TokenRequest: "invalid-token",
			},
			wantErr: true,
			errMsg:  "invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			queue := &mockQueue{}
			email := &mockEmailClient{}
			tokenizer := &mockTokenizer{}
			storage := &mockApprovalStorage{}

			svc, err := core.NewService(core.ServiceConfig{
				Queue:           queue,
				Email:           email,
				Tokenizer:       tokenizer,
				Approval:        &mockApproval{}, // not used
				ApprovalStorage: storage,
			})
			require.NoError(t, err)

			tt.setupMocks(ctx, email, tokenizer, storage)

			err = svc.HandleReject(ctx, tt.request)
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
			mock.AssertExpectationsForObjects(t, email, tokenizer, storage)
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

func (m *mockEmailClient) SendBulkApprovalRequest(ctx context.Context, req shcore.SubmitRequest, tokenReqVacancies []string) ([]string, error) {
	args := m.Called(ctx, req, tokenReqVacancies)
	return args.Get(0).([]string), args.Error(1)
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

func (m *mockApprovalStorage) GetApprovalState(ctx context.Context, messageID string) (core.ApprovalState, error) {
	args := m.Called(ctx, messageID)
	return args.Get(0).(core.ApprovalState), args.Error(1)
}

func (m *mockApprovalStorage) UpdateApprovalState(ctx context.Context, messageID string, state core.ApprovalState) error {
	args := m.Called(ctx, messageID, state)
	return args.Error(0)
}

func (m *mockApprovalStorage) SaveApprovalRequest(ctx context.Context, messageID string, req shcore.SubmitRequest) error {
	args := m.Called(ctx, messageID, req)
	return args.Error(0)
}

func (m *mockApprovalStorage) SaveBulkApprovalRequest(ctx context.Context, req shcore.SubmitRequest, messageIDs []string) error {
	args := m.Called(ctx, req, messageIDs)
	return args.Error(0)
}
