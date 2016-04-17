package httperrors

import (
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HTTPError", func() {
	var (
		err          error
		outerHTTPErr HTTPError
		httpErr      HTTPError
		emptyErr     HTTPError
		castErr      HTTPError
	)

	Describe("Creating", func() {

		Describe("a wrapped error", func() {
			BeforeEach(func() {
				err = fmt.Errorf("test err")
				httpErr = Wrap(err, "http test err")
				outerHTTPErr = Wrap(httpErr, "outer http test err")
			})

			It("gets the innermost message", func() {
				Expect(httpErr.InnerMessage()).To(Equal(err.Error()))
				Expect(outerHTTPErr.InnerMessage()).To(Equal(err.Error()))
			})
			It("gets the outermost message", func() {
				Expect(httpErr.Message()).To(Equal("http test err"))
				Expect(outerHTTPErr.Message()).To(Equal("outer http test err"))
			})
			It("prints out the stack of error messages", func() {
				Expect(httpErr.Error()).To(Equal("http test err\ntest err"))
				Expect(outerHTTPErr.Error()).To(Equal("outer http test err\nhttp test err\ntest err"))
			})
			It("at any level has the same stack trace", func() {
				Expect(outerHTTPErr.StackTrace()).To(Equal(httpErr.StackTrace()))
			})
		})

		Describe("a wrapped error with printf parameters", func() {
			BeforeEach(func() {
				err = fmt.Errorf("test err")
				httpErr = Wrapf(err, "%s test err", "http")
				outerHTTPErr = Wrapf(httpErr, "%s test err", "outer http")
			})

			It("gets the innermost message", func() {
				Expect(httpErr.InnerMessage()).To(Equal(err.Error()))
				Expect(outerHTTPErr.InnerMessage()).To(Equal(err.Error()))
			})
			It("gets the outermost message", func() {
				Expect(httpErr.Message()).To(Equal("http test err"))
				Expect(outerHTTPErr.Message()).To(Equal("outer http test err"))
			})
			It("prints out the stack of error messages", func() {
				Expect(httpErr.Error()).To(Equal("http test err\ntest err"))
				Expect(outerHTTPErr.Error()).To(Equal("outer http test err\nhttp test err\ntest err"))
			})
			It("at any level has the same stack trace", func() {
				Expect(outerHTTPErr.StackTrace()).To(Equal(httpErr.StackTrace()))
			})
		})

		Describe("a new error", func() {
			BeforeEach(func() {
				httpErr = New("test err")
			})

			It("gets the innermost message", func() {
				Expect(httpErr.InnerMessage()).To(Equal("test err"))
			})
			It("gets the outermost message", func() {
				Expect(httpErr.Message()).To(Equal("test err"))
			})
			It("prints out the stack of error messages", func() {
				Expect(httpErr.Error()).To(Equal("test err"))
			})
			It("has a stack trace", func() {
				Expect(outerHTTPErr.StackTrace()).ToNot(Equal(UninitializedStackTrace))
			})
		})

		Describe("a new error with printf parameters", func() {
			BeforeEach(func() {
				httpErr = Newf("%s err", "test")
			})

			It("gets the innermost message", func() {
				Expect(httpErr.InnerMessage()).To(Equal("test err"))
			})
			It("gets the outermost message", func() {
				Expect(httpErr.Message()).To(Equal("test err"))
			})
			It("prints out the stack of error messages", func() {
				Expect(httpErr.Error()).To(Equal("test err"))
			})
			It("has a stack trace", func() {
				Expect(outerHTTPErr.StackTrace()).ToNot(Equal(UninitializedStackTrace))
			})
		})

		Describe("an empty error", func() {
			BeforeEach(func() {
				emptyErr = New("")
			})

			It("has an unknown innermost message", func() {
				Expect(emptyErr.InnerMessage()).To(Equal(UnknownErrorMsg))
			})
			It("has an unknown outermost message", func() {
				Expect(emptyErr.Message()).To(Equal(UnknownErrorMsg))
			})
			It("prints out the stack of error messages", func() {
				Expect(emptyErr.Error()).To(Equal(UnknownErrorMsg))
			})
			It("has a stack trace", func() {
				Expect(emptyErr.StackTrace()).ToNot(Equal(UninitializedStackTrace))
			})
		})

		Describe("a cast error", func() {
			BeforeEach(func() {
				err = fmt.Errorf("test err")
				castErr = ToHTTPError(err)
			})

			It("gets the innermost message", func() {
				Expect(httpErr.InnerMessage()).To(Equal("test err"))
			})
			It("gets the outermost message", func() {
				Expect(httpErr.Message()).To(Equal("test err"))
			})
			It("prints out the stack of error messages", func() {
				Expect(httpErr.Error()).To(Equal("test err"))
			})
			It("has a stack trace", func() {
				Expect(outerHTTPErr.StackTrace()).ToNot(Equal(UninitializedStackTrace))
			})
		})

		Describe("a cast HTTPError", func() {
			It("gets the message", func() {
				httpErr = New("test err")
				castErr = ToHTTPError(httpErr)
				Expect(castErr.Message()).To(Equal("test err"))
			})
		})

		Describe("multiple layers of empty wrapped errors", func() {
			It("gets the message", func() {
				err = fmt.Errorf("test err")
				httpErr = Wrap(err, "")
				outerHTTPErr = Wrap(httpErr, "")
				Expect(outerHTTPErr.Message()).To(Equal("test err"))
			})
		})
	})

	Describe("Information", func() {
		BeforeEach(func() {
			err = fmt.Errorf("test err")
			httpErr = Wrap(err, "http test err")
			outerHTTPErr = Wrap(httpErr, "outer http test err")
		})

		Describe("ResponseCode", func() {
			It("Responds with UninitializedResponseCode if it has not been set", func() {
				Expect(outerHTTPErr.ResponseCode()).To(Equal(UninitializedResponseCode))
			})
			It("Responds with the outermost response code when set", func() {
				httpErr.SetResponseCode(http.StatusInternalServerError)
				Expect(outerHTTPErr.ResponseCode()).To(Equal(http.StatusInternalServerError))

				outerHTTPErr.SetResponseCode(http.StatusUnauthorized)
				Expect(outerHTTPErr.ResponseCode()).To(Equal(http.StatusUnauthorized))
			})
		})

		Describe("StackTrace", func() {
			It("Gets the stack trace from the innermost HTTPError", func() {
				Expect(outerHTTPErr.StackTrace()).To(Equal(httpErr.StackTrace()))
			})

			Describe("strips out parts of the stack trace relating to httperrors library", func() {
				It("for a New error", func() {
					httpErr := New("test err")
					Expect(httpErr.StackTrace()).ToNot(ContainSubstring("httperrors/httperrors.go"))
				})
				It("for a Newf error", func() {
					httpErr := Newf("test err %d", 2)
					Expect(httpErr.StackTrace()).ToNot(ContainSubstring("httperrors/httperrors.go"))
				})
				It("for a Wrap error", func() {
					err := fmt.Errorf("base test err")
					httpErr := Wrap(err, "test err")
					Expect(httpErr.StackTrace()).ToNot(ContainSubstring("httperrors/httperrors.go"))
				})
				It("for a Wrapf error", func() {
					err := fmt.Errorf("base test err")
					httpErr := Wrapf(err, "test err %d", 2)
					Expect(httpErr.StackTrace()).ToNot(ContainSubstring("httperrors/httperrors.go"))
				})
			})
		})
	})

	Describe("a wrapped error where the root error is HTTPError", func() {
		BeforeEach(func() {
			httpErr = New("http test err")
			outerHTTPErr = Wrap(httpErr, "outer http test err")
		})

		It("Gets the innermost stack trace", func() {
			Expect(outerHTTPErr.StackTrace()).To(Equal(httpErr.StackTrace()))
		})

		It("Gets the innermost message", func() {
			Expect(outerHTTPErr.InnerMessage()).To(Equal("http test err"))
		})
	})
})
