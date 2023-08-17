package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/datphamcode295/distributed-tracing/models"
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

	r.POST("/users/reset-password", resetPasswordHandler)

	fmt.Println("Server started on :3000")
	r.Run(":3000")
}

func resetPasswordHandler(c *gin.Context) {
	const spanID = "resetPasswordHandler"
	newCtx, span := otel.Tracer(spanID).Start(c.Request.Context(), spanID)
	defer span.End()

	// Update c ctx to newCtx
	c.Request = c.Request.WithContext(newCtx)

	traceID := c.GetHeader("Sentry-Trace")
	log.SetFormatter(&log.JSONFormatter{})
	l := log.WithFields(log.Fields{
		"trace_id": traceID,
		"span_id":  spanID,
	})
	l.Info("Receive reset password request !")

	var request models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		l.Error("Error parsing request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
		return
	}

	queueMessage := queue.QueueMessage{
		TraceID:       traceID,
		ParentTraceID: spanID,
		Email:         request.Email,
	}

	err := queue.EnqueueMessage(c.Request.Context(), queueMessage)
	if err != nil {
		l.Error("Error enqueue message")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{
		Message: "Password reset email sent",
	})
}
