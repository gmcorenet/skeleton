package main

//go:generate go run github.com/gmcorenet/framework/routing/generator internal/controller internal/generated github.com/gmcorenet/skeleton

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gmcorenet/framework/kernel"
	"github.com/gmcorenet/framework/routing"
	_ "github.com/gmcorenet/framework/csrf"
	_ "github.com/gmcorenet/framework/session"
	_ "github.com/gmcorenet/sdk/gmcore-debugbar"
	"github.com/gmcorenet/skeleton/internal/generated"
	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Server struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	} `yaml:"server"`
	App struct {
		Env   string `yaml:"env"`
		Debug bool   `yaml:"debug"`
	} `yaml:"app"`
}

func loadConfig(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := &AppConfig{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func main() {
	cfg := &kernel.Config{
		Host:     "0.0.0.0",
		Port:     "8080",
		Env:      "dev",
		Debug:    false,
		RootPath: getCwd(),
	}

	if appCfg, err := loadConfig("app.yaml"); err == nil {
		if appCfg.Server.Host != "" {
			cfg.Host = appCfg.Server.Host
		}
		if appCfg.Server.Port != "" {
			cfg.Port = appCfg.Server.Port
		}
		if appCfg.App.Env != "" {
			cfg.Env = appCfg.App.Env
		}
		cfg.Debug = appCfg.App.Debug
	} else {
		log.Printf("Could not load app.yaml: %v, using defaults", err)
	}

	cfg.Host = getEnv("SERVER_HOST", cfg.Host)
	cfg.Port = getEnv("SERVER_PORT", cfg.Port)
	cfg.Env = getEnv("APP_ENV", cfg.Env)
	if envDebug := getEnv("APP_DEBUG", ""); envDebug != "" {
		cfg.Debug = envDebug == "true"
	}

	k := kernel.New(cfg)
	k.RegisterDefaultServices()
	k.Bootstrap(context.Background())

	routing.PopulateControllers(k.Container())
	exposed := generated.RegisterGeneratedRoutes(k.RouteBuilder(), k.Container())
	generated.RegisterGeneratedRateLimits(k.Container())
	generated.RegisterGeneratedSecurity(k.Container())
	generated.RegisterGeneratedCaches(k.Container())
	generated.RegisterGeneratedValidators(k.Container())
	routing.InjectAll(k.Container())

	handler := routing.ApplyMiddlewares(k.Container(), k.RouteBuilder(), k)

	if len(exposed) > 0 {
		k.GET("/_routes", func(w http.ResponseWriter, r *http.Request, params map[string]string) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(exposed)
		})
	}

	httpServer := &http.Server{
		Addr:    cfg.Host + ":" + cfg.Port,
		Handler: handler,
	}

	go func() {
		log.Printf("GMCore server on %s", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	httpServer.Shutdown(ctx)
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
