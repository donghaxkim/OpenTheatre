package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/OpenTheatre/OpenTheatre/internal/qps"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/unrolled/render"
)

var _ = Describe("V1 HTTP", func() {
	var v1Srv *V1Service
	var srv *httptest.Server

	BeforeEach(func() {
		vtSrv := NewOpenTheatreService(time.Minute * 3)
		v1Srv = NewV1Service()
		api := newSlashFix(render.New(), vtSrv, v1Srv, qps.NewQP(time.Second, 3600), http.DefaultClient)
		srv = httptest.NewServer(api)
	})

	AfterEach(func() {
		srv.Close()
	})

	Describe("POST /v1/rooms", func() {
		It("creates a room and returns a 6-char id", func() {
			body := `{"member":{"id":"m_alex","displayName":"Alex","avatarColor":"#ff6b9d"}}`
			resp, err := http.Post(srv.URL+"/v1/rooms", "application/json", bytes.NewReader([]byte(body)))
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(200))
			var out map[string]string
			Expect(json.NewDecoder(resp.Body).Decode(&out)).To(Succeed())
			Expect(out["roomId"]).To(HaveLen(6))
		})

		It("defaults room name to <displayName>'s room", func() {
			body := `{"member":{"id":"m_alex","displayName":"Alex","avatarColor":"#fff"}}`
			resp, _ := http.Post(srv.URL+"/v1/rooms", "application/json", bytes.NewReader([]byte(body)))
			var out map[string]string
			json.NewDecoder(resp.Body).Decode(&out)
			room, _ := v1Srv.GetRoom(out["roomId"])
			Expect(room.Name).To(Equal("Alex's room"))
		})

		It("rejects requests with no displayName", func() {
			body := `{"member":{"id":"m_alex","avatarColor":"#fff"}}`
			resp, _ := http.Post(srv.URL+"/v1/rooms", "application/json", bytes.NewReader([]byte(body)))
			Expect(resp.StatusCode).To(Equal(400))
		})
	})

	Describe("GET /v1/rooms/{id}", func() {
		It("returns the room when it exists", func() {
			room := v1Srv.CreateRoom("Movie night", "m_alex")
			resp, err := http.Get(srv.URL + "/v1/rooms/" + room.Id)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(200))
			var out V1Room
			Expect(json.NewDecoder(resp.Body).Decode(&out)).To(Succeed())
			Expect(out.Id).To(Equal(room.Id))
			Expect(out.Name).To(Equal("Movie night"))
		})

		It("returns 404 when the room does not exist", func() {
			resp, _ := http.Get(srv.URL + "/v1/rooms/ZZZZZZ")
			Expect(resp.StatusCode).To(Equal(404))
		})
	})
})
