package repository

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	"github.com/Zando74/GopherFS/control-plane/config"
	"github.com/Zando74/GopherFS/control-plane/internal/domain/entity"
	"github.com/segmentio/kafka-go"
)

type FileMetadataRepository struct {
	producer                    *kafka.Writer
	consumer                    *kafka.Reader
	cfg                         *config.Config
	mutex                       sync.Mutex
	metadataSaveSuccessHandlers map[string]func()
}

func NewFileMetadataRepository() *FileMetadataRepository {
	cfg := config.ConfigSingleton.GetInstance()
	fmr := &FileMetadataRepository{
		producer: &kafka.Writer{
			Addr:        kafka.TCP(cfg.Kafka.Brokers...),
			Topic:       cfg.Kafka.TopicFileMetadataSave,
			Balancer:    &kafka.Hash{},
			MaxAttempts: 3,
			BatchSize:   256,
			BatchBytes:  100000000, // 100 MB
		},
		consumer: kafka.NewReader(kafka.ReaderConfig{
			Brokers:     cfg.Kafka.Brokers,
			Topic:       cfg.Kafka.TopicFileMetadataRead,
			MaxAttempts: 3,
			MaxBytes:    100000000, // 100 MB
		}),
		metadataSaveSuccessHandlers: make(map[string]func()),
	}
	go fmr.ListenForMetadataSaveSuccess()
	return fmr
}

func (fmr *FileMetadataRepository) ListenForMetadataSaveSuccess() {
	for {
		msg, err := fmr.consumer.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error while reading message: %v", err)
			continue
		}

		fmr.mutex.Lock()
		if handler, ok := fmr.metadataSaveSuccessHandlers[string(msg.Key)]; ok {
			handler()
			delete(fmr.metadataSaveSuccessHandlers, string(msg.Key))
		}
		fmr.mutex.Unlock()
	}
}

func (fmr *FileMetadataRepository) SaveFileMetadata(metadata *entity.FileMetadata, onSuccess func()) error {
	message, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	fmr.mutex.Lock()
	fmr.metadataSaveSuccessHandlers[string(metadata.FileID)] = onSuccess
	fmr.mutex.Unlock()

	err = fmr.producer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(metadata.FileID),
			Value: message,
		},
	)
	if err != nil {
		fmr.mutex.Lock()
		delete(fmr.metadataSaveSuccessHandlers, string(metadata.FileID))
		fmr.mutex.Unlock()
		return err
	}

	return nil
}

func (fmr *FileMetadataRepository) DeleteFileMetadata(fileID string) error {
	message, err := json.Marshal(map[string]string{"fileID": fileID})
	if err != nil {
		return err
	}

	err = fmr.producer.WriteMessages(context.Background(),
		kafka.Message{
			Topic: fmr.cfg.Kafka.TopicFileMetadataDelete,
			Key:   []byte(fileID),
			Value: message,
		},
	)
	if err != nil {
		return err
	}

	return nil
}
