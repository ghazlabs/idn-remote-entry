package core

import (
	"context"
	"fmt"

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

	token, err := s.Tokenizer.EncodeRequest(req)
	if err != nil {
		return fmt.Errorf("failed to encode token: %w", err)
	}

	if s.Approval.NeedsApproval(req.SubmissionEmail) {
		messageID, err := s.Email.SendApprovalRequest(ctx, req, token)
		if err != nil {
			return fmt.Errorf("failed to send approval request: %w", err)
		}

		err = s.ApprovalStorage.SaveApprovalRequest(messageID, req)
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
		approvalState, err := s.ApprovalStorage.GetApprovalState(approvalReq.MessageID)
		if err != nil {
			return err
		}

		if approvalState != ApprovalStatePending {
			return core.NewBadRequestError("approval already processed")
		}

		err = s.Email.ApproveRequest(ctx, approvalReq.MessageID)
		if err != nil {
			return fmt.Errorf("failed to send approval request: %w", err)
		}

		err = s.ApprovalStorage.UpdateApprovalState(approvalReq.MessageID, ApprovalStateApproved)
		if err != nil {
			return err
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
		approvalState, err := s.ApprovalStorage.GetApprovalState(approvalReq.MessageID)
		if err != nil {
			return err
		}

		if approvalState != ApprovalStatePending {
			return core.NewBadRequestError("approval already processed")
		}

		err = s.Email.RejectRequest(ctx, approvalReq.MessageID)
		if err != nil {
			return fmt.Errorf("failed to send approval request: %w", err)
		}

		err = s.ApprovalStorage.UpdateApprovalState(approvalReq.MessageID, ApprovalStateRejected)
		if err != nil {
			return err
		}
	}

	err = s.Queue.Put(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to put request in queue: %w", err)
	}

	return nil
}
