package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gmcORE/framework/kernel"
	"github.com/gmcORE/framework/container"
	"github.com/gmcORE/framework/router"
)

func main() {
	ctx := context.Background()

	cfg := &kernel.Config{
		Host:     getEnv("SERVER_HOST", "0.0.0.0"),
		Port:     getEnv("SERVER_PORT", "8080"),
		Env:      getEnv("APP_ENV", "dev"),
		Debug:    getEnv("APP_DEBUG", "false") == "true",
		RootPath: getCwd(),
	}

	k := kernel.New(cfg)

	container := container.New()
	k.SetContainer(container)

	k.RegisterDefaultServices()

	if err := k.Bootstrap(ctx); err != nil {
		log.Fatalf("Failed to bootstrap kernel: %v", err)
	}

	r := router.New()
	k.SetRouter(r)

	setupRoutes(r)

	httpServer := &http.Server{
		Addr:    cfg.Host + ":" + cfg.Port,
		Handler: k,
	}

	go func() {
		log.Printf("Starting GMCore server on %s", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	waitForShutdown(httpServer, k)
}

func waitForShutdown(server *http.Server, k *kernel.Kernel) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server shutdown error: %v", err)
	}

	k.Shutdown()

	log.Println("Server stopped")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getCwd() string {
	dir, err := os.Getwd()
	if err != nil {
		return "."
	}
	return dir
}

func setupRoutes(r *router.Router) {
	r.GET("/", func(w http.ResponseWriter, req *http.Request, params map[string]string) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("GMCore Framework"))
	})

	r.GET("/health", func(w http.ResponseWriter, req *http.Request, params map[string]string) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
}
