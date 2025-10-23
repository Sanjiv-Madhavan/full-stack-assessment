package main

import (
	"context"
	"full-stack-assesment/internal/api"
	"full-stack-assesment/internal/middleware"
	"full-stack-assesment/internal/migrate"
	projectsRepo "full-stack-assesment/internal/repo/projects"
	tasksRepo "full-stack-assesment/internal/repo/task"
	projectsService "full-stack-assesment/internal/service/projects"
	taskService "full-stack-assesment/internal/service/task"

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

	projectsRepo := projectsRepo.NewSQLiteProjectsRepo(db)
	taskRepo := tasksRepo.NewSQLiteTaskRepo(db)

	projectsService := projectsService.NewService(*projectsRepo)
	tasksService := taskService.NewService(*taskRepo, *projectsService)

	server := api.NewServer(*projectsService, *tasksService)
	router := http.NewServeMux()
	h := api.HandlerFromMux(server, router)

	handler := middleware.RecoverMiddleware(
		middleware.LoggingMiddleware(
			middleware.CORSMiddleware(h),
		),
	)

	s := &http.Server{
		Handler: handler,
		Addr:    address,
	}

	slog.LogAttrs(ctx, slog.LevelInfo, "Starting server", slog.String("address", address))

	if err := s.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
