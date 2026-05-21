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

	Describe("chat message", func() {
		It("broadcasts chat to all clients including the sender", func() {
			conn1, _ := dialV1Ws(srv.URL, room.Id)
			defer conn1.Close()
			conn1.WriteMessage(websocket.TextMessage, []byte(
				`{"type":"join","data":{"member":{"id":"m_alex","displayName":"Alex"}}}`))
			readV1Msg(conn1) // room-state

			conn2, _ := dialV1Ws(srv.URL, room.Id)
			defer conn2.Close()
			conn2.WriteMessage(websocket.TextMessage, []byte(
				`{"type":"join","data":{"member":{"id":"m_bob","displayName":"Bob"}}}`))
			readV1Msg(conn2) // room-state
			readV1Msg(conn1) // member-joined for Bob

			conn1.WriteMessage(websocket.TextMessage, []byte(
				`{"type":"chat","data":{"text":"hello world"}}`))

			msg, err := readV1Msg(conn2)
			Expect(err).ToNot(HaveOccurred())
			Expect(msg["type"]).To(Equal("chat"))
			message := msg["data"].(map[string]any)["message"].(map[string]any)
			Expect(message["memberId"]).To(Equal("m_alex"))
			Expect(message["text"]).To(Equal("hello world"))
			Expect(message["id"]).ToNot(BeNil())

			echoed, _ := readV1Msg(conn1)
			Expect(echoed["type"]).To(Equal("chat"))
		})
	})

	Describe("typing indicators", func() {
		It("broadcasts typing-start to other clients", func() {
			conn1, _ := dialV1Ws(srv.URL, room.Id)
			defer conn1.Close()
			conn1.WriteMessage(websocket.TextMessage, []byte(
				`{"type":"join","data":{"member":{"id":"m_alex","displayName":"Alex"}}}`))
			readV1Msg(conn1)

			conn2, _ := dialV1Ws(srv.URL, room.Id)
			defer conn2.Close()
			conn2.WriteMessage(websocket.TextMessage, []byte(
				`{"type":"join","data":{"member":{"id":"m_bob","displayName":"Bob"}}}`))
			readV1Msg(conn2)
			readV1Msg(conn1) // member-joined

			conn1.WriteMessage(websocket.TextMessage, []byte(`{"type":"typing-start"}`))
			msg, err := readV1Msg(conn2)
			Expect(err).ToNot(HaveOccurred())
			Expect(msg["type"]).To(Equal("typing-start"))
			Expect(msg["data"].(map[string]any)["memberId"]).To(Equal("m_alex"))
		})

		It("broadcasts typing-stop to other clients", func() {
			conn1, _ := dialV1Ws(srv.URL, room.Id)
			defer conn1.Close()
			conn1.WriteMessage(websocket.TextMessage, []byte(
				`{"type":"join","data":{"member":{"id":"m_alex","displayName":"Alex"}}}`))
			readV1Msg(conn1)

			conn2, _ := dialV1Ws(srv.URL, room.Id)
			defer conn2.Close()
			conn2.WriteMessage(websocket.TextMessage, []byte(
				`{"type":"join","data":{"member":{"id":"m_bob","displayName":"Bob"}}}`))
			readV1Msg(conn2)
			readV1Msg(conn1)

			conn1.WriteMessage(websocket.TextMessage, []byte(`{"type":"typing-start"}`))
			readV1Msg(conn2) // typing-start

			conn1.WriteMessage(websocket.TextMessage, []byte(`{"type":"typing-stop"}`))
			msg, _ := readV1Msg(conn2)
			Expect(msg["type"]).To(Equal("typing-stop"))
		})
	})

	Describe("reaction-burst", func() {
		It("broadcasts burst to all clients including the sender", func() {
			conn1, _ := dialV1Ws(srv.URL, room.Id)
			defer conn1.Close()
			conn1.WriteMessage(websocket.TextMessage, []byte(
				`{"type":"join","data":{"member":{"id":"m_alex","displayName":"Alex"}}}`))
			readV1Msg(conn1)

			conn2, _ := dialV1Ws(srv.URL, room.Id)
			defer conn2.Close()
			conn2.WriteMessage(websocket.TextMessage, []byte(
				`{"type":"join","data":{"member":{"id":"m_bob","displayName":"Bob"}}}`))
			readV1Msg(conn2)
			readV1Msg(conn1)

			conn1.WriteMessage(websocket.TextMessage, []byte(
				`{"type":"reaction-burst","data":{"emoji":"🔥"}}`))

			msg, err := readV1Msg(conn2)
			Expect(err).ToNot(HaveOccurred())
			Expect(msg["type"]).To(Equal("reaction-burst"))
			data := msg["data"].(map[string]any)
			Expect(data["memberId"]).To(Equal("m_alex"))
			Expect(data["emoji"]).To(Equal("🔥"))
		})
	})

	Describe("settings-update (host-only)", func() {
		It("accepts the patch from host and broadcasts settings-updated", func() {
			conn1, _ := dialV1Ws(srv.URL, room.Id)
			defer conn1.Close()
			conn1.WriteMessage(websocket.TextMessage, []byte(
				`{"type":"join","data":{"member":{"id":"m_alex","displayName":"Alex"}}}`))
			readV1Msg(conn1)

			conn1.WriteMessage(websocket.TextMessage, []byte(
				`{"type":"settings-update","data":{"controlMode":"host-only"}}`))

			msg, err := readV1Msg(conn1)
			Expect(err).ToNot(HaveOccurred())
			Expect(msg["type"]).To(Equal("settings-updated"))
			Expect(msg["data"].(map[string]any)["controlMode"]).To(Equal("host-only"))

			updated, _ := v1Srv.GetRoom(room.Id)
			Expect(updated.ControlMode).To(Equal("host-only"))
		})

		It("rejects settings-update from a non-host (no broadcast, no change)", func() {
			conn1, _ := dialV1Ws(srv.URL, room.Id)
			defer conn1.Close()
			conn1.WriteMessage(websocket.TextMessage, []byte(
				`{"type":"join","data":{"member":{"id":"m_bob","displayName":"Bob"}}}`))
			readV1Msg(conn1)

			conn1.WriteMessage(websocket.TextMessage, []byte(
				`{"type":"settings-update","data":{"controlMode":"host-only"}}`))

			conn1.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
			_, _, err := conn1.ReadMessage()
			Expect(err).To(HaveOccurred())

			updated, _ := v1Srv.GetRoom(room.Id)
			Expect(updated.ControlMode).To(Equal("democratic"))
		})
	})

	Describe("rename-room (host-only)", func() {
		It("renames and broadcasts room-renamed when host", func() {
			conn1, _ := dialV1Ws(srv.URL, room.Id)
			defer conn1.Close()
			conn1.WriteMessage(websocket.TextMessage, []byte(
				`{"type":"join","data":{"member":{"id":"m_alex","displayName":"Alex"}}}`))
			readV1Msg(conn1)

			conn1.WriteMessage(websocket.TextMessage, []byte(
				`{"type":"rename-room","data":{"name":"Pizza & a movie"}}`))

			msg, _ := readV1Msg(conn1)
			Expect(msg["type"]).To(Equal("room-renamed"))
			Expect(msg["data"].(map[string]any)["name"]).To(Equal("Pizza & a movie"))

			updated, _ := v1Srv.GetRoom(room.Id)
			Expect(updated.Name).To(Equal("Pizza & a movie"))
		})
	})
})
