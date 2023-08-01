package specs_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSpecs(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Sanity Test")
}

var _ = Describe("Sanity", func() {
	Describe("Testing framework", func() {
		Context("simple scenario", func() {
			It("should pass", func() {
				Expect(1).To(Equal(1))
			})
		})
	})
})
