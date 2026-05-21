package main

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("V1RoomCode", func() {
	Describe("GenerateV1RoomCode", func() {
		It("returns a 6-character string", func() {
			Expect(GenerateV1RoomCode()).To(HaveLen(6))
		})

		It("only uses characters from the allowed alphabet", func() {
			const allowed = "23456789ABCDEFGHJKMNPQRSTVWXYZ"
			for i := 0; i < 100; i++ {
				for _, ch := range GenerateV1RoomCode() {
					Expect(strings.ContainsRune(allowed, ch)).To(BeTrue(), "char %q not in alphabet", ch)
				}
			}
		})

		It("never produces 0, O, 1, I, L, or U", func() {
			const forbidden = "0O1ILU"
			for i := 0; i < 100; i++ {
				code := GenerateV1RoomCode()
				for _, ch := range forbidden {
					Expect(code).ToNot(ContainSubstring(string(ch)))
				}
			}
		})

		It("produces highly variable output over 1000 calls", func() {
			seen := make(map[string]bool)
			for i := 0; i < 1000; i++ {
				seen[GenerateV1RoomCode()] = true
			}
			Expect(len(seen)).To(BeNumerically(">=", 995))
		})
	})
})
