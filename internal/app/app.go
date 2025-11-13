package app

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jonx8/pr-review-service/internal/config"
	"github.com/jonx8/pr-review-service/internal/database"
	"github.com/jonx8/pr-review-service/internal/handlers"
)

func RunApplication() {
	cfg := config.Load()

	db, err := database.InitDB(*cfg.DBConfig)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	log.Println("Database connected successfully")

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.GET("/health", handlers.HealthCheck)

	log.Printf("Server starting on %s", cfg.ServerAddress)
	log.Printf("Environment: %s", cfg.Environment)

	if err := r.Run(cfg.ServerAddress); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
