package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/manisprasad/llm-serve/internal/config"
	"github.com/manisprasad/llm-serve/internal/http/handlers/llm"
	"github.com/rs/cors"
)

func main() {
	cfg := config.MustLoad()

	router := http.NewServeMux()

	// Correct path, no method in HandleFunc
	router.HandleFunc("/api/chat", llm.New(cfg.LlmBaseUrl))

	// Wrap the router with rs/cors
	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // allow all origins
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: false,
	}).Handler(router)

	server := http.Server{
		Addr:    cfg.HttpServer.Address,
		Handler: handler,
	}

	slog.Info("Server is running", slog.String("address", cfg.HttpServer.Address))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %s", err.Error())
		}
	}()

	<-done
	slog.Info("Shutting down the server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Failed to shutdown server gracefully", slog.String("error", err.Error()))
	}

	slog.Info("Server shutdown successfully")
}
