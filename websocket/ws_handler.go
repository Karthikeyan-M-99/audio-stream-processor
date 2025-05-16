package websocket

import (
	"log"
	"net/http"
	"time"

	"audio-stream-processor/pipeline"
	"audio-stream-processor/storage"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for demo, restrict in prod
	},
}

func WSHandler(p *pipeline.Processor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("ws upgrade error: %v", err)
			return
		}
		defer conn.Close()

		for {
			mt, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("ws read error: %v", err)
				break
			}

			chunk, err := parseWSMessage(message)
			if err != nil {
				conn.WriteMessage(websocket.TextMessage, []byte("invalid message"))
				continue
			}

			if err := p.Submit(chunk); err != nil {
				conn.WriteMessage(websocket.TextMessage, []byte("backpressure: queue full"))
				continue
			}

			ack := []byte("chunk accepted: " + chunk.ID)
			if err := conn.WriteMessage(mt, ack); err != nil {
				break
			}
		}
	}
}

func parseWSMessage(msg []byte) (*pipeline.AudioChunk, error) {
	return &pipeline.AudioChunk{
		storage.AudioChunk{
			ID:        "",
			UserID:    "ws_user",
			SessionID: "ws_session",
			Timestamp: time.Now().Unix(),
			Data:      msg,
		},
	}, nil
}
