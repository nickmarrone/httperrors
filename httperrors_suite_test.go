package httperrors

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestHttperrs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Httperrs Suite")
}
