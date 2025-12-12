package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/manisprasad/llm-serve/internal/config"
)

func main() {
	cfg := config.MustLoad()

	router := http.NewServeMux()

	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Lamma api is working fine"))
	})

	server := http.Server{
		Addr:    cfg.HttpServer.Address,
		Handler: router,
	}
	fmt.Printf("server is running on: %s", cfg.Address)

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		error := server.ListenAndServe()
		if error != nil {
			log.Fatalf("Failed to start the server %s", error.Error())
		}
	}()

	<-done
	slog.Info("Shutting down the server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Failed to shutdown server gracefully", slog.String("error", err.Error()))
	}

}
