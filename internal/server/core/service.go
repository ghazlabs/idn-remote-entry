package core

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"gopkg.in/validator.v2"
)

type Service interface {
	HandleRequest(ctx context.Context, req core.SubmitRequest) error
	HandleApprove(ctx context.Context, approvalReq ApprovalRequest) error
	HandleReject(ctx context.Context, approvalReq ApprovalRequest) error
}

type ServiceConfig struct {
	Queue           Queue           `validate:"nonnil"`
	Email           EmailClient     `validate:"nonnil"`
	Tokenizer       Tokenizer       `validate:"nonnil"`
	Approval        Approval        `validate:"nonnil"`
	ApprovalStorage ApprovalStorage `validate:"nonnil"`
}

func NewService(cfg ServiceConfig) (Service, error) {
	err := validator.Validate(cfg)
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &service{
		ServiceConfig: cfg,
	}, nil
}

type service struct {
	ServiceConfig
}

func (s *service) HandleRequest(ctx context.Context, req core.SubmitRequest) error {
	err := req.Validate()
	if err != nil {
		return core.NewBadRequestError(fmt.Sprintf("invalid request: %v", err))
	}

	if req.SubmissionType == core.SubmitTypeBulk {
		return s.handleBulkRequest(ctx, req)
	}

	token, err := s.Tokenizer.EncodeRequest(req)
	if err != nil {
		return fmt.Errorf("failed to encode token: %w", err)
	}

	if s.Approval.NeedsApproval(req.SubmissionEmail) {
		messageID, err := s.Email.SendApprovalRequest(ctx, req, token)
		if err != nil {
			return fmt.Errorf("failed to send approval request: %w", err)
		}

		err = s.ApprovalStorage.SaveApprovalRequest(ctx, messageID, req)
		if err != nil {
			return fmt.Errorf("failed to save approval request: %w", err)
		}

		return nil
	}

	err = s.Queue.Put(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to put request in queue: %w", err)
	}

	return nil
}

func (s *service) HandleApprove(ctx context.Context, approvalReq ApprovalRequest) error {
	req, err := s.Tokenizer.DecodeToken(approvalReq.TokenRequest)
	if err != nil {
		return err
	}

	err = req.Validate()
	if err != nil {
		return core.NewBadRequestError(fmt.Sprintf("invalid request: %v", err))
	}

	if approvalReq.MessageID != "" {
		approvalState, err := s.ApprovalStorage.GetApprovalState(ctx, approvalReq.MessageID)
		if err != nil {
			return err
		}

		if approvalState != ApprovalStatePending {
			return core.NewBadRequestError("approval already processed")
		}

		err = s.ApprovalStorage.UpdateApprovalState(ctx, approvalReq.MessageID, ApprovalStateApproved)
		if err != nil {
			return err
		}

		// if bulk request, we dont need to send approval email
		if !strings.Contains(approvalReq.MessageID, "bulk") {
			err = s.Email.ApproveRequest(ctx, approvalReq.MessageID)
			if err != nil {
				return fmt.Errorf("failed to send approval request: %w", err)
			}
		}
	}

	err = s.Queue.Put(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to put request in queue: %w", err)
	}

	return nil
}

func (s *service) HandleReject(ctx context.Context, approvalReq ApprovalRequest) error {
	req, err := s.Tokenizer.DecodeToken(approvalReq.TokenRequest)
	if err != nil {
		return err
	}

	err = req.Validate()
	if err != nil {
		return core.NewBadRequestError(fmt.Sprintf("invalid request: %v", err))
	}

	if approvalReq.MessageID != "" {
		approvalState, err := s.ApprovalStorage.GetApprovalState(ctx, approvalReq.MessageID)
		if err != nil {
			return err
		}

		if approvalState != ApprovalStatePending {
			return core.NewBadRequestError("approval already processed")
		}

		err = s.ApprovalStorage.UpdateApprovalState(ctx, approvalReq.MessageID, ApprovalStateRejected)
		if err != nil {
			return err
		}

		// if bulk request, we dont need to send rejection email
		if !strings.Contains(approvalReq.MessageID, "bulk") {
			err = s.Email.RejectRequest(ctx, approvalReq.MessageID)
			if err != nil {
				return fmt.Errorf("failed to send approval request: %w", err)
			}
		}
	}

	return nil
}

func (s *service) handleBulkRequest(ctx context.Context, bulkReq core.SubmitRequest) error {
	tokenReqs := make([]string, 0)
	for _, v := range bulkReq.BulkVacancies {
		req := core.SubmitRequest{
			// for now we only support URL submission in bulk request
			SubmissionType:  core.SubmitTypeURL,
			SubmissionEmail: bulkReq.SubmissionEmail,
			Vacancy:         v,
		}

		token, err := s.Tokenizer.EncodeRequest(req)
		if err != nil {
			log.Printf("failed to encode token: %v", err)
			continue
		}
		tokenReqs = append(tokenReqs, token)
	}

	// For bulk request, we need to send approval request to admin
	messageIDs, err := s.Email.SendBulkApprovalRequest(ctx, bulkReq, tokenReqs)
	if err != nil {
		return fmt.Errorf("failed to send approval request: %w", err)
	}

	err = s.ApprovalStorage.SaveBulkApprovalRequest(ctx, bulkReq, messageIDs)
	if err != nil {
		return fmt.Errorf("failed to save approval request: %w", err)
	}

	return nil
}
