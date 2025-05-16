package api

import (
	"net/http"

	"audio-stream-processor/pipeline"
	"audio-stream-processor/websocket"

	"github.com/gorilla/mux"
)

func NewRouter(p *pipeline.Processor) http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/upload", UploadHandler(p)).Methods("POST")
	r.HandleFunc("/chunks/{id}", GetChunkHandler(p)).Methods("GET")
	r.HandleFunc("/sessions/{user_id}", GetSessionHandler(p)).Methods("GET")
	r.HandleFunc("/ws", websocket.WSHandler(p))

	return r
}
