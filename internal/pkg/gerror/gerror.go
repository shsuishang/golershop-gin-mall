package gerror

import (
	"errors"
	"fmt"
	"net/http"
)

type Error struct {
	HTTPStatus int
	Code       int
	Message    string
	Err        error
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func New(message string) error {
	return &Error{
		HTTPStatus: http.StatusOK,
		Code:       0,
		Message:    message,
	}
}

func Newf(format string, args ...interface{}) error {
	return New(fmt.Sprintf(format, args...))
}

func NewCode(code int, message string) error {
	return &Error{
		HTTPStatus: http.StatusOK,
		Code:       code,
		Message:    message,
	}
}

func NewCodef(code int, format string, args ...interface{}) error {
	return NewCode(code, fmt.Sprintf(format, args...))
}

func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return &Error{
		HTTPStatus: http.StatusOK,
		Code:       Code(err),
		Message:    message + ": " + err.Error(),
		Err:        err,
	}
}

func Wrapf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return Wrap(err, fmt.Sprintf(format, args...))
}

func WrapCode(code int, err error, message string) error {
	if err == nil {
		return nil
	}
	return &Error{
		HTTPStatus: http.StatusOK,
		Code:       code,
		Message:    message + ": " + err.Error(),
		Err:        err,
	}
}

func WrapCodef(code int, err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	return WrapCode(code, err, fmt.Sprintf(format, args...))
}

func Code(err error) int {
	if err == nil {
		return 0
	}
	var ge *Error
	if errors.As(err, &ge) {
		return ge.Code
	}
	return 0
}

func Message(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
