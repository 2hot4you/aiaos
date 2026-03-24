package main

import (
	"fmt"
	"log"

	"github.com/2hot4you/aiaos/backend/internal/handler"
	"github.com/2hot4you/aiaos/backend/internal/repository/postgres"
	"github.com/2hot4you/aiaos/backend/internal/service"
	"github.com/2hot4you/aiaos/backend/pkg/config"
	"github.com/2hot4you/aiaos/backend/pkg/snowflake"
	"github.com/redis/go-redis/v9"
	gormPg "gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Load config
	cfg := config.Load()

	// Init snowflake
	if err := snowflake.Init(1); err != nil {
		log.Fatalf("Failed to init snowflake: %v", err)
	}

	// Connect to PostgreSQL
	db, err := gorm.Open(gormPg.Open(cfg.DB.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(10)

	log.Println("✅ Database connected")

	// Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	log.Println("✅ Redis connected")

	// Repositories
	userRepo := postgres.NewUserRepo(db)
	projectRepo := postgres.NewProjectRepo(db)
	seasonRepo := postgres.NewSeasonRepo(db)
	episodeRepo := postgres.NewEpisodeRepo(db)
	modelRepo := postgres.NewModelConfigRepo(db)

	// Services
	authSvc := service.NewAuthService(userRepo, rdb, cfg.JWT)
	adminSvc := service.NewAdminService(userRepo, modelRepo, cfg.Encryption)
	projectSvc := service.NewProjectService(projectRepo)
	seasonSvc := service.NewSeasonService(seasonRepo, projectRepo)
	episodeSvc := service.NewEpisodeService(episodeRepo, seasonRepo)

	// Router
	r := handler.NewRouter(authSvc, adminSvc, projectSvc, seasonSvc, episodeSvc)

	// Start server
	addr := ":" + cfg.Server.Port
	log.Printf("🚀 AIAOS Backend starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
