package utils

import (
	"context"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func GenerateUUID() string {
	uuid := uuid.New()
	return uuid.String()
}

func ExtractParentSpantID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	if span == nil {
		return ""
	}
	return span.SpanContext().SpanID().String()
}

func CtxToHeader(ctx context.Context) amqp.Table {
	headers := make(map[string]string)
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(headers))

	bodyHeader := make(amqp.Table)
	for key, value := range headers {
		bodyHeader[key] = value
	}

	return bodyHeader
}
