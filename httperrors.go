package httperrors

import (
	"fmt"
	"runtime"
	"strings"
)

// HTTPError is a replacement for the standard go error to be used in HTTP servers. It contains
// extra information about the error like a response code and a stack trace.
type HTTPError struct {
	msg   string
	code  int
	stack string
	inner error
}

const (
	// UninitializedResponseCode indicates that the response code for this HTTPError was never set
	UninitializedResponseCode = -1
)

func (e *HTTPError) Error() string {
	errorMsgs := []string{
		"ERROR:\t* " + e.msg,
	}

	inner := e.inner
	for inner != nil {
		httpError, ok := inner.(*HTTPError)
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

// Message gets the outermost message
func (e *HTTPError) Message() string {
	return e.msg
}

// InnerMessage gets the innermost message
func (e *HTTPError) InnerMessage() string {
	var ok bool
	var msgErr, nextMsgErr *HTTPError
	msgErr = e
	for msgErr.inner != nil {
		nextMsgErr, ok = msgErr.inner.(*HTTPError)
		if !ok {
			return msgErr.inner.Error()
		}
		msgErr = nextMsgErr
	}
	return msgErr.msg
}

// SetResponseCode sets the response code of this HTTPError
func (e *HTTPError) SetResponseCode(code int) {
	e.code = code
}

// ResponseCode gets the outermost response code
func (e *HTTPError) ResponseCode() int {
	var ok bool
	var codeErr, nextCodeErr *HTTPError
	codeErr = e
	for codeErr.code == UninitializedResponseCode {
		nextCodeErr, ok = codeErr.inner.(*HTTPError)
		if !ok {
			return codeErr.code
		}
		codeErr = nextCodeErr
	}
	return codeErr.code
}

// StackTrace gets the innermost available stacktrace
func (e *HTTPError) StackTrace() string {
	var ok bool
	var stackErr, nextStackErr *HTTPError
	stackErr = e
	for stackErr.inner != nil {
		nextStackErr, ok = stackErr.inner.(*HTTPError)
		if !ok {
			return stackErr.stack
		}
		stackErr = nextStackErr
	}
	return stackErr.stack
}

// Wrap takes an existing error and turns it into a *HTTPError
func Wrap(err error, msg string) *HTTPError {
	resp := HTTPError{
		msg:   msg,
		code:  UninitializedResponseCode,
		inner: err,
	}

	// Wrap will only get a new stack trace if one does not exist
	_, ok := err.(*HTTPError)
	if !ok {
		resp.stack = stackTrace()
	}
	return &resp
}

// Wrapf wraps an existing error with printf paramaters
func Wrapf(err error, format string, args ...interface{}) *HTTPError {
	return Wrap(err, fmt.Sprintf(format, args...))
}

// New creates a new *HTTPError
func New(msg string) *HTTPError {
	return Wrap(nil, msg)
}

// Newf creates a new *HTTPError with printf parameters
func Newf(format string, args ...interface{}) *HTTPError {
	return Wrap(nil, fmt.Sprintf(format, args...))
}

// stackTrace returns the current stack trace
func stackTrace() string {
	buf := make([]byte, 1024)
	bytesRead := 0
	for {
		bytesRead = runtime.Stack(buf, false)
		if bytesRead < len(buf) {
			break
		}
		buf = make([]byte, len(buf)*2)
	}
	return string(buf[:bytesRead])
}
