package httperrors

import (
	"fmt"
	"runtime"
	"strings"
)

// HTTPError is a replacement for the standard go error to be used in HTTP servers. It contains
// extra information about the error like a response code and a stack trace.
type HTTPError interface {
	// Error implements error interface. Responds with all messages wrapped in stack of HTTPErrors.
	Error() string

	// Message gets the outermost message
	Message() string

	// InnerMessage gets the innermost message
	InnerMessage() string

	// SetResponseCode sets the response code of this HTTPError
	SetResponseCode(respCode int) HTTPError

	// ResponseCode gets the outermost response code
	ResponseCode() int

	// SetErrorCode sets the error code of this HTTPError to the provided string
	SetErrorCode(errCode string) HTTPError

	// ErrorCode gets the outermost error code
	ErrorCode() string

	// StackTrace gets the innermost available stacktrace
	StackTrace() string
}

type baseHTTPError struct {
	msg      string
	respCode int
	errCode  string
	stack    string
	inner    error
}

const (
	// UninitializedResponseCode indicates that the response code for this HTTPError was never set
	UninitializedResponseCode = -1

	// UninitializedErrorCode indicates that the error code for this HTTPError was never set
	UninitializedErrorCode = ""

	// UnknownErrorMsg is returned when an error message is not specified
	UnknownErrorMsg = "Unknown error"

	// UninitializedStackTrace is returned when a standard error is cast to an HTTPError
	UninitializedStackTrace = "Stack trace unavailable"
)

func (e *baseHTTPError) Error() string {
	var errorMsgs []string
	var inner error

	inner = e
	for inner != nil {
		httpError, ok := inner.(*baseHTTPError)
		if ok {
			if httpError.msg != "" {
				errorMsgs = append(errorMsgs, httpError.msg)
			}
			inner = httpError.inner
		} else {
			errorMsgs = append(errorMsgs, inner.Error())
			inner = nil
		}
	}

	if len(errorMsgs) == 0 {
		errorMsgs = append(errorMsgs, UnknownErrorMsg)
	}
	return strings.Join(errorMsgs, "\n")
}

func (e *baseHTTPError) Message() string {
	var ok bool
	var validErr, nextValidErr *baseHTTPError

	validErr = e
	for validErr.msg == "" {
		if validErr.inner == nil {
			return UnknownErrorMsg
		}
		nextValidErr, ok = validErr.inner.(*baseHTTPError)
		if !ok {
			return validErr.inner.Error()
		}
		validErr = nextValidErr
	}
	return validErr.msg
}

func (e *baseHTTPError) InnerMessage() string {
	var ok bool
	var msgErr, nextMsgErr *baseHTTPError
	msgErr = e
	for msgErr.inner != nil {
		nextMsgErr, ok = msgErr.inner.(*baseHTTPError)
		if !ok {
			return msgErr.inner.Error()
		}
		msgErr = nextMsgErr
	}

	if msgErr.msg == "" {
		return UnknownErrorMsg
	}
	return msgErr.msg
}

func (e *baseHTTPError) SetResponseCode(respCode int) HTTPError {
	e.respCode = respCode
	return e
}

func (e *baseHTTPError) ResponseCode() int {
	var ok bool
	var codeErr, nextCodeErr *baseHTTPError
	codeErr = e
	for codeErr.respCode == UninitializedResponseCode {
		nextCodeErr, ok = codeErr.inner.(*baseHTTPError)
		if !ok {
			return codeErr.respCode
		}
		codeErr = nextCodeErr
	}
	return codeErr.respCode
}

func (e *baseHTTPError) SetErrorCode(errCode string) HTTPError {
	e.errCode = errCode
	return e
}

func (e *baseHTTPError) ErrorCode() string {
	var ok bool
	var codeErr, nextCodeErr *baseHTTPError
	codeErr = e
	for codeErr.errCode == UninitializedErrorCode {
		nextCodeErr, ok = codeErr.inner.(*baseHTTPError)
		if !ok {
			return codeErr.errCode
		}
		codeErr = nextCodeErr
	}
	return codeErr.errCode
}

func (e *baseHTTPError) StackTrace() string {
	var ok bool
	var stackErr, nextStackErr *baseHTTPError
	stackErr = e
	for stackErr.inner != nil {
		nextStackErr, ok = stackErr.inner.(*baseHTTPError)
		if !ok {
			return stackErr.stack
		}
		stackErr = nextStackErr
	}
	return stackErr.stack
}

// ToHTTPError detects if the error is an HTTPError and returns it or
// creates an HTTPError from a standard error
func ToHTTPError(err error) HTTPError {
	httpErr, ok := err.(HTTPError)
	if !ok {
		return &baseHTTPError{
			msg:      "",
			respCode: UninitializedResponseCode,
			inner:    err,
			stack:    UninitializedStackTrace,
		}
	}
	return httpErr
}

// Wrap takes an existing error and turns it into a HTTPError
func Wrap(err error, msg string) HTTPError {
	resp := baseHTTPError{
		msg:      msg,
		respCode: UninitializedResponseCode,
		inner:    err,
	}

	// Wrap will only get a new stack trace if one does not exist
	_, ok := err.(*baseHTTPError)
	if !ok {
		resp.stack = stackTrace()
	}
	return &resp
}

// Wrapf wraps an existing error with printf paramaters
func Wrapf(err error, format string, args ...interface{}) HTTPError {
	resp := baseHTTPError{
		msg:      fmt.Sprintf(format, args...),
		respCode: UninitializedResponseCode,
		inner:    err,
	}

	// Wrap will only get a new stack trace if one does not exist
	_, ok := err.(*baseHTTPError)
	if !ok {
		resp.stack = stackTrace()
	}
	return &resp
}

// New creates a new HTTPError
func New(msg string) HTTPError {
	return &baseHTTPError{
		msg:      msg,
		respCode: UninitializedResponseCode,
		stack:    stackTrace(),
	}
}

// Newf creates a new HTTPError with printf parameters
func Newf(format string, args ...interface{}) HTTPError {
	return &baseHTTPError{
		msg:      fmt.Sprintf(format, args...),
		respCode: UninitializedResponseCode,
		stack:    stackTrace(),
	}
}

// stackTrace returns the current stack trace
func stackTrace() string {
	buf := make([]byte, 2048)
	bytesRead := 0
	for {
		bytesRead = runtime.Stack(buf, false)
		if bytesRead < len(buf) {
			break
		}
		buf = make([]byte, len(buf)*2)
	}

	// split stack trace to remove lines inside httperrors
	lines := strings.Split(string(buf[:bytesRead]), "\n")
	trimmedLines := append(lines[:1], lines[5:]...)
	return strings.Join(trimmedLines, "\n")
}
