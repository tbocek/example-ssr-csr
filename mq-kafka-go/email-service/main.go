package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/segmentio/kafka-go"
)

func main() {
	// Get Kafka broker addresses from environment
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "kafka:9092"
	}

	// Topic to consume from
	topic := "game-events"

	// Create Kafka reader (consumer)
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     strings.Split(brokers, ","),
		Topic:       topic,
		GroupID:     "email-service-group",
		MinBytes:    10e1, // 1KB
		MaxBytes:    10e6, // 10MB
		MaxWait:     1 * time.Second,
		StartOffset: kafka.LastOffset, // Start from latest messages
	})
	defer reader.Close()

	log.Printf("📧 Email service started, consuming from topic: %s", topic)
	log.Printf("📧 Connected to Kafka brokers: %s", brokers)

	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Listen for interrupt signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("📧 Shutting down email service...")
		cancel()
	}()

	// Consume messages
	for {
		select {
		case <-ctx.Done():
			log.Println("📧 Email service stopped")
			return
		default:
			// Set read deadline to allow checking for shutdown
			ctxWithTimeout, cancelTimeout := context.WithTimeout(ctx, 5*time.Second)
			
			message, err := reader.ReadMessage(ctxWithTimeout)
			cancelTimeout()

			if err != nil {
				if err == context.DeadlineExceeded {
					// Timeout is normal, continue loop
					continue
				}
				log.Printf("No new message: %v", err)
				continue
			}

			log.Printf("📧 Processing email notification: %s", string(message.Value))
			
			// Simulate email processing
			time.Sleep(100 * time.Millisecond)
			
			log.Printf("✅ Email sent successfully (partition: %d, offset: %d)", 
				message.Partition, message.Offset)
		}
	}
}