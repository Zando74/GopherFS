package controller

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/Zando74/GopherFS/metadata-service/config"
	"github.com/Zando74/GopherFS/metadata-service/internal/domain/entity"
	"github.com/Zando74/GopherFS/metadata-service/internal/domain/factory"
	"github.com/Zando74/GopherFS/metadata-service/internal/domain/usecase"
	"github.com/Zando74/GopherFS/metadata-service/logger"
	"github.com/segmentio/kafka-go"
)

type OnFileMetadataSave struct {
	storeFileMetadataUseCase usecase.StoreFileMetadataUseCase
	logger                   logger.Interface
}

func NewOnFileMetadataSave() *OnFileMetadataSave {
	return &OnFileMetadataSave{
		storeFileMetadataUseCase: usecase.StoreFileMetadataUseCase{
			FileMetadataRepository: fileMetadataRepositoryImpl,
			FileMetadataFactory:    factory.FileMetadataFactory{},
		},
	}
}

func (o *OnFileMetadataSave) Listen() {
	cfg := config.ConfigSingleton.GetInstance()
	o.logger = logger.LoggerSingleton.GetInstance()

	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     cfg.Kafka.Brokers,
		Topic:       cfg.Kafka.TopicFileMetadataSave,
		MaxBytes:    100000000,
		MaxWait:     1 * time.Second,
		StartOffset: kafka.FirstOffset,
	})
	defer r.Close()

	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      cfg.Kafka.Brokers,
		Topic:        cfg.Kafka.TopicFileMetadataRead,
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

func (o *OnFileMetadataSave) processMessage(m kafka.Message, w *kafka.Writer) {
	var fileMetadata entity.FileMetadata
	if err := json.Unmarshal(m.Value, &fileMetadata); err != nil {
		o.logger.Error("could not unmarshal file metadata: %v", err)
		return
	}

	err := o.storeFileMetadataUseCase.Execute(
		fileMetadata.FileID,
		fileMetadata.FileName,
		fileMetadata.FileSize,
		fileMetadata.Chunks,
		func() {
			o.emitResponse(m.Key, w)
		})

	if err != nil {
		o.logger.Error("error during file metadata saving: %v", err)
	}
}

func (o *OnFileMetadataSave) emitResponse(key []byte, w *kafka.Writer) {
	responseMessage := kafka.Message{
		Key:   key,
		Value: []byte("OK"),
	}

	retryCount := 0
	for {
		err := w.WriteMessages(context.Background(), responseMessage)
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
