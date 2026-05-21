package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/OpenTheatre/OpenTheatre/internal/qps"
	"github.com/gorilla/websocket"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/unrolled/render"
)

// dialV1Ws opens a WS to /v1/ws/{roomId} on the test server.
func dialV1Ws(srvURL, roomId string) (*websocket.Conn, error) {
	wsURL := "ws" + strings.TrimPrefix(srvURL, "http") + "/v1/ws/" + roomId
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	return conn, err
}

// readV1Msg reads one JSON-framed message with a 1s deadline.
func readV1Msg(conn *websocket.Conn) (map[string]any, error) {
	conn.SetReadDeadline(time.Now().Add(time.Second))
	_, raw, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	return m, nil
}

var _ = Describe("V1 WebSocket", func() {
	var v1Srv *V1Service
	var srv *httptest.Server
	var room *V1Room

	BeforeEach(func() {
		vtSrv := NewOpenTheatreService(time.Minute * 3)
		v1Srv = NewV1Service()
		api := newSlashFix(render.New(), vtSrv, v1Srv, qps.NewQP(time.Second, 3600), http.DefaultClient)
		srv = httptest.NewServer(api)
		room = v1Srv.CreateRoom("Movie night", "m_alex")
	})

	AfterEach(func() {
		srv.Close()
	})

	Describe("connection", func() {
		It("accepts a connection for an existing room", func() {
			conn, err := dialV1Ws(srv.URL, room.Id)
			Expect(err).ToNot(HaveOccurred())
			conn.Close()
		})

		It("rejects a connection for a non-existent room with 404", func() {
			wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/v1/ws/ZZZZZZ"
			_, resp, err := websocket.DefaultDialer.Dial(wsURL, nil)
			Expect(err).To(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(404))
		})
	})

	Describe("envelope format", func() {
		It("ignores garbage messages without dropping the connection", func() {
			conn, _ := dialV1Ws(srv.URL, room.Id)
			defer conn.Close()
			conn.WriteMessage(websocket.TextMessage, []byte("not json"))
			conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"unknown"}`))
			err := conn.WriteMessage(websocket.TextMessage, []byte(`{"type":"unknown"}`))
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("join message", func() {
		It("returns room-state with the member included", func() {
			conn, _ := dialV1Ws(srv.URL, room.Id)
			defer conn.Close()
			conn.WriteMessage(websocket.TextMessage, []byte(
				`{"type":"join","data":{"member":{"id":"m_alex","displayName":"Alex","avatarColor":"#ff6b9d"}}}`))

			msg, err := readV1Msg(conn)
			Expect(err).ToNot(HaveOccurred())
			Expect(msg["type"]).To(Equal("room-state"))
			data := msg["data"].(map[string]any)
			Expect(data["id"]).To(Equal(room.Id))
			members := data["members"].([]any)
			Expect(members).To(HaveLen(1))
			Expect(members[0].(map[string]any)["id"]).To(Equal("m_alex"))
		})

		It("broadcasts member-joined to other clients", func() {
			conn1, _ := dialV1Ws(srv.URL, room.Id)
			defer conn1.Close()
			conn1.WriteMessage(websocket.TextMessage, []byte(
				`{"type":"join","data":{"member":{"id":"m_alex","displayName":"Alex"}}}`))
			readV1Msg(conn1) // room-state

			conn2, _ := dialV1Ws(srv.URL, room.Id)
			defer conn2.Close()
			conn2.WriteMessage(websocket.TextMessage, []byte(
				`{"type":"join","data":{"member":{"id":"m_bob","displayName":"Bob"}}}`))

			msg, err := readV1Msg(conn1)
			Expect(err).ToNot(HaveOccurred())
			Expect(msg["type"]).To(Equal("member-joined"))
			data := msg["data"].(map[string]any)
			Expect(data["member"].(map[string]any)["id"]).To(Equal("m_bob"))
		})
	})
})
