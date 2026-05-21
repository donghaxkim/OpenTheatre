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

	Describe("AddMember / RemoveMember", func() {
		It("AddMember stores the member by id", func() {
			room := NewV1Room("B3K7M9", "n", "m_alex")
			room.AddMember(V1Member{Id: "m_alex", DisplayName: "Alex", AvatarColor: "#ff6b9d"})
			m, ok := room.GetMember("m_alex")
			Expect(ok).To(BeTrue())
			Expect(m.DisplayName).To(Equal("Alex"))
		})

		It("AddMember twice with the same id increments connection count, not member count", func() {
			room := NewV1Room("B3K7M9", "n", "m_alex")
			room.AddMember(V1Member{Id: "m_alex", DisplayName: "Alex"})
			room.AddMember(V1Member{Id: "m_alex", DisplayName: "Alex"})
			Expect(room.MemberCount()).To(Equal(1))
			Expect(room.ConnectionCount("m_alex")).To(Equal(2))
		})

		It("RemoveMember decrements connection count; member stays until last", func() {
			room := NewV1Room("B3K7M9", "n", "m_alex")
			room.AddMember(V1Member{Id: "m_alex"})
			room.AddMember(V1Member{Id: "m_alex"})
			Expect(room.RemoveMember("m_alex")).To(BeFalse())
			Expect(room.ConnectionCount("m_alex")).To(Equal(1))
			Expect(room.RemoveMember("m_alex")).To(BeTrue())
			Expect(room.MemberCount()).To(Equal(0))
		})

		It("MemberList returns a snapshot of all members", func() {
			room := NewV1Room("B3K7M9", "n", "m_alex")
			room.AddMember(V1Member{Id: "m_alex"})
			room.AddMember(V1Member{Id: "m_bob"})
			ids := []string{}
			for _, m := range room.MemberList() {
				ids = append(ids, m.Id)
			}
			Expect(ids).To(ConsistOf("m_alex", "m_bob"))
		})
	})
})
