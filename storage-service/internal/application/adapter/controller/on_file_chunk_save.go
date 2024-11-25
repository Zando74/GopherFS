package controller

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/Zando74/GopherFS/storage-service/config"
	"github.com/Zando74/GopherFS/storage-service/internal/domain/entity"
	"github.com/Zando74/GopherFS/storage-service/internal/domain/factory"
	"github.com/Zando74/GopherFS/storage-service/internal/domain/usecase"
	"github.com/Zando74/GopherFS/storage-service/logger"
	"github.com/segmentio/kafka-go"
)

type OnFileChunkSave struct {
	storeFileChunkSaveUseCase usecase.StoreFileChunkUseCase
	logger                    logger.Interface
	cfg                       *config.Config
}

func NewOnFileChunkSave() *OnFileChunkSave {
	return &OnFileChunkSave{
		storeFileChunkSaveUseCase: usecase.StoreFileChunkUseCase{
			FileChunkRepository: fileChunkRepositoryImpl,
			FileChunkFactory:    factory.FileChunkFactory{},
		},
	}
}

func (o *OnFileChunkSave) Listen() {
	o.cfg = config.ConfigSingleton.GetInstance()
	o.logger = logger.LoggerSingleton.GetInstance()
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     o.cfg.Kafka.Brokers,
		Topic:       o.cfg.Kafka.TopicFileChunkSave,
		MaxBytes:    100000000,
		MaxWait:     1 * time.Second,
		StartOffset: kafka.FirstOffset,
	})
	defer r.Close()

	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      o.cfg.Kafka.Brokers,
		Topic:        o.cfg.Kafka.TopicFileChunkRead,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 1 * time.Second,
	})
	defer w.Close()

	var wg sync.WaitGroup
	workerCount := 15
	messageChan := make(chan kafka.Message, workerCount)

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for m := range messageChan {
				o.processMessage(m, w)
			}
		}()
	}

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			o.logger.Error("could not read message: %v", err)
			continue
		}

		messageChan <- m
	}
}

func (o *OnFileChunkSave) processMessage(m kafka.Message, w *kafka.Writer) {
	var fileChunk entity.FileChunk
	if err := json.Unmarshal(m.Value, &fileChunk); err != nil {
		o.logger.Error("could not unmarshal file chunk: %v", err)
		return
	}

	err := o.storeFileChunkSaveUseCase.Execute(
		fileChunk.FileID,
		fileChunk.SequenceNumber,
		[]byte(fileChunk.ChunkData),
		func(location entity.ChunkLocations) {
			o.emitResponse(m.Key, location, w)
		})

	if err != nil {
		o.logger.Error("error during file chunk saving: %v", err)
	}
}

func (o *OnFileChunkSave) emitResponse(key []byte, location entity.ChunkLocations, w *kafka.Writer) {
	chunkLocationBytes, err := json.Marshal(location)
	if err != nil {
		o.logger.Error("could not marshal chunk location: %v", err)
		return
	}

	responseMessage := kafka.Message{
		Key:   key,
		Value: chunkLocationBytes,
	}

	retryCount := 0
	for {
		err = w.WriteMessages(context.Background(), responseMessage)
		if err != nil {
			o.logger.Error("could not write response message on attempt %d: %v", retryCount+1, err)
			retryCount++
			if retryCount >= 5 {
				o.logger.Error("max retry attempts reached, message dropped: %v", err)
				break
			}
			time.Sleep(time.Duration(retryCount) * time.Second)
		} else {
			break
		}
	}
}
