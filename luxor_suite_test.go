package luxor_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestLuxor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Luxor Suite")
}
