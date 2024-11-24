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

type Location struct {
	StorageID string
}

type ChunkLocations struct {
	ChunkID            string
	SequenceNumber     string
	PrimaryLocation    Location
	SecondaryLocations []Location
}

type FileChunk struct {
	FileID         string
	SequenceNumber string
	ChunkID        string
	ChunkData      string
}

func main() {
	// Define Kafka broker address and topic
	brokerAddress := "kafka:9092"
	topicFileChunkSave := "emit-file-chunk-save"
	topicFileChunkRead := "emit-file-chunk-read"

	// Create a new Kafka reader with improved configuration
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{brokerAddress},
		Topic:       topicFileChunkSave,
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
		Topic:        topicFileChunkRead,
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

	// Ensure all messages are processed before closing the channel
	//wg.Wait()
	//close(messageChan)
}

func processMessage(m kafka.Message, w *kafka.Writer) {
	// Process the message
	fmt.Printf("Received message chunk id: %s\n", string(m.Key))

	var fileChunk FileChunk
	if err := json.Unmarshal(m.Value, &fileChunk); err != nil {
		log.Printf("could not unmarshal file chunk: %v", err)
		return
	}

	// Emit a response message in the same channel
	chunkLocation := ChunkLocations{
		ChunkID:        fileChunk.ChunkID,
		SequenceNumber: fileChunk.SequenceNumber,
		PrimaryLocation: Location{
			StorageID: "primary-storage-id", // Set the appropriate primary storage ID here
		},
		SecondaryLocations: []Location{},
	}

	chunkLocationBytes, err := json.Marshal(chunkLocation)
	if err != nil {
		log.Printf("could not marshal chunk location: %v", err)
		return
	}

	responseMessage := kafka.Message{
		Key:   m.Key,
		Value: chunkLocationBytes,
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

	// Here you can add your logic to handle the file chunk save event
	// For example, save the chunk to your storage service
}
