package main

import (
	"expenses-backend/internal/expense"
	"expenses-backend/pkg/expense/v1/expensev1connect"
	"net/http"

	"connectrpc.com/grpcreflect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {

	mux := http.NewServeMux()

	e := expense.NewService()
	expenseServicePath, expenseServiceHandler := expensev1connect.NewExpenseServiceHandler(e)

	mux.Handle(expenseServicePath, expenseServiceHandler)

	reflector := grpcreflect.NewStaticReflector(
		"expense.v1.ExpenseService",
	)

	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	http.ListenAndServe(
		"0.0.0.0:8080",
		// Use h2c so we can serve HTTP/2 without TLS.
		h2c.NewHandler(withCORS(mux), &http2.Server{}),
	)
}
