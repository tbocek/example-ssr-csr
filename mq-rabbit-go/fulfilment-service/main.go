package main

import (
	"log"
	"os"
	"time"

	"github.com/streadway/amqp"
)

func main() {
	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	if rabbitMQURL == "" {
		rabbitMQURL = "amqp://guest:guest@rabbitmq:5672/"
	}

	// Connect with retry
	var conn *amqp.Connection
	var err error
	for i := 0; i < 120; i++ {
		conn, err = amqp.Dial(rabbitMQURL)
		if err == nil {
			log.Println("Connected to RabbitMQ")
			break
		}
		time.Sleep(250 * time.Millisecond)
	}
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open channel: %v", err)
	}
	defer ch.Close()

	// Declare fulfillment queue
	queue, err := ch.QueueDeclare(
		"fulfillment_queue",
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	// Declare exchange for fan-out
	err = ch.ExchangeDeclare(
		"game_events", // name
		"fanout",      // type
		true,          // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		log.Fatalf("Failed to declare exchange: %v", err)
	}

	// Bind queue to exchange
	// Backend publishes to exchange
	// Exchange copies message to all bound queues
	// Each service consumes from its own queue
	err = ch.QueueBind(
		queue.Name,    // queue name
		"",            // routing key
		"game_events", // exchange
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to bind queue: %v", err)
	}

	// Start consumer
	msgs, err := ch.Consume(
		queue.Name,
		"",    // consumer tag
		false, // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		log.Fatalf("Failed to register consumer: %v", err)
	}

	log.Println("📦 Fulfillment service started, waiting for messages...")

	forever := make(chan bool)
	go func() {
		for d := range msgs {
			log.Printf("📦 Processing fulfillment: %s", d.Body)

			// Simulate fulfillment processing
			time.Sleep(150 * time.Millisecond)

			d.Ack(false)
			log.Printf("✅ Fulfillment processed")
		}
		log.Println("❌ Fulfillment message channel closed")
		forever <- true
	}()

	<-forever
}
