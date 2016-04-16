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
	SetResponseCode(code int)

	// ResponseCode gets the outermost response code
	ResponseCode() int

	// StackTrace gets the innermost available stacktrace
	StackTrace() string
}

type baseHTTPError struct {
	msg   string
	code  int
	stack string
	inner error
}

const (
	// UninitializedResponseCode indicates that the response code for this HTTPError was never set
	UninitializedResponseCode = -1
)

func (e *baseHTTPError) Error() string {
	errorMsgs := []string{
		"ERROR:\t* " + e.msg,
	}

	inner := e.inner
	for inner != nil {
		httpError, ok := inner.(*baseHTTPError)
		if ok {
			errorMsgs = append(errorMsgs, "\t* "+httpError.msg)
			inner = httpError.inner
		} else {
			errorMsgs = append(errorMsgs, "\t* "+inner.Error())
			inner = nil
		}
	}
	return strings.Join(errorMsgs, "\n")
}

func (e *baseHTTPError) Message() string {
	return e.msg
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
	return msgErr.msg
}

func (e *baseHTTPError) SetResponseCode(code int) {
	e.code = code
}

func (e *baseHTTPError) ResponseCode() int {
	var ok bool
	var codeErr, nextCodeErr *baseHTTPError
	codeErr = e
	for codeErr.code == UninitializedResponseCode {
		nextCodeErr, ok = codeErr.inner.(*baseHTTPError)
		if !ok {
			return codeErr.code
		}
		codeErr = nextCodeErr
	}
	return codeErr.code
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

// Wrap takes an existing error and turns it into a HTTPError
func Wrap(err error, msg string) HTTPError {
	resp := baseHTTPError{
		msg:   msg,
		code:  UninitializedResponseCode,
		inner: err,
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
		msg:   fmt.Sprintf(format, args...),
		code:  UninitializedResponseCode,
		inner: err,
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
		msg:   msg,
		code:  UninitializedResponseCode,
		stack: stackTrace(),
	}
}

// Newf creates a new HTTPError with printf parameters
func Newf(format string, args ...interface{}) HTTPError {
	return &baseHTTPError{
		msg:   fmt.Sprintf(format, args...),
		code:  UninitializedResponseCode,
		stack: stackTrace(),
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
