package controller

import (
	"context"
	"time"

	"github.com/Zando74/GopherFS/metadata-service/config"
	"github.com/Zando74/GopherFS/metadata-service/internal/domain/usecase"
	"github.com/Zando74/GopherFS/metadata-service/logger"
	"github.com/segmentio/kafka-go"
)

type OnFileMetadataDelete struct {
	DeleteFileMetadataUseCase usecase.DeleteFileMetadataUseCase
	logger                    logger.Interface
	cfg                       *config.Config
}

func NewOnFileMetadataDelete() *OnFileMetadataDelete {
	return &OnFileMetadataDelete{
		DeleteFileMetadataUseCase: usecase.DeleteFileMetadataUseCase{
			FileMetadataRepository: fileMetadataRepositoryImpl,
		},
		logger: logger.LoggerSingleton.GetInstance(),
		cfg:    config.ConfigSingleton.GetInstance(),
	}
}

func (o *OnFileMetadataDelete) Listen() {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     o.cfg.Kafka.Brokers,
		Topic:       o.cfg.Kafka.TopicFileMetadataDelete,
		MaxBytes:    100000000,
		MaxWait:     1 * time.Second,
		StartOffset: kafka.FirstOffset,
	})
	defer r.Close()

	for {
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			o.logger.Error("could not read message: %v", err)
			continue
		}

		o.processMessage(m)
	}
}

func (o *OnFileMetadataDelete) processMessage(m kafka.Message) {
	fileID := string(m.Key)
	err := o.DeleteFileMetadataUseCase.Execute(fileID)
	if err != nil {
		o.logger.Error(err)
		return
	}
}
