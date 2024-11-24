package repository

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/Zando74/GopherFS/control-plane/config"
	"github.com/Zando74/GopherFS/control-plane/internal/domain/entity"
	"github.com/Zando74/GopherFS/control-plane/logger"
	"github.com/segmentio/kafka-go"
)

type FileChunkRepository struct {
	producer                 *kafka.Writer
	consumer                 *kafka.Reader
	cfg                      *config.Config
	mutex                    sync.Mutex
	chunkSaveSuccessHandlers map[string]func(fileChunkLocation entity.ChunkLocations)
}

func NewFileChunkRepository() *FileChunkRepository {
	cfg := config.ConfigSingleton.GetInstance()
	fcr := &FileChunkRepository{
		producer: &kafka.Writer{
			Addr:        kafka.TCP(cfg.Kafka.Brokers...),
			Topic:       cfg.Kafka.TopicFileChunkSave,
			Balancer:    &kafka.Hash{},
			MaxAttempts: 3,
			BatchSize:   256,
			BatchBytes:  100000000, // 100 MB
		},
		consumer: kafka.NewReader(kafka.ReaderConfig{
			Brokers:     cfg.Kafka.Brokers,
			Topic:       cfg.Kafka.TopicFileChunkRead,
			MaxAttempts: 3,
			MaxBytes:    100000000, // 100 MB
		}),
		chunkSaveSuccessHandlers: make(map[string]func(fileChunkLocation entity.ChunkLocations)),
	}
	fcr.cfg = cfg
	go fcr.ListenForChunkSaveSuccess()
	return fcr
}

func (fcr *FileChunkRepository) ListenForChunkSaveSuccess() {
	for {
		msg, err := fcr.consumer.ReadMessage(context.Background())
		if err != nil {
			logger.LoggerSingleton.GetInstance().Error("Error while reading message: %v", err)
			continue
		}

		var response entity.ChunkLocations
		if err := json.Unmarshal(msg.Value, &response); err != nil {
			logger.LoggerSingleton.GetInstance().Error("Error while unmarshalling message: %v", err)
			continue
		}

		fcr.mutex.Lock()
		if handler, ok := fcr.chunkSaveSuccessHandlers[string(msg.Key)]; ok {
			handler(response)
			delete(fcr.chunkSaveSuccessHandlers, string(msg.Key))
		}
		fcr.mutex.Unlock()
	}
}

func (fcr *FileChunkRepository) SaveFileChunk(chunk *entity.FileChunk, onSuccess func(fileChunkLocation entity.ChunkLocations)) error {
	message, err := json.Marshal(chunk)
	if err != nil {
		return err
	}

	fcr.mutex.Lock()
	fcr.chunkSaveSuccessHandlers[string(chunk.ChunkID)] = onSuccess
	fcr.mutex.Unlock()

	err = fcr.producer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(chunk.ChunkID),
			Value: message,
		},
	)
	if err != nil {
		fcr.mutex.Lock()
		delete(fcr.chunkSaveSuccessHandlers, string(chunk.ChunkID))
		fcr.mutex.Unlock()
		return err
	}

	return nil
}

func (fcr *FileChunkRepository) DeleteAllFilesChunks(fileID string) error {
	message, err := json.Marshal(map[string]string{"fileID": fileID})
	if err != nil {
		return err
	}

	err = fcr.producer.WriteMessages(context.Background(),
		kafka.Message{
			Topic: fcr.cfg.Kafka.TopicFileChunkDelete,
			Key:   []byte(fileID),
			Value: message,
		},
	)
	if err != nil {
		return err
	}

	return nil
}
