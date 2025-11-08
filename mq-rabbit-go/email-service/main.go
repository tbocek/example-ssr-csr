package main

import (
	"log"
	"os"
	"time"

	"github.com/streadway/amqp"
)

func main() {
	log.Println("starting.... service started, waiting for messages...")
	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	if rabbitMQURL == "" {
		rabbitMQURL = "amqp://guest:guest@rabbitmq:5672/"
	}

	conn, err := amqp.Dial(rabbitMQURL)
	for range 120 {
		if err == nil {
			break
		}
		time.Sleep(250 * time.Millisecond)
		conn, err = amqp.Dial(rabbitMQURL)
	}
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer ch.Close()

	queue, err := ch.QueueDeclare(
		"email_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare a queue: %v", err)
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

	msgs, err := ch.Consume(
		queue.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	log.Println("Email service started, waiting for messages...")

	forever := make(chan bool)
	go func() {
		for d := range msgs {
			log.Printf("📧 Processing email: %s", d.Body)

			// Simulate processing time
			time.Sleep(100 * time.Millisecond)

			// Acknowledge message after processing
			d.Ack(false)
			log.Printf("✅ Email processed and acknowledged")
		}
		log.Println("❌ Message channel closed")
		forever <- true
	}()

	<-forever
}
