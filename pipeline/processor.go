package pipeline

import (
	"context"
	"errors"
	"sync"
	"time"

	"audio-stream-processor/storage"

	"github.com/google/uuid"
)

type AudioChunk struct {
	storage.AudioChunk
}

type Processor struct {
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	MemStore  storage.Store
	DiskStore storage.Store

	ingestCh    chan *AudioChunk
	validateCh  chan *AudioChunk
	transformCh chan *AudioChunk
	metadataCh  chan *AudioChunk
	storageCh   chan *AudioChunk

	maxQueueSize int
}

func NewProcessor(memStore, diskStore storage.Store) *Processor {
	ctx, cancel := context.WithCancel(context.Background())
	maxQueue := 100

	return &Processor{
		ctx:          ctx,
		cancel:       cancel,
		MemStore:     memStore,
		DiskStore:    diskStore,
		ingestCh:     make(chan *AudioChunk, maxQueue),
		validateCh:   make(chan *AudioChunk, maxQueue),
		transformCh:  make(chan *AudioChunk, maxQueue),
		metadataCh:   make(chan *AudioChunk, maxQueue),
		storageCh:    make(chan *AudioChunk, maxQueue),
		maxQueueSize: maxQueue,
	}
}

func (p *Processor) Start() {
	p.wg.Add(5)
	go p.ingestWorker()
	go p.validationWorkerPool(5)
	go p.transformationWorkerPool(5)
	go p.metadataExtractionWorkerPool(5)
	go p.storageWorker()
}

func (p *Processor) Stop() {
	p.cancel()
	p.wg.Wait()
}

func (p *Processor) Submit(chunk *AudioChunk) error {
	select {
	case p.ingestCh <- chunk:
		return nil
	default:
		return errors.New("backpressure: ingestion queue full")
	}
}

func (p *Processor) ingestWorker() {
	defer p.wg.Done()
	for {
		select {
		case <-p.ctx.Done():
			return
		case chunk := <-p.ingestCh:
			select {
			case p.validateCh <- chunk:
			case <-p.ctx.Done():
				return
			}
		}
	}
}

func (p *Processor) validationWorkerPool(num int) {
	defer p.wg.Done()
	var wg sync.WaitGroup
	wg.Add(num)
	for i := 0; i < num; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-p.ctx.Done():
					return
				case chunk := <-p.validateCh:
					if p.validate(chunk) {
						select {
						case p.transformCh <- chunk:
						case <-p.ctx.Done():
							return
						}
					}
				}
			}
		}()
	}
	wg.Wait()
}

func (p *Processor) transformationWorkerPool(num int) {
	defer p.wg.Done()
	var wg sync.WaitGroup
	wg.Add(num)
	for i := 0; i < num; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-p.ctx.Done():
					return
				case chunk := <-p.transformCh:
					p.transform(chunk)
					select {
					case p.metadataCh <- chunk:
					case <-p.ctx.Done():
						return
					}
				}
			}
		}()
	}
	wg.Wait()
}

func (p *Processor) metadataExtractionWorkerPool(num int) {
	defer p.wg.Done()
	var wg sync.WaitGroup
	wg.Add(num)
	for i := 0; i < num; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-p.ctx.Done():
					return
				case chunk := <-p.metadataCh:
					p.extractMetadata(chunk)
					select {
					case p.storageCh <- chunk:
					case <-p.ctx.Done():
						return
					}
				}
			}
		}()
	}
	wg.Wait()
}

func (p *Processor) storageWorker() {
	defer p.wg.Done()
	for {
		select {
		case <-p.ctx.Done():
			return
		case chunk := <-p.storageCh:
			p.MemStore.Save(&chunk.AudioChunk)
			p.DiskStore.Save(&chunk.AudioChunk)
		}
	}
}

func (p *Processor) validate(chunk *AudioChunk) bool {
	if chunk.ID == "" {
		chunk.ID = uuid.New().String()
	}
	if chunk.UserID == "" || chunk.SessionID == "" || len(chunk.Data) == 0 {
		return false
	}
	return true
}

func (p *Processor) transform(chunk *AudioChunk) {
	if chunk.Metadata == nil {
		chunk.Metadata = make(map[string]string)
	}
	chunk.Metadata["checksum"] = "deadbeef"
	chunk.Metadata["transformed_at"] = time.Now().Format(time.RFC3339)
}

func (p *Processor) extractMetadata(chunk *AudioChunk) {
	if chunk.Metadata == nil {
		chunk.Metadata = make(map[string]string)
	}
	chunk.Metadata["fake_transcript"] = "Hello World"
	chunk.Metadata["extracted_at"] = time.Now().Format(time.RFC3339)
}
