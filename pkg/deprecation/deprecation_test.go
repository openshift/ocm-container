package deprecation

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Pkg/Deprecation/Deprecation", func() {
	Context("Message()", func() {
		It("Returns deprecation message without alternative", func() {
			msg := Message("old-flag")
			Expect(msg).To(Equal("WARNING: 'old-flag' is deprecated and will be removed in a future release."))
		})

		It("Returns deprecation message with alternative", func() {
			msg := Message("old-flag", "new-flag")
			Expect(msg).To(ContainSubstring("WARNING: 'old-flag' is deprecated"))
			Expect(msg).To(ContainSubstring("Please use 'new-flag' instead"))
		})
	})

	Context("ShortMessage()", func() {
		It("Returns short message without alternative", func() {
			msg := ShortMessage("old-flag")
			Expect(msg).To(Equal("(deprecated)"))
		})

		It("Returns short message with alternative", func() {
			msg := ShortMessage("old-flag", "new-flag")
			Expect(msg).To(Equal("(deprecated: use 'new-flag' instead)"))
		})
	})
})
