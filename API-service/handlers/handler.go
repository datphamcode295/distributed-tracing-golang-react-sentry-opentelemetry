package handlers

import (
	"net/http"

	"github.com/datphamcode295/distributed-tracing/models"
	"github.com/datphamcode295/distributed-tracing/queue"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

type Handler interface {
	ResetPasswordHandler(c *gin.Context)
}

type handler struct {
	queue queue.Queue
}

func New(q queue.Queue) Handler {
	return handler{
		queue: q,
	}
}

func (h handler) ResetPasswordHandler(c *gin.Context) {
	const spanName = "resetPasswordHandler"
	newCtx, span := otel.Tracer(spanName).Start(c.Request.Context(), spanName)
	defer span.End()

	// Update c ctx to newCtx
	c.Request = c.Request.WithContext(newCtx)

	traceID := c.GetHeader("Sentry-Trace")
	log.SetFormatter(&log.JSONFormatter{})
	l := log.WithFields(log.Fields{
		"trace_id": traceID,
		"span_id":  span.SpanContext().SpanID().String(),
	})
	l.Info("Receive reset password request !")

	var request models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		l.Error("Error parsing request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
		return
	}

	err := h.queue.EnqueueMessage(c.Request.Context(), request.Email)
	if err != nil {
		l.Error("Error enqueue message")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{
		Message: "Password reset email sent",
	})
}
