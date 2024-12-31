package core

import "fmt"

type Error struct {
	ErrCode string `json:"err"`
	Message string `json:"msg"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%v - %v", e.ErrCode, e.Message)
}

func NewBadRequestError(msg string) *Error {
	return &Error{
		ErrCode: "ERR_BAD_REQUEST",
		Message: msg,
	}
}

func NewInternalError(err error) *Error {
	return &Error{
		ErrCode: "ERR_INTERNAL_ERROR",
		Message: err.Error(),
	}
}
