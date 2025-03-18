package core

import (
	"context"

	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
)

type Queue interface {
	Put(ctx context.Context, req core.SubmitRequest) error
}

type EmailClient interface {
	SendApprovalRequest(ctx context.Context, req core.SubmitRequest, tokenReq string) error
}

type Tokenizer interface {
	EncodeRequest(req core.SubmitRequest) (string, error)
	DecodeToken(tokenStr string) (core.SubmitRequest, error)
}

type Approval interface {
	NeedsApproval(submitterEmail string) bool
}
