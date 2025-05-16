package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"audio-stream-processor/api"
	"audio-stream-processor/pipeline"
	"audio-stream-processor/storage"
)

func main() {
	MemStore := storage.NewMemoryStore()
	DiskStore, err := storage.NewDiskStore("data.db")
	if err != nil {
		log.Fatalf("failed to initialize disk store: %v", err)
	}

	proc := pipeline.NewProcessor(MemStore, DiskStore)
	proc.Start()
	defer proc.Stop()

	router := api.NewRouter(proc)

	srv := &http.Server{
		Addr:         ":8081",
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Println("Server starting at :8081")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")
	srv.Close()
}
