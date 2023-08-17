package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	sentryotel "github.com/getsentry/sentry-go/otel"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type QueueMessage struct {
	TraceID       string `json:"trace_id"`
	ParentTraceID string `json:"parent_trace_id"`
	Email         string `json:"email"`
}

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	// Sentry tracing
	err = sentry.Init(sentry.ClientOptions{
		Dsn:              os.Getenv("SENTRY_DSN"),
		Environment:      os.Getenv("ENVIRONMENT"),
		EnableTracing:    true,
		TracesSampleRate: 1.0,
		Debug:            true,
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(sentryotel.NewSentrySpanProcessor()),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(sentryotel.NewSentryPropagator())

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("Failed to open a channel:", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"email_queue", // queue name
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("Failed to declare a queue:", err)
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,
	)
	if err != nil {
		log.Fatal("Failed to register a consumer:", err)
	}

	fmt.Println("Waiting for messages...")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			handleMessage(d)
		}
	}()

	<-forever
}

func handleMessage(d amqp.Delivery) {

	const spanID = "handleSentMailMessage"
	msg := string(d.Body)

	log.SetFormatter(&log.JSONFormatter{})
	l := log.WithFields(log.Fields{
		"message": msg,
		"span_id": spanID,
	})
	l.Info("Received sent mail message !")

	headers := d.Headers
	convertedHeaders := make(map[string]string)
	for k, v := range headers {
		convertedHeaders[k] = v.(string)
	}

	ctx := otel.GetTextMapPropagator().Extract(context.Background(), propagation.MapCarrier(convertedHeaders))
	ctx, span := otel.Tracer("sentEmail").Start(ctx, "sentEmail")

	// sleep 200 miliseconds to simulate a slow process
	time.Sleep(200 * time.Millisecond)

	defer span.End()

	// parse d to QueueMessage struct
	var message QueueMessage
	err := json.Unmarshal(d.Body, &message)
	if err != nil {
		l.Errorf("Failed to unmarshal message: %v", err)
		return
	}

	l.Info("Sent message !")
}
