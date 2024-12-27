package myerror

import (
	"fmt"
)

type ErrorCode int

const (
	CodeValidtaionFailed ErrorCode = 1000 + iota
	CodeContextUserNotFound
	CodeHashPasswordFailed
	CodeCreateSessionFailed
	CodeNoLogin
)

const (
	CodeUserAlreadyExists ErrorCode = 2000 + iota
	CodeInvalidPassword
)

const (
	CodeQueryFailed ErrorCode = 3000 + iota
	CodeTaskNotFound
	CodeUserNotFound
	CodeGrantPermissionFailed
	CodePermissionNotFound
	CodePermissionDenied
	CodeTransactionNotFound
)

const (
	CodeUnExpected ErrorCode = 9999
)

var ErrMessages = map[ErrorCode]string{
	// 1000
	CodeValidtaionFailed:    "validation failed",
	CodeContextUserNotFound: "user not found in context",
	CodeHashPasswordFailed:  "failed to hash password",
	CodeCreateSessionFailed: "failed to create session",
	CodeNoLogin:             "user not logged in",

	// 2000
	CodeUserAlreadyExists: "user already exists",
	CodeInvalidPassword:   "invalid password",

	// 3000
	CodeQueryFailed:           "failed to execute query",
	CodeTaskNotFound:          "task not found",
	CodeUserNotFound:          "user not found",
	CodeGrantPermissionFailed: "failed to grant permission",
	CodePermissionNotFound:    "permission not found",
	CodePermissionDenied:      "permission denied",
	CodeTransactionNotFound:   "failed to get transaction from context",

	// 9999
	CodeUnExpected: "unexpected error occurred",
}

var (
	// 1000
	ErrValidation          = &AppError{Code: CodeValidtaionFailed, Message: ErrMessages[CodeValidtaionFailed]}
	ErrContextUserNotFound = &AppError{Code: CodeContextUserNotFound, Message: ErrMessages[CodeContextUserNotFound]}
	ErrHashPassword        = &AppError{Code: CodeHashPasswordFailed, Message: ErrMessages[CodeHashPasswordFailed]}
	ErrCreateSession       = &AppError{Code: CodeCreateSessionFailed, Message: ErrMessages[CodeCreateSessionFailed]}
	ErrNoLogin             = &AppError{Code: CodeNoLogin, Message: ErrMessages[CodeNoLogin]}

	// 2000
	ErrUserAlreadyExists = &AppError{Code: CodeUserAlreadyExists, Message: ErrMessages[CodeUserAlreadyExists]}
	ErrInvalidPassword   = &AppError{Code: CodeInvalidPassword, Message: ErrMessages[CodeInvalidPassword]}

	// 3000
	ErrQueryFailed         = &AppError{Code: CodeQueryFailed, Message: ErrMessages[CodeQueryFailed]}
	ErrTaskNotFound        = &AppError{Code: CodeTaskNotFound, Message: ErrMessages[CodeTaskNotFound]}
	ErrUserNotFound        = &AppError{Code: CodeUserNotFound, Message: ErrMessages[CodeUserNotFound]}
	ErrTransactionNotFound = &AppError{Code: CodeTransactionNotFound, Message: ErrMessages[CodeTransactionNotFound]}
	ErrGrantPermission     = &AppError{Code: CodeGrantPermissionFailed, Message: ErrMessages[CodeGrantPermissionFailed]}
	ErrPermissionNotFound  = &AppError{Code: CodePermissionNotFound, Message: ErrMessages[CodePermissionNotFound]}
	ErrPermissionDenied    = &AppError{Code: CodePermissionDenied, Message: ErrMessages[CodePermissionDenied]}

	// 9999
	ErrUnExpected = &AppError{Code: CodeUnExpected, Message: ErrMessages[CodeUnExpected]}
)

type AppError struct {
	Code        ErrorCode `json:"code"`
	Message     string    `json:"message"`
	Description string    `json:"description,omitempty"`
	err         error
}

func (e *AppError) Error() string {
	return fmt.Sprintf("code: %d, message: %s, description: %s, cause: %v", e.Code, e.Message, e.Description, e.err)
}

func (e *AppError) WithDescription(description string) *AppError {
	return &AppError{
		Code:        e.Code,
		Message:     e.Message,
		Description: description,
		err:         e.err,
	}
}

func (e *AppError) Unwrap() error {
	return e.err
}

func (e *AppError) Wrap(err error) *AppError {
	e.err = err
	return e
}

func (e *AppError) WrapWithDescription(err error, description string) *AppError {
	return &AppError{
		Code:        e.Code,
		Message:     e.Message,
		Description: description,
		err:         err,
	}
}

func (e *AppError) Is(target error) bool {
	// fmt.Println("called")
	t, ok := target.(*AppError)
	if !ok {
		return false
	}
	// fmt.Println(e.Code, t.Code)
	return e.Code == t.Code
}
