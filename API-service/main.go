package main

import (
	"fmt"
	"os"

	"github.com/datphamcode295/distributed-tracing/handlers"
	"github.com/datphamcode295/distributed-tracing/queue"

	"github.com/getsentry/sentry-go"
	sentryotel "github.com/getsentry/sentry-go/otel"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}

	// Sentry
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

	// Opentelemetry
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(sentryotel.NewSentrySpanProcessor()),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(sentryotel.NewSentryPropagator())

	q := queue.New(os.Getenv("RABBITMQ_URL"))
	defer q.Close()

	h := handlers.New(q)

	r := gin.Default()

	//gin OTEL instrumentation
	r.Use(otelgin.Middleware("api-service"))
	if err != nil {
		fmt.Printf("Failed to create Sentry OTLP exporter: %v\n", err)
		return
	}

	// Apply CORS middleware
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}                                                // Update with your frontend URL
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"} // Allow the required methods
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Baggage", "Sentry-Trace"}
	r.Use(cors.New(config))

	r.POST("/users/reset-password", h.ResetPasswordHandler)

	fmt.Println("Server started on :3000")
	r.Run(":3000")
}
