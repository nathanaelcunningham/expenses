package main

import (
	"context"
	"expenses-backend/internal/auth"
	"expenses-backend/internal/database"
	"expenses-backend/internal/database/migrations"
	"expenses-backend/internal/database/sql/familydb"
	"expenses-backend/internal/database/turso"
	"expenses-backend/internal/expense"
	"expenses-backend/internal/family"
	"expenses-backend/internal/middleware"
	"expenses-backend/pkg/auth/v1/authv1connect"
	"expenses-backend/pkg/expense/v1/expensev1connect"
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

	// migrationManager := migrations.NewMigrationManager(log)
	//
	// // Check/run migrations on master
	// log.Info("Running migrations on master database")
	// if err := migrationManager.RunMigrations(ctx, dbManager.GetMasterDatabase(), migrations.MasterMigration); err != nil {
	// 	log.Fatal("Failed to run master migrations", err)
	// }
	//
	// // Check/run migrations on family-seed
	// log.Info("Running migrations on family-seed database")
	// familySeedMigrated, err := dbManager.RunFamilySeedMigrations(ctx, migrationManager)
	// if err != nil {
	// 	log.Fatal("Failed to run family-seed migrations", err)
	// }
	//
	// // If migrations run on family-seed, run migrations on all family databases
	// if familySeedMigrated {
	// 	log.Info("Family-seed migrations applied, updating all family databases")
	// 	if err := dbManager.RunFamilyDatabaseMigrations(ctx, migrationManager); err != nil {
	// 		log.Fatal("Failed to run migrations on family databases", err)
	// 	}
	// }
	//
	// // Initialize database factory
	// dbFactory := database.NewFactory(dbManager)
	//
	// // Initialize services
	// masterDB := dbManager.GetMasterDatabase()
	// masterQueries := masterdb.New(masterDB)

	tursoClient := turso.NewClient(turso.Config{
		AuthToken:    os.Getenv("TURSO_AUTH_TOKEN"),
		ApiToken:     os.Getenv("TURSO_API_TOKEN"),
		Organization: os.Getenv("TURSO_ORGANIZATION"),
	})

	masterDB, err := tursoClient.Connect(context.Background(), os.Getenv("TURSO_MASTER_DB_URL"))
	if err != nil {
		panic(err)
	}
	familyDB, err := tursoClient.Connect(context.Background(), os.Getenv("TURSO_FAMILY_SEED_URL"))
	if err != nil {
		panic(err)
	}
	defer familyDB.Close()

	dbManager := database.New(masterDB, tursoClient, log)
	defer dbManager.Close()

	familyQueries := familydb.New(familyDB)

	migrationManager := migrations.NewMigrationManager(log, dbManager.GetMasterQueries(), familyQueries)

	err = migrationManager.RunStartupMigrations(context.Background(), masterDB, familyDB)
	if err != nil {
		panic(err)
	}

	familyService := family.NewService(dbManager, log)
	authService := auth.NewService(dbManager, familyService, log)
	expenseService := expense.NewService(dbManager, log)

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
