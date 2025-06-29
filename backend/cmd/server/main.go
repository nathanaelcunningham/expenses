package main

import (
	"context"
	"expenses-backend/internal/auth"
	"expenses-backend/internal/database"
	"expenses-backend/internal/database/migrations"
	"expenses-backend/internal/database/sql/familydb"
	"expenses-backend/internal/database/sql/masterdb"
	"expenses-backend/internal/database/turso"
	"expenses-backend/internal/expense"
	"expenses-backend/internal/family"
	"expenses-backend/internal/middleware"
	"expenses-backend/internal/simplefin"
	"expenses-backend/internal/transaction"
	"expenses-backend/pkg/auth/v1/authv1connect"
	"expenses-backend/pkg/expense/v1/expensev1connect"
	"expenses-backend/pkg/transaction/v1/transactionv1connect"
	"net/http"
	"os"

	"expenses-backend/internal/logger"

	"connectrpc.com/connect"
	"connectrpc.com/grpcreflect"
	"github.com/joho/godotenv"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	// Initialize env
	godotenv.Load()

	log := logger.New(nil)

	tursoClient := turso.NewClient(turso.Config{
		AuthToken:    os.Getenv("TURSO_AUTH_TOKEN"),
		ApiToken:     os.Getenv("TURSO_API_TOKEN"),
		Organization: os.Getenv("TURSO_ORGANIZATION"),
	})

	simplefinClient, err := simplefin.NewClient(os.Getenv("SIMPLEFIN_ACCESS_TOKEN"))
	if err != nil {
		panic(err)
	}

	masterDB, err := tursoClient.Connect(context.Background(), os.Getenv("TURSO_MASTER_DB_URL"))
	if err != nil {
		panic(err)
	}
	familyDB, err := tursoClient.Connect(context.Background(), os.Getenv("TURSO_FAMILY_SEED_URL"))
	if err != nil {
		panic(err)
	}
	defer familyDB.Close()

	familyQueries := familydb.New(familyDB)
	migrationManager := migrations.NewMigrationManager(log, masterdb.New(masterDB), familyQueries)

	dbManager := database.New(masterDB, tursoClient, migrationManager, log)
	defer dbManager.Close()

	err = migrationManager.RunStartupMigrations(context.Background(), masterDB, familyDB)
	if err != nil {
		panic(err)
	}

	// Load existing family databases
	err = dbManager.LoadExistingFamilyDatabases(context.Background())
	if err != nil {
		log.Warn("Failed to load existing family databases", err)
	}

	familyService := family.NewService(dbManager, log)
	authService := auth.NewService(dbManager, familyService, log)
	expenseService := expense.NewService(dbManager, log)
	transactionService := transaction.NewService(dbManager, log, simplefinClient)

	// Initialize middleware
	authInterceptor := middleware.NewAuthInterceptor(authService, dbManager, log)
	loggingInterceptor := middleware.NewLoggingInterceptor(log)

	// Create interceptor chain
	interceptors := connect.WithInterceptors(
		loggingInterceptor,
		authInterceptor,
	)

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

	transactionServicePath, transactionServiceHandler := transactionv1connect.NewTransactionServiceHandler(transactionService, interceptors)
	mux.Handle(transactionServicePath, transactionServiceHandler)

	reflector := grpcreflect.NewStaticReflector(
		"expense.v1.ExpenseService",
		"auth.v1.AuthService",
		"transaction.v1.TransactionService",
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
