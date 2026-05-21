package main

import (
	"fmt"
	"strings"

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

	Describe("Chat ring buffer", func() {
		It("AppendChat stores the message with id + timestamp", func() {
			room := NewV1Room("B3K7M9", "n", "m_alex")
			msg := room.AppendChat("m_alex", "hello")
			Expect(msg.Id).ToNot(BeEmpty())
			Expect(msg.MemberId).To(Equal("m_alex"))
			Expect(msg.Text).To(Equal("hello"))
			Expect(msg.Timestamp).To(BeNumerically(">", 0))
		})

		It("ChatHistory returns all messages in order", func() {
			room := NewV1Room("B3K7M9", "n", "m_alex")
			room.AppendChat("m_a", "first")
			room.AppendChat("m_b", "second")
			room.AppendChat("m_c", "third")
			history := room.ChatHistory()
			Expect(history).To(HaveLen(3))
			Expect(history[0].Text).To(Equal("first"))
			Expect(history[2].Text).To(Equal("third"))
		})

		It("ChatHistory retains only the last 100 messages", func() {
			room := NewV1Room("B3K7M9", "n", "m_alex")
			for i := 0; i < 150; i++ {
				room.AppendChat("m_a", fmt.Sprintf("msg %d", i))
			}
			history := room.ChatHistory()
			Expect(history).To(HaveLen(100))
			Expect(history[0].Text).To(Equal("msg 50"))
			Expect(history[99].Text).To(Equal("msg 149"))
		})

		It("AppendChat caps text at 500 characters", func() {
			room := NewV1Room("B3K7M9", "n", "m_alex")
			msg := room.AppendChat("m_alex", strings.Repeat("a", 600))
			Expect(msg.Text).To(HaveLen(500))
		})
	})
})
