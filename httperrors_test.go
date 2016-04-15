package httperrors

import (
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("HTTPError", func() {
	var (
		err                   error
		outerHTTPErr, httpErr *HTTPError
	)

	Describe("Creation", func() {

		Describe("Wrap", func() {
			BeforeEach(func() {
				err = fmt.Errorf("test err")
				httpErr = Wrap(err, "http test err")
				outerHTTPErr = Wrap(httpErr, "outer http test err")
			})

			It("Wraps an error", func() {
				Expect(httpErr.inner).To(Equal(err))
			})
			It("Wraps another HTTPError", func() {
				Expect(outerHTTPErr.msg).To(Equal("outer http test err"))
				Expect(outerHTTPErr.inner).To(Equal(httpErr))

				tempErr, ok := outerHTTPErr.inner.(*HTTPError)
				Expect(ok).To(BeTrue())
				Expect(tempErr.inner).To(Equal(err))
			})
			It("Only the innermost HTTPError has a stack trace", func() {
				Expect(outerHTTPErr.stack).To(BeEmpty())

				tempErr, ok := outerHTTPErr.inner.(*HTTPError)
				Expect(ok).To(BeTrue())
				Expect(tempErr.stack).ToNot(BeEmpty())
			})
		})

		Describe("Wrapf", func() {
			It("Wraps an error with printf parameters", func() {
				err = fmt.Errorf("test error")
				httpErr = Wrapf(err, "%s test err", "http")

				Expect(httpErr.msg).To(Equal("http test err"))
				Expect(httpErr.inner).To(Equal(err))
			})
		})

		Describe("New", func() {
			It("Creates a new HTTPError", func() {
				httpErr = New("http test err")
				Expect(httpErr.msg).To(Equal("http test err"))
				Expect(httpErr.inner).To(BeNil())
			})
		})

		Describe("Newf", func() {
			It("Creates a new HTTPError with printf parameters", func() {
				httpErr = Newf("%s test err", "http")
				Expect(httpErr.msg).To(Equal("http test err"))
				Expect(httpErr.inner).To(BeNil())
			})
		})
	})

	Describe("Information", func() {
		BeforeEach(func() {
			err = fmt.Errorf("test err")
			httpErr = Wrap(err, "http test err")
			outerHTTPErr = Wrap(httpErr, "outer http test err")
		})

		Describe("Error", func() {
			It("Prints out wrapped errors in succession", func() {
				Expect(err.Error()).To(Equal("test err"))
				Expect(httpErr.Error()).To(Equal("ERROR:\t* http test err\n\t* test err"))
				Expect(outerHTTPErr.Error()).To(Equal("ERROR:\t* outer http test err\n\t* http test err\n\t* test err"))
			})
		})

		Describe("Message", func() {
			It("Gets the outermost message", func() {
				Expect(httpErr.Message()).To(Equal("http test err"))
				Expect(outerHTTPErr.Message()).To(Equal("outer http test err"))
			})
		})

		Describe("InnerMessage", func() {
			It("Gets the innermost message", func() {
				Expect(httpErr.InnerMessage()).To(Equal("test err"))
				Expect(outerHTTPErr.InnerMessage()).To(Equal("test err"))
			})
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
				Expect(outerHTTPErr.StackTrace()).To(Equal(httpErr.stack))
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

	Describe("Root error is HTTPError", func() {
		BeforeEach(func() {
			httpErr = New("http test err")
			outerHTTPErr = Wrap(httpErr, "outer http test err")
		})

		It("Gets the innermost stack trace", func() {
			Expect(outerHTTPErr.StackTrace()).To(Equal(httpErr.stack))
		})

		It("Gets the innermost message", func() {
			Expect(outerHTTPErr.InnerMessage()).To(Equal("http test err"))
		})
	})
})
