package test

import "errors"

var (
	ErrTestNotFound        = errors.New("test not found")
	ErrTestConflict        = errors.New("test already exists")
	ErrQuestionNotFound    = errors.New("question not found")
	ErrInvalidQuestionData = errors.New("question must contain exactly one correct option")
	ErrResultNotFound      = errors.New("result not found")
	ErrResultAccessDenied  = errors.New("result access denied")
)
