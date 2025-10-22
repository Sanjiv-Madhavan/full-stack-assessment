package main

import (
	"context"
	"full-stack-assesment/internal/api"
	"full-stack-assesment/internal/middleware"
	"full-stack-assesment/internal/migrate"
	"full-stack-assesment/internal/store"
	"log"
	"log/slog"
	"net/http"
	"os"
)

const (
	address = "0.0.0.0:8080"
)

func init() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
}

func main() {
	ctx := context.Background()

	if err := run(ctx); err != nil {
		panic(err)
	}

}

func run(ctx context.Context) error {
	db, err := store.InMemory(ctx)
	if err != nil {
		log.Fatalf("db init: %v", err)
	}

	if err := migrate.Apply(ctx, db); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	server := api.NewServer(db)
	strictHandler := api.NewStrictHandler(server, nil)
	router := http.NewServeMux()
	h := api.HandlerFromMux(strictHandler, router)

	corsHandler := middleware.CORSMiddleware(h)
	loggingHandler := middleware.LoggingMiddleware(corsHandler)

	s := &http.Server{
		Handler: loggingHandler,
		Addr:    address,
	}

	slog.LogAttrs(ctx, slog.LevelInfo, "Starting server", slog.String("address", address))

	if err := s.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
