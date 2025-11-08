package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
)

type GameEvent struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Stars       int    `json:"stars"`
}

type Message struct {
	MsgID       int64           `json:"msg_id"`
	ReadCount   int64           `json:"read_count"`
	EnqueuedAt  time.Time       `json:"enqueued_at"`
	VT          time.Time       `json:"vt"`
	Message     json.RawMessage `json:"message"`
}

var db *sql.DB

func main() {
	var err error

	// Database setup
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@db:5432/gamedb?sslmode=disable"
	}

	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Wait for database
	for range 120 {
		if err := db.Ping(); err == nil {
			log.Println("Connected to database")
			break
		}
		time.Sleep(250 * time.Millisecond)
	}
	defer db.Close()

	
	queueName := "email_queue"
	log.Printf("📧 Email service started, consuming from queue: %s", queueName)

	// Handle graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start consuming messages
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				// Read message with 30-second visibility timeout using SQL
				var message Message
				err := db.QueryRow(`
					SELECT msg_id, read_ct, enqueued_at, vt, message 
					FROM pgmq.read($1, $2, $3)
				`, queueName, 30, 1).Scan(
					&message.MsgID, 
					&message.ReadCount, 
					&message.EnqueuedAt, 
					&message.VT, 
					&message.Message,
				)

				if err != nil {
					if err == sql.ErrNoRows {
						// No messages available, wait and try again
						time.Sleep(1 * time.Second)
						continue
					}
					log.Printf("❌ Error reading message: %v", err)
					time.Sleep(5 * time.Second)
					continue
				}

				log.Printf("📧 Processing email notification (ID: %d): %s", 
					message.MsgID, string(message.Message))

				// Parse the game event
				var gameEvent GameEvent
				if err := json.Unmarshal(message.Message, &gameEvent); err != nil {
					log.Printf("❌ Failed to parse message: %v", err)
					// Delete malformed message
					db.Exec("SELECT pgmq.delete($1, $2)", queueName, message.MsgID)
					continue
				}

				// Simulate email processing
				time.Sleep(100 * time.Millisecond)

				// Archive the message (keeps a record)
				var archived bool
				err = db.QueryRow("SELECT pgmq.archive($1::text, $2::bigint)", queueName, message.MsgID).Scan(&archived)
				if err != nil {
					log.Printf("❌ Failed to archive message %d: %v", message.MsgID, err)
					continue
				}

				if archived {
					log.Printf("✅ Email sent successfully for game: %s (archived message %d)",
						gameEvent.Title, message.MsgID)
				}
			}
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	log.Println("📧 Shutting down email service...")
	cancel()
}