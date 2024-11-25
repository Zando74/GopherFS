package controller

import (
	"context"
	"sync"
	"time"

	"github.com/Zando74/GopherFS/storage-service/config"
	"github.com/Zando74/GopherFS/storage-service/internal/domain/usecase"
	"github.com/Zando74/GopherFS/storage-service/logger"
	"github.com/segmentio/kafka-go"
)

type OnFileDelete struct {
	DeleteFileUseCase usecase.DeleteFileUseCase
	logger            logger.Interface
	cfg               *config.Config
}

func NewOnFileDelete() *OnFileDelete {
	return &OnFileDelete{
		DeleteFileUseCase: usecase.DeleteFileUseCase{
			FileChunkRepository: fileChunkRepositoryImpl,
		},
		logger: logger.LoggerSingleton.GetInstance(),
		cfg:    config.ConfigSingleton.GetInstance(),
	}
}

func (o *OnFileDelete) Listen() {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     o.cfg.Kafka.Brokers,
		Topic:       o.cfg.Kafka.TopicFileChunkDelete,
		GroupID:     o.cfg.Storage.ServiceName,
		MaxBytes:    100000000,
		MaxWait:     1 * time.Second,
		StartOffset: kafka.FirstOffset,
	})
	defer r.Close()

	var wg sync.WaitGroup
	workerCount := 15
	messageChan := make(chan kafka.Message, workerCount)

	// Start workers
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for m := range messageChan {
				o.processMessage(m)
			}
		}()
	}

	for {
		// Read a message from Kafka
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			o.logger.Error("could not read message: %v", err)
			continue
		}

		messageChan <- m // Send the message to a worker
	}
}

func (o *OnFileDelete) processMessage(m kafka.Message) {
	fileID := string(m.Key)
	err := o.DeleteFileUseCase.Execute(fileID)
	if err != nil {
		o.logger.Error(err)
		return
	}
}
