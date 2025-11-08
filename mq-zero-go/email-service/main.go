package main

import (
	"log"
	"os"
	"time"

	zmq "github.com/pebbe/zmq4"
)

func main() {
	// Get publisher address from environment or use default
	publisherAddr := os.Getenv("PUBLISHER_ADDR")
	if publisherAddr == "" {
		publisherAddr = "tcp://backend:5557"
	}

	// Create ZeroMQ context and subscriber socket
	context, err := zmq.NewContext()
	if err != nil {
		log.Fatalf("Failed to create ZMQ context: %v", err)
	}
	defer context.Term()

	subscriber, err := context.NewSocket(zmq.PULL) //SUB
	if err != nil {
		log.Fatalf("Failed to create subscriber socket: %v", err)
	}
	defer subscriber.Close()

	// Connect to publisher with retry
	for i := 0; i < 120; i++ {
		err = subscriber.Connect(publisherAddr)
		if err == nil {
			log.Printf("Connected to publisher at %s", publisherAddr)
			break
		}
		log.Printf("Connection attempt %d failed: %v", i+1, err)
		time.Sleep(250 * time.Millisecond)
	}
	if err != nil {
		log.Fatalf("Failed to connect to publisher after retries: %v", err)
	}

	// Subscribe to all messages (empty prefix = all topics)
	//err = subscriber.SetSubscribe("")
	//if err != nil {
	//	log.Fatalf("Failed to set subscription: %v", err)
	//}

	log.Println("📧 Email service started, waiting for messages...")

	// Listen for messages
	for {
		// Receive message (blocks until message arrives)
		message, err := subscriber.Recv(0)
		if err != nil {
			log.Printf("Error receiving message: %v", err)
			continue
		}

		log.Printf("📧 Processing email: %s", message)
		
		// Simulate email processing
		time.Sleep(100 * time.Millisecond)
		
		log.Printf("✅ Email sent successfully")
	}
}