package queue

import (
	"context"
	"encoding/json"

	"github.com/datphamcode295/distributed-tracing/utils"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type Queue interface {
	EnqueueMessage(ctx context.Context, email string) error
	Close()
}

type RabbitMQ struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
	Queue   amqp.Queue
}

type QueueMessage struct {
	Email string `json:"email"`
}

func New(url string) Queue {
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Errorf("Failed to connect to RabbitMQ: %v", err)
		panic(err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("Failed to open a channel: %v", err)
		panic(err)
	}

	q, err := ch.QueueDeclare(
		"email_queue", // queue name
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Errorf("Failed to declare a queue: %v", err)
		panic(err)
	}

	return &RabbitMQ{
		Conn:    conn,
		Channel: ch,
		Queue:   q,
	}
}

func (r *RabbitMQ) EnqueueMessage(ctx context.Context, email string) error {
	// start span
	parentSpanID := utils.ExtractParentSpantID(ctx)
	const spanName = "EnqueueMessage"
	opts := []oteltrace.SpanStartOption{
		oteltrace.WithAttributes(attribute.KeyValue{Key: attribute.Key("messaging.system"), Value: attribute.StringValue("rabbitmq")}),
		oteltrace.WithAttributes(attribute.KeyValue{Key: attribute.Key("messaging.operation"), Value: attribute.StringValue("publish")}),
	}

	ctx, span := otel.Tracer(spanName).Start(ctx, spanName, opts...)
	defer span.End()

	log.SetFormatter(&log.JSONFormatter{})
	l := log.WithFields(log.Fields{
		"parent_id": parentSpanID,
		"trace_id":  span.SpanContext().TraceID().String(),
		"span_id":   span.SpanContext().SpanID().String(),
		"email":     email,
	})
	l.Info("Enqueue message !")

	messageBytes, err := json.Marshal(QueueMessage{Email: email})
	if err != nil {
		l.Errorf("Failed to marshal message to JSON: %v", err)
		return err
	}

	body := amqp.Publishing{
		ContentType: "text/plain",
		Body:        messageBytes,
	}

	// propagate trace context
	body.Headers = utils.CtxToHeader(ctx)
	err = r.Channel.Publish(
		"",           // exchange
		r.Queue.Name, // routing key
		false,
		false,
		body)
	if err != nil {
		l.Errorf("Failed to publish message: %v", err)
		return err
	}

	return nil
}

func (r *RabbitMQ) Close() {
	r.Channel.Close()
	r.Conn.Close()
}
