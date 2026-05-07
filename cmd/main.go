package main

//go:generate go run github.com/gmcorenet/framework/routing/generator internal/controller internal/generated github.com/gmcorenet/skeleton

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gmcorenet/framework/kernel"
	"github.com/gmcorenet/framework/routing"
	_ "github.com/gmcorenet/framework/csrf"
	_ "github.com/gmcorenet/framework/session"
	_ "github.com/gmcorenet/sdk/gmcore-debugbar"
	_ "github.com/gmcorenet/sdk/gmcore-filesystem"
	_ "github.com/gmcorenet/sdk/gmcore-form"
	_ "github.com/gmcorenet/sdk/gmcore-transport"

	gmcore_error "github.com/gmcorenet/sdk/gmcore-error"
	gmcore_httpclient "github.com/gmcorenet/sdk/gmcore-httpclient"
	gmcore_i18n "github.com/gmcorenet/sdk/gmcore-i18n"
	gmcore_log "github.com/gmcorenet/sdk/gmcore-log"
	gmcore_mailer "github.com/gmcorenet/sdk/gmcore-mailer"
	gmcore_messenger "github.com/gmcorenet/sdk/gmcore-messenger"
	gmcore_scheduler "github.com/gmcorenet/sdk/gmcore-scheduler"
	gmcore_serializer "github.com/gmcorenet/sdk/gmcore-serializer"
	gmcore_templating "github.com/gmcorenet/sdk/gmcore-templating"
	gmcore_webhook "github.com/gmcorenet/sdk/gmcore-webhook"

	"github.com/gmcorenet/skeleton/internal/generated"
	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Type        string `yaml:"type"`
	Description string `yaml:"description"`
	Server      struct {
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
	defer gmcore_error.Recover()

	logger := setupLogger()
	logger.Info("GMCore Skeleton starting")

	cfg := &kernel.Config{
		Host:     "0.0.0.0",
		Port:     "8080",
		Env:      "dev",
		Debug:    false,
		RootPath: getCwd(),
	}

	if appCfg, err := loadConfig("manifest.yaml"); err == nil {
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
		if appCfg.Name != "" {
			logger.WithField("app", appCfg.Name).WithField("version", appCfg.Version).Info("Loaded manifest")
		}
	} else {
		logger.Warn("Could not load manifest.yaml, using defaults")
	}

	cfg.Host = getEnv("SERVER_HOST", cfg.Host)
	cfg.Port = getEnv("SERVER_PORT", cfg.Port)
	cfg.Env = getEnv("APP_ENV", cfg.Env)
	if envDebug := getEnv("APP_DEBUG", ""); envDebug != "" {
		cfg.Debug = envDebug == "true"
	}

	k := kernel.New(cfg)
	k.RegisterDefaultServices()

	registerSDKServices(k, logger, cfg)

	k.Bootstrap(context.Background())

	routing.PopulateControllers(k.Container())
	exposed := generated.RegisterGeneratedRoutes(k.RouteBuilder(), k.Container())
	generated.RegisterGeneratedRateLimits(k.Container())
	generated.RegisterGeneratedSecurity(k.Container())
	generated.RegisterGeneratedCaches(k.Container())
	generated.RegisterGeneratedValidators(k.Container())
	routing.InjectAll(k.Container())

	handler := routing.ApplyMiddlewares(k.Container(), k.RouteBuilder(), k)

	setupHomeRoute(k, logger)

	k.Container().Set("serializer", gmcore_serializer.NewSerializer())
	k.Container().Set("json_serializer", gmcore_serializer.NewJSONSerializer())

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
		logger.WithField("addr", httpServer.Addr).Info("GMCore server listening")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	httpServer.Shutdown(ctx)
	k.Shutdown()
	logger.Info("Server stopped")
}

func setupLogger() *gmcore_log.Logger {
	logger := gmcore_log.New()
	logger.AddHandler(gmcore_log.NewConsoleHandler(os.Stdout))
	logger.SetLevel(gmcore_log.LevelInfo)

	logCfg, err := gmcore_log.LoadConfig("config")
	if err == nil && logCfg != nil {
		if built, err := logCfg.Build(); err == nil {
			return built
		}
	}
	return logger
}

func registerSDKServices(k *kernel.Kernel, logger *gmcore_log.Logger, cfg *kernel.Config) {
	c := k.Container()
	count := 0

	mailerCfg, _ := gmcore_mailer.LoadConfig("config")
	if mailerCfg != nil && mailerCfg.Host != "" {
		c.Set("mailer", gmcore_mailer.NewSMTPMailer(mailerCfg.Host, mailerCfg.Port, mailerCfg.Username, mailerCfg.Password))
	} else {
		c.Set("mailer", gmcore_mailer.NewMemoryMailer())
	}
	count++

	tmplCfg := gmcore_templating.Config{
		AppRoot:      getCwd(),
		SystemRoot:   "",
		BundleRoots:  nil,
		Mode:         "dev",
		DisableCache: true,
		Funcs:        gmcore_templating.GetFuncs(),
	}
	if cfg.Env == "prod" {
		tmplCfg.Mode = "prod"
		tmplCfg.DisableCache = false
	}
	c.Set("templating", gmcore_templating.New(tmplCfg))
	count++

	c.Set("messenger", gmcore_messenger.NewBus())
	count++

	c.Set("scheduler", gmcore_scheduler.NewScheduler())
	count++

	c.Set("webhook_manager", gmcore_webhook.NewManager())
	count++

	c.Set("http_client", gmcore_httpclient.NewClient())
	count++

	i18nCfg, _ := gmcore_i18n.LoadConfig("config")
	if i18nCfg != nil {
		if translator, err := i18nCfg.Build(); err == nil {
			c.Set("translator", translator)
			count++
		}
	}
	if _, err := c.Get("translator"); err != nil {
		translationsDir := filepath.Join(getCwd(), "resources", "translations")
		if translator, err := gmcore_i18n.LoadDir(translationsDir, "en"); err == nil {
			c.Set("translator", translator)
			count++
		}
	}

	logger.WithField("count", count).Info("SDK services registered")
}

func setupHomeRoute(k *kernel.Kernel, logger *gmcore_log.Logger) {
	tmpl, err := k.Container().Get("templating")
	if err != nil {
		return
	}
	engine, ok := tmpl.(*gmcore_templating.Engine)
	if !ok {
		return
	}

	k.GET("/", func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		data := map[string]interface{}{
			"title":   "GMCore Application",
			"message": "Welcome to GMCore Framework — a Symfony-like foundation for Go",
			"env":     k.Config().Env,
			"sdks": []string{
				"gmcore-log", "gmcore-templating", "gmcore-mailer",
				"gmcore-serializer", "gmcore-form", "gmcore-filesystem",
				"gmcore-httpclient", "gmcore-messenger", "gmcore-scheduler",
				"gmcore-webhook", "gmcore-i18n", "gmcore-transport",
				"gmcore-error", "gmcore-debugbar",
			},
		}
		rendered, err := engine.RenderContext(r.Context(), "home.html", data)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "ok",
				"message": "GMCore Framework is running",
				"env":     k.Config().Env,
			})
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(rendered))
	})

	logger.Info("Home route registered")
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
	return filepath.ToSlash(dir)
}
