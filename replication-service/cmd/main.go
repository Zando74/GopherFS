package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

type FileReplicationRequest struct {
	FileID string
}

func main() {
	brokerAddress := "kafka:9092"
	topicFileReplicationRequestSave := "emit-file-replication-request-save"
	topicFileReplicationRequestRead := "emit-file-replication-request-read"

	// Create a new Kafka reader with improved configuration
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{brokerAddress},
		Topic:       topicFileReplicationRequestSave,
		Partition:   0,
		MinBytes:    10e3,      // 10KB
		MaxBytes:    100000000, // 100MB
		MaxWait:     1 * time.Second,
		StartOffset: kafka.FirstOffset,
	})
	defer r.Close()

	// Create a new Kafka writer with improved configuration
	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      []string{brokerAddress},
		Topic:        topicFileReplicationRequestRead,
		Balancer:     &kafka.LeastBytes{},
		BatchTimeout: 1 * time.Second,
	})
	defer w.Close()

	fmt.Println("Starting Kafka listener for file chunk save events...")

	var wg sync.WaitGroup
	workerCount := 15 // Number of concurrent workers
	messageChan := make(chan kafka.Message, workerCount)

	// Start workers
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for m := range messageChan {
				processMessage(m, w)
			}
		}()
	}

	for {
		// Read a message from Kafka
		m, err := r.ReadMessage(context.Background())
		if err != nil {
			log.Printf("could not read message: %v", err)
			continue
		}

		messageChan <- m // Send the message to a worker
	}
}

func processMessage(m kafka.Message, w *kafka.Writer) {
	// Process the message
	fmt.Printf("Received message chunk id: %s\n", string(m.Key))

	var fileReplicationRequest FileReplicationRequest
	if err := json.Unmarshal(m.Value, &fileReplicationRequest); err != nil {
		log.Printf("could not unmarshal file chunk: %v", err)
		return
	}

	fileReplicationResponseBytes, err := json.Marshal(fileReplicationRequest)
	if err != nil {
		log.Printf("could not marshal chunk location: %v", err)
		return
	}

	responseMessage := kafka.Message{
		Key:   m.Key,
		Value: fileReplicationResponseBytes,
	}

	retryCount := 0
	for {
		err = w.WriteMessages(context.Background(), responseMessage)
		if err != nil {
			log.Printf("could not write response message on attempt %d: %v", retryCount+1, err)
			retryCount++
			if retryCount >= 5 {
				log.Printf("max retry attempts reached, message dropped: %v", err)
				break
			}
			time.Sleep(time.Duration(retryCount) * time.Second) // Exponential backoff
		} else {
			break
		}
	}

}
