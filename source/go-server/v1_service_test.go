package main

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("V1Service", func() {
	var svc *V1Service
	BeforeEach(func() {
		svc = NewV1Service()
	})

	Describe("CreateRoom", func() {
		It("returns a room with a 6-char id", func() {
			room := svc.CreateRoom("Alex's room", "m_alex")
			Expect(room.Id).To(HaveLen(6))
			Expect(room.Name).To(Equal("Alex's room"))
			Expect(room.HostId).To(Equal("m_alex"))
		})

		It("stores the room so GetRoom can retrieve it", func() {
			room := svc.CreateRoom("n", "m_alex")
			got, ok := svc.GetRoom(room.Id)
			Expect(ok).To(BeTrue())
			Expect(got).To(Equal(room))
		})

		It("produces distinct ids across calls", func() {
			ids := make(map[string]bool)
			for i := 0; i < 50; i++ {
				ids[svc.CreateRoom("n", "h").Id] = true
			}
			Expect(len(ids)).To(Equal(50))
		})
	})

	Describe("GetRoom", func() {
		It("returns false for an unknown id", func() {
			_, ok := svc.GetRoom("ZZZZZZ")
			Expect(ok).To(BeFalse())
		})
	})

	Describe("DeleteRoom", func() {
		It("removes the room", func() {
			room := svc.CreateRoom("n", "h")
			svc.DeleteRoom(room.Id)
			_, ok := svc.GetRoom(room.Id)
			Expect(ok).To(BeFalse())
		})

		It("is a no-op for an unknown id", func() {
			svc.DeleteRoom("ZZZZZZ")
		})
	})

	Describe("RoomCount", func() {
		It("counts active rooms", func() {
			Expect(svc.RoomCount()).To(Equal(0))
			svc.CreateRoom("a", "h")
			svc.CreateRoom("b", "h")
			Expect(svc.RoomCount()).To(Equal(2))
		})
	})
})
