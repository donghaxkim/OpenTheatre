package main

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("V1Room", func() {
	Describe("NewV1Room", func() {
		It("returns a room with the given id, name, and host", func() {
			room := NewV1Room("B3K7M9", "Alex's room", "m_alex")
			Expect(room.Id).To(Equal("B3K7M9"))
			Expect(room.Name).To(Equal("Alex's room"))
			Expect(room.HostId).To(Equal("m_alex"))
		})

		It("defaults ControlMode to democratic", func() {
			Expect(NewV1Room("B3K7M9", "n", "h").ControlMode).To(Equal("democratic"))
		})

		It("defaults UrlSyncMode to ask", func() {
			Expect(NewV1Room("B3K7M9", "n", "h").UrlSyncMode).To(Equal("ask"))
		})

		It("sets CreatedAt to a non-zero unix-ms timestamp", func() {
			Expect(NewV1Room("B3K7M9", "n", "h").CreatedAt).To(BeNumerically(">", 0))
		})
	})

	Describe("Rename", func() {
		It("updates Name when called by host", func() {
			room := NewV1Room("B3K7M9", "old", "m_alex")
			Expect(room.Rename("m_alex", "new name")).To(Succeed())
			Expect(room.Name).To(Equal("new name"))
		})

		It("returns an error when called by non-host", func() {
			room := NewV1Room("B3K7M9", "old", "m_alex")
			Expect(room.Rename("m_bob", "new name")).To(MatchError("not host"))
			Expect(room.Name).To(Equal("old"))
		})
	})

	Describe("UpdateSettings", func() {
		It("updates ControlMode when called by host", func() {
			room := NewV1Room("B3K7M9", "n", "m_alex")
			Expect(room.UpdateSettings("m_alex", V1SettingsPatch{ControlMode: "host-only"})).To(Succeed())
			Expect(room.ControlMode).To(Equal("host-only"))
		})

		It("updates UrlSyncMode when called by host", func() {
			room := NewV1Room("B3K7M9", "n", "m_alex")
			Expect(room.UpdateSettings("m_alex", V1SettingsPatch{UrlSyncMode: "auto"})).To(Succeed())
			Expect(room.UrlSyncMode).To(Equal("auto"))
		})

		It("rejects unknown ControlMode values", func() {
			room := NewV1Room("B3K7M9", "n", "m_alex")
			Expect(room.UpdateSettings("m_alex", V1SettingsPatch{ControlMode: "anarchy"})).To(MatchError("invalid ControlMode"))
		})

		It("rejects unknown UrlSyncMode values", func() {
			room := NewV1Room("B3K7M9", "n", "m_alex")
			Expect(room.UpdateSettings("m_alex", V1SettingsPatch{UrlSyncMode: "teleport"})).To(MatchError("invalid UrlSyncMode"))
		})

		It("returns an error when called by non-host", func() {
			room := NewV1Room("B3K7M9", "n", "m_alex")
			Expect(room.UpdateSettings("m_bob", V1SettingsPatch{ControlMode: "host-only"})).To(MatchError("not host"))
		})
	})
})
