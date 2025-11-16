package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthCheckResponse struct {
	Status      string    `json:"status"`
	ServiceName string    `json:"serviceName"`
	Version     string    `json:"version"`
	Timestamp   time.Time `json:"timestamp"`
}

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, HealthCheckResponse{
		Status:      "OK",
		ServiceName: "PR review service",
		Version:     "1.0.0",
		Timestamp:   time.Now(),
	})
}
