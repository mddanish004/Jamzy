package apperr

import "errors"

type Error struct {
	StatusCode int
	Message    string
}

func (e *Error) Error() string {
	return e.Message
}

func New(message string, statusCode int) *Error {
	return &Error{StatusCode: statusCode, Message: message}
}

func StatusCode(err error) int {
	var target *Error
	if errors.As(err, &target) {
		return target.StatusCode
	}
	return 0
}
