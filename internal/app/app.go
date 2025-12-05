package app

import (
	"log"
	"log/slog"

	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/gin-gonic/gin"

	"github.com/jonx8/pr-review-service/internal/config"
	"github.com/jonx8/pr-review-service/internal/database"
	"github.com/jonx8/pr-review-service/internal/handlers"
	"github.com/jonx8/pr-review-service/internal/repositories"
	"github.com/jonx8/pr-review-service/internal/services"
)

func RunApplication() error {
	cfg := config.Load()

	db, err := database.InitDB(*cfg.DBConfig)
	if err != nil {
		slog.Error("Failed to connect to database:", "error", err)
		return err
	}
	defer db.Close()

	if err := database.RunMigrations(db); err != nil {
		slog.Error("Failed to run migrations", "error", err)
		return err
	}

	slog.Info("Database connected successfully")

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	trManager := manager.Must(trmsqlx.NewDefaultFactory(db))

	teamRepo := repositories.NewTeamRepository(db)
	userRepo := repositories.NewUserRepository(db)
	prRepo := repositories.NewPRRepository(db)

	teamService := services.NewTeamService(teamRepo, trManager)
	userService := services.NewUserService(userRepo, trManager)
	prService := services.NewPRService(prRepo, userService, teamService, trManager)

	router := SetupRouter(teamService, userService, prService)

	log.Printf("Server starting on %s", cfg.ServerAddress)
	log.Printf("Environment: %s", cfg.Environment)

	if err := router.Run(cfg.ServerAddress); err != nil {
		slog.Error("Failed to start server", "error", err)
		return err
	}

	return nil
}

func SetupRouter(teamService services.TeamService, userService services.UserService, prService services.PRService) *gin.Engine {
	router := gin.Default()

	healthHandler := handlers.NewHealthHandler()
	teamHandler := handlers.NewTeamHandler(teamService)
	userHandler := handlers.NewUserHandler(userService, prService)
	prHandler := handlers.NewPRHandler(prService)

	// Health routes
	router.GET("/health", healthHandler.HealthCheck)

	// Team routes
	teamRoutes := router.Group("/team")
	{
		teamRoutes.POST("/add", teamHandler.CreateTeam)
		teamRoutes.GET("/get", teamHandler.GetTeam)
	}

	// User routes
	userRoutes := router.Group("/users")
	{
		userRoutes.POST("/setIsActive", userHandler.SetUserActive)
		userRoutes.GET("/getReview", userHandler.GetUserReviewPRs)
	}

	// PR routes
	prRoutes := router.Group("/pullRequest")
	{
		prRoutes.POST("/create", prHandler.CreatePR)
		prRoutes.POST("/merge", prHandler.MergePR)
		prRoutes.POST("/reassign", prHandler.ReassignReviewer)
	}

	return router
}
