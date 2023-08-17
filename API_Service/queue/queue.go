package queue

import (
	"context"
	"encoding/json"

	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type QueueMessage struct {
	TraceID       string `json:"trace_id"`
	ParentTraceID string `json:"parent_trace_id"`
	Email         string `json:"email"`
}

func EnqueueMessage(ctx context.Context, message QueueMessage) error {
	const spanID = "EnqueueMessage"
	ctx, span := otel.Tracer(spanID).Start(ctx, spanID)
	defer span.End()

	log.SetFormatter(&log.JSONFormatter{})
	l := log.WithFields(log.Fields{
		"parent_id": message.ParentTraceID,
		"trace_id":  message.TraceID,
		"span_id":   spanID,
		"email":     message.Email,
	})
	l.Info("Enqueue message !")

	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		l.Errorf("Failed to connect to RabbitMQ: %v", err)
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		l.Errorf("Failed to open a channel: %v", err)
		return err
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
		l.Errorf("Failed to declare a queue: %v", err)
		return err
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		l.Errorf("Failed to marshal message to JSON: %v", err)
		return err
	}

	body := amqp.Publishing{
		ContentType: "text/plain",
		Body:        messageBytes,
	}

	headers := make(map[string]string)
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(headers))

	bodyHeader := make(amqp.Table)
	for key, value := range headers {
		bodyHeader[key] = value
	}

	body.Headers = bodyHeader
	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,
		false,
		body)
	if err != nil {
		l.Errorf("Failed to publish message: %v", err)
		return err
	}

	return nil
}
