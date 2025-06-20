package main

import (
	"context"
	"expenses-backend/internal/auth"
	"expenses-backend/internal/database"
	"expenses-backend/internal/database/migrations"
	"expenses-backend/internal/database/sql/masterdb"
	"expenses-backend/internal/database/turso"
	"expenses-backend/internal/expense"
	"expenses-backend/internal/family"
	"expenses-backend/internal/middleware"
	"expenses-backend/pkg/auth/v1/authv1connect"
	"expenses-backend/pkg/expense/v1/expensev1connect"
	"net/http"
	"os"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	godotenv.Load()

	zerolog.TimeFieldFormat = time.RFC3339
	logger := log.With().Timestamp().Caller().Logger()

	dbConfig := database.Config{
		MasterDatabaseURL: os.Getenv("TURSO_MASTER_DB_URL"),
		TursoConfig: turso.Config{
			AuthToken:    os.Getenv("TURSO_AUTH_TOKEN"),
			Organization: os.Getenv("TURSO_ORGANIZATION"),
			MaxRetries:   3,
			RetryDelay:   time.Second,
		},
	}

	ctx := context.Background()
	dbManager, err := database.NewManager(ctx, dbConfig, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize database manager")
	}
	defer dbManager.Close()

	migrationManager := migrations.NewMigrationManager(logger)

	if err := dbManager.RunMigrations(ctx, migrationManager); err != nil {
		logger.Fatal().Err(err).Msg("Failed to run migrations")
	}

	// Initialize database factory
	dbFactory := database.NewFactory(dbManager)

	// Initialize services
	masterDB := dbManager.GetMasterDatabase()
	masterQueries := masterdb.New(masterDB)
	authService := auth.NewService(masterDB, masterQueries, logger)
	familyService := family.NewService(dbManager, logger)
	expenseService := expense.NewService(dbFactory, logger)

	// Initialize middleware
	authInterceptor := middleware.NewAuthInterceptor(authService, dbManager, logger)
	loggingInterceptor := middleware.NewLoggingInterceptor(logger)

	// Create interceptor chain
	interceptors := connect.WithInterceptors(
		loggingInterceptor,
		authInterceptor,
	)

	// Suppress unused variable warnings for now
	_ = familyService

	mux := http.NewServeMux()

	// Register services with interceptors
	expenseServicePath, expenseServiceHandler := expensev1connect.NewExpenseServiceHandler(
		expenseService,
		interceptors,
	)
	mux.Handle(expenseServicePath, expenseServiceHandler)

	authServicePath, authServiceHandler := authv1connect.NewAuthServiceHandler(
		authService,
		connect.WithInterceptors(loggingInterceptor),
	)
	mux.Handle(authServicePath, authServiceHandler)

	reflector := grpcreflect.NewStaticReflector(
		"expense.v1.ExpenseService",
		"auth.v1.AuthService",
	)

	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	if err := http.ListenAndServe(
		":8080",
		// Use h2c so we can serve HTTP/2 without TLS.
		h2c.NewHandler(withCORS(mux), &http2.Server{}),
	); err != nil {
		panic(err)
	}
}
