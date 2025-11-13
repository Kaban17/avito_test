package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"reviewer-service/internal/domain/service"
	"reviewer-service/internal/repository/postgres"
	"reviewer-service/internal/usecase"

	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Setup logger
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// Config
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbName := getEnv("DB_NAME", "reviewer_service")
	dbUser := getEnv("DB_USER", "postgres")
	dbPass := getEnv("DB_PASSWORD", "postgres")
	//	port := getEnv("PORT", "8080")

	// Connect to DB
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPass, dbName)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to open database")
	}
	defer db.Close()

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatal().Err(err).Msg("failed to ping database")
	}

	log.Info().Msg("connected to database")

	// Initialize layers
	txManager := postgres.NewTxManager(db)
	selector := service.NewReviewerSelector()

	_ = usecase.NewTeamUseCase(txManager)
	_ = usecase.NewUserUseCase(txManager)
	_ = usecase.NewPullRequestUseCase(txManager, selector)

	//	teamHandler := handler.NewTeamHandler(teamUC)
	//	userHandler := handler.NewUserHandler(userUC)
	//	prHandler := handler.NewPullRequestHandler(prUC)

	//	router := httphandler.NewRouter(teamHandler, userHandler, prHandler)

	// HTTP Server
	//	srv := &http.Server{
	//		Addr:         ":" + port,
	//		Handler:      router.Setup(),
	//		ReadTimeout:  15 * time.Second,
	//		WriteTimeout: 15 * time.Second,
	//		IdleTimeout:  60 * time.Second,
	//	}
	//
	//	// Start server
	//	go func() {
	//		log.Info().Str("port", port).Msg("starting http server")
	//		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
	//			log.Fatal().Err(err).Msg("server failed")
	//		}
	//	}()
	//
	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down server...")

	_, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	//	if err := srv.Shutdown(ctx); err != nil {
	//		log.Fatal().Err(err).Msg("server forced to shutdown")
	//	}

	log.Info().Msg("server exited")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
