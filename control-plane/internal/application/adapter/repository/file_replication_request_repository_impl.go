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

type FileReplicationRequestRepository struct {
	producer        *kafka.Writer
	consumer        *kafka.Reader
	cfg             *config.Config
	mutex           sync.Mutex
	successHandlers map[string]func()
}

func NewFileReplicationRequestRepository() *FileReplicationRequestRepository {
	cfg := config.ConfigSingleton.GetInstance()
	frr := &FileReplicationRequestRepository{
		producer: &kafka.Writer{
			Addr:        kafka.TCP(cfg.Kafka.Brokers...),
			Topic:       cfg.Kafka.TopicFileReplicationRequestSave,
			Balancer:    &kafka.Hash{},
			MaxAttempts: 3,
			BatchSize:   256,
			BatchBytes:  100000000, // 100 MB
		},
		consumer: kafka.NewReader(kafka.ReaderConfig{
			Brokers:     cfg.Kafka.Brokers,
			Topic:       cfg.Kafka.TopicFileReplicationRequestRead,
			MaxAttempts: 3,
			MaxBytes:    100000000, // 100 MB
		}),
		cfg:             cfg,
		successHandlers: make(map[string]func()),
	}
	go frr.ListenForReplicationSuccess()
	frr.cfg = cfg
	return frr
}

func (frr *FileReplicationRequestRepository) ListenForReplicationSuccess() {
	for {
		msg, err := frr.consumer.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Error while reading message: %v", err)
			continue
		}

		frr.mutex.Lock()
		if handler, ok := frr.successHandlers[string(msg.Key)]; ok {
			handler()
			delete(frr.successHandlers, string(msg.Key))
		}
		frr.mutex.Unlock()
	}
}

func (frr *FileReplicationRequestRepository) SaveReplicationRequest(replicationRequest *entity.FileReplicationRequest, onSuccess func()) error {
	message, err := json.Marshal(replicationRequest)
	if err != nil {
		return err
	}

	frr.mutex.Lock()
	frr.successHandlers[string(replicationRequest.FileID)] = onSuccess
	frr.mutex.Unlock()

	err = frr.producer.WriteMessages(context.Background(),
		kafka.Message{
			Key:   []byte(replicationRequest.FileID),
			Value: message,
		},
	)
	if err != nil {
		frr.mutex.Lock()
		delete(frr.successHandlers, string(replicationRequest.FileID))
		frr.mutex.Unlock()
		return err
	}

	return nil
}

func (frr *FileReplicationRequestRepository) DeleteReplicationRequest(replicationRequest *entity.FileReplicationRequest) error {
	message, err := json.Marshal(replicationRequest)
	if err != nil {
		return err
	}

	err = frr.producer.WriteMessages(context.Background(),
		kafka.Message{
			Topic: frr.cfg.Kafka.TopicFileReplicationRequestDelete,
			Key:   []byte(replicationRequest.FileID),
			Value: message,
		},
	)
	if err != nil {
		return err
	}

	return nil
}
