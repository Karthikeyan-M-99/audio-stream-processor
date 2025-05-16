package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"audio-stream-processor/pipeline"
	"audio-stream-processor/storage"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func UploadHandler(p *pipeline.Processor) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        userID := r.Header.Get("userId")
        sessionID := r.Header.Get("sessionId")
        timestampStr := r.Header.Get("timeStamp")

        if userID == "" || sessionID == "" || timestampStr == "" {
            http.Error(w, "missing required headers", http.StatusBadRequest)
            return
        }

        timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
        if err != nil {
            http.Error(w, "invalid timestamp", http.StatusBadRequest)
            return
        }

        data, err := io.ReadAll(r.Body)
        if err != nil {
            http.Error(w, "failed to read body", http.StatusInternalServerError)
            return
        }
        defer r.Body.Close()

        chunk := &pipeline.AudioChunk{
            storage.AudioChunk{
                ID:        uuid.New().String(),
                UserID:    userID,
                SessionID: sessionID,
                Timestamp: timestamp,
                Data:      data,
            },
        }

        if err := p.Submit(chunk); err != nil {
            http.Error(w, "backpressure: ingestion queue full", http.StatusTooManyRequests)
            return
        }

        w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(chunk)
    }
}




func GetChunkHandler(p *pipeline.Processor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		chunk, err := p.MemStore.GetByID(id)
		if err != nil {
			http.Error(w, "chunk not found", http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(chunk)
	}
}

func GetSessionHandler(p *pipeline.Processor) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := mux.Vars(r)["user_id"]
		chunks, err := p.MemStore.GetByUser(userID)
		if err != nil {
			http.Error(w, "failed to get sessions", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(chunks)
	}
}
