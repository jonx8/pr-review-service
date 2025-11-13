package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type HealthCheckResponse struct {
	Status      string `json:"status"`
	ServiceName string `json:"serviceName"`
	Version     string `json:"version"`
	Timestamp   int64  `json:"timestamp"`
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, HealthCheckResponse{
		Status:      "OK",
		ServiceName: "PR review service",
		Version:     "0.1.0",
		Timestamp:   time.Now().Unix(),
	})
}
