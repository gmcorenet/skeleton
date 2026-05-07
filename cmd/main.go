package main

//go:generate go run github.com/gmcorenet/framework/routing/generator internal/controller internal/generated github.com/gmcorenet/skeleton

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gmcorenet/framework/kernel"
	"github.com/gmcorenet/framework/routing"
	_ "github.com/gmcorenet/framework/csrf"
	_ "github.com/gmcorenet/framework/session"
	_ "github.com/gmcorenet/sdk-gmcore-asset-mapper"
	_ "github.com/gmcorenet/sdk-gmcore-debugbar"
	_ "github.com/gmcorenet/sdk-gmcore-filesystem"
	_ "github.com/gmcorenet/sdk-gmcore-form"
	_ "github.com/gmcorenet/sdk-gmcore-transport"

	gmcore_asset_mapper "github.com/gmcorenet/sdk-gmcore-asset-mapper"
	gmcore_cert "github.com/gmcorenet/sdk-gmcore-cert"
	gmcore_error "github.com/gmcorenet/sdk-gmcore-error"
	gmcore_httpclient "github.com/gmcorenet/sdk-gmcore-httpclient"
	gmcore_i18n "github.com/gmcorenet/sdk-gmcore-i18n"
	gmcore_log "github.com/gmcorenet/sdk-gmcore-log"
	gmcore_mailer "github.com/gmcorenet/sdk-gmcore-mailer"
	gmcore_messenger "github.com/gmcorenet/sdk-gmcore-messenger"
	gmcore_migrations "github.com/gmcorenet/sdk-gmcore-migrations"
	gmcore_scheduler "github.com/gmcorenet/sdk-gmcore-scheduler"
	gmcore_serializer "github.com/gmcorenet/sdk-gmcore-serializer"
	gmcore_templating "github.com/gmcorenet/sdk-gmcore-templating"
	gmcore_webhook "github.com/gmcorenet/sdk-gmcore-webhook"

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

	mode := os.Getenv("GMCORE_MODE")
	switch mode {
	case "worker":
		runWorkerMode()
		return
	case "migrate":
		runMigrateMode(false)
		return
	case "migrate-rollback":
		runMigrateMode(true)
		return
	case "scheduler":
		runSchedulerMode()
		return
	case "cert":
		runCertMode()
		return
	}

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

	if err := gmcore_asset_mapper.InstallAssets(filepath.Join(getCwd(), "public")); err != nil {
		logger.Warn("Asset install warning: %v", err)
	}
	http.Handle("/assets/", gmcore_asset_mapper.AssetHandler())

	k.GET("/health", func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "healthy",
			"env":    cfg.Env,
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	routing.PopulateControllers(k.Container())
	exposed := generated.RegisterGeneratedRoutes(k.RouteBuilder(), k.Container())
	generated.RegisterGeneratedRateLimits(k.Container())
	generated.RegisterGeneratedSecurity(k.Container())
	generated.RegisterGeneratedCaches(k.Container())
	generated.RegisterGeneratedValidators(k.Container())
	generated.RegisterGeneratedSubscribers(k.Container())
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
	tmplCfg.Funcs["asset"] = gmcore_asset_mapper.AssetFunc()
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

func runWorkerMode() {
	logger := setupLogger()
	logger.Info("Starting messenger worker")

	appRoot := os.Getenv("GMCORE_APP_ROOT")
	if appRoot == "" {
		appRoot = getCwd()
	}

	transport := gmcore_messenger.NewInMemoryTransport()
	bus := gmcore_messenger.NewBus()

	cfg := &kernel.Config{
		RootPath: appRoot,
		Env:      getEnv("APP_ENV", "dev"),
	}
	k := kernel.New(cfg)
	k.RegisterDefaultServices()

	c := k.Container()
	c.Set("messenger", bus)

	k.Bootstrap(context.Background())

	worker := gmcore_messenger.NewWorker(transport, bus)
	worker.Start()
	logger.Info("Worker started, consuming messages")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down worker")
	worker.Stop()
	logger.Info("Worker stopped")
}

func runMigrateMode(rollback bool) {
	logger := setupLogger()

	appRoot := os.Getenv("GMCORE_APP_ROOT")
	if appRoot == "" {
		appRoot = getCwd()
	}

	manager := gmcore_migrations.NewMigrationManager()
	executor := gmcore_migrations.NewExecutor(manager)

	migrationsDir := filepath.Join(appRoot, "migrations")
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		logger.Error("No migrations directory found: %v", err)
		os.Exit(1)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".sql")
		data, err := os.ReadFile(filepath.Join(migrationsDir, entry.Name()))
		if err != nil {
			logger.Error("Failed to read migration %s: %v", entry.Name(), err)
			continue
		}
		content := string(data)
		upSQL, downSQL := parseMigrationSQL(content)
		m := gmcore_migrations.NewMigrationFile(name, name)
		for _, sql := range upSQL {
			m.AddUp(sql)
		}
		for _, sql := range downSQL {
			m.AddDown(sql)
		}
		manager.RegisterMigration(m)
	}

	if rollback {
		all := manager.GetAllMigrations()
		if len(all) == 0 {
			fmt.Println("No migrations to rollback")
			return
		}
		last := all[len(all)-1]
		fmt.Printf("Rolling back: %s\n", last.GetName())
		if err := executor.ExecuteDown(last); err != nil {
			logger.Error("Rollback failed: %v", err)
			os.Exit(1)
		}
		fmt.Println("Rollback completed")
		return
	}

	pending := manager.GetPendingMigrations()
	if len(pending) == 0 {
		fmt.Println("No pending migrations")
		return
	}

	for _, m := range pending {
		fmt.Printf("Migrating: %s\n", m.GetName())
		if err := executor.ExecuteUp(m); err != nil {
			logger.Error("Migration %s failed: %v", m.GetName(), err)
			os.Exit(1)
		}
	}
	fmt.Printf("Executed %d migration(s)\n", len(pending))
}

func parseMigrationSQL(content string) (upSQL, downSQL []string) {
	lines := strings.Split(content, "\n")
	var current *[]string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "-- UP" {
			current = &upSQL
			continue
		}
		if trimmed == "-- DOWN" {
			current = &downSQL
			continue
		}
		if current != nil && trimmed != "" && !strings.HasPrefix(trimmed, "--") {
			trimmed = strings.TrimSuffix(trimmed, ";")
			*current = append(*current, trimmed)
		}
	}
	return
}

func runSchedulerMode() {
	logger := setupLogger()
	logger.Info("Starting scheduler daemon")

	appRoot := os.Getenv("GMCORE_APP_ROOT")
	if appRoot == "" {
		appRoot = getCwd()
	}

	s := gmcore_scheduler.NewScheduler()

	cfg := &kernel.Config{
		RootPath: appRoot,
		Env:      getEnv("APP_ENV", "dev"),
	}
	k := kernel.New(cfg)
	k.RegisterDefaultServices()

	c := k.Container()
	c.Set("scheduler", s)

	k.Bootstrap(context.Background())

	generated.RegisterGeneratedSubscribers(c)

	s.Start()
	logger.Info("Scheduler started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down scheduler")
	s.Stop()
	logger.Info("Scheduler stopped")
}

func runCertMode() {
	logger := setupLogger()

	appRoot := os.Getenv("GMCORE_APP_ROOT")
	if appRoot == "" {
		appRoot = getCwd()
	}

	action := os.Getenv("GMCORE_CERT_ACTION")
	domain := os.Getenv("CERT_HOSTNAME")
	email := os.Getenv("CERT_EMAIL")

	if domain == "" && (action == "request" || action == "info") {
		logger.Error("CERT_HOSTNAME must be set for cert:%s", action)
		os.Exit(1)
	}

	lifecycle := gmcore_cert.NewCertLifecycle(appRoot)

	switch action {
	case "request":
		if domain == "" || email == "" {
			logger.Error("CERT_HOSTNAME and CERT_EMAIL must be set for cert:request")
			os.Exit(1)
		}
		logger.Info("Requesting Let's Encrypt certificate for %s", domain)
		if err := lifecycle.Request(domain, email); err != nil {
			logger.Error("cert:request failed: %v", err)
			os.Exit(1)
		}
		logger.Info("Certificate obtained successfully for %s", domain)

	case "renew":
		logger.Info("Renewing Let's Encrypt certificate")
		if err := lifecycle.Renew(); err != nil {
			logger.Error("cert:renew failed: %v", err)
			os.Exit(1)
		}
		logger.Info("Certificate renewed successfully")

	case "revoke":
		if domain == "" {
			logger.Error("CERT_HOSTNAME must be set for cert:revoke")
			os.Exit(1)
		}
		logger.Info("Revoking certificate for %s", domain)
		if err := lifecycle.Revoke(domain); err != nil {
			logger.Error("cert:revoke failed: %v", err)
			os.Exit(1)
		}

	case "import":
		certFile := os.Getenv("CERT_IMPORT_CERT")
		keyFile := os.Getenv("CERT_IMPORT_KEY")
		if certFile == "" || keyFile == "" {
			logger.Error("CERT_IMPORT_CERT and CERT_IMPORT_KEY must be set for cert:import")
			os.Exit(1)
		}
		logger.Info("Importing certificate from %s", certFile)
		if err := lifecycle.ImportPEM(certFile, keyFile, domain); err != nil {
			logger.Error("cert:import failed: %v", err)
			os.Exit(1)
		}

	case "export":
		if domain == "" {
			logger.Error("CERT_HOSTNAME must be set for cert:export")
			os.Exit(1)
		}
		certPEM, keyPEM, err := lifecycle.ExportPEM(domain)
		if err != nil {
			logger.Error("cert:export failed: %v", err)
			os.Exit(1)
		}
		fmt.Println(certPEM)
		fmt.Println(keyPEM)

	case "info":
		if domain == "" {
			logger.Error("CERT_HOSTNAME must be set for cert:info")
			os.Exit(1)
		}
		info, err := lifecycle.Info(domain)
		if err != nil {
			logger.Error("cert:info failed: %v", err)
			os.Exit(1)
		}
		gmcore_cert.PrintCertInfo(info)

	default:
		logger.Error("Unknown cert action: %s (use request, renew, revoke, import, export, info)", action)
		fmt.Println("Usage: GMCORE_MODE=cert GMCORE_CERT_ACTION=<action> [CERT_HOSTNAME=domain] [CERT_EMAIL=email] ...")
		fmt.Println()
		fmt.Println("Available actions:")
		fmt.Println("  request  - Request a new Let's Encrypt certificate")
		fmt.Println("  renew    - Renew an existing certificate")
		fmt.Println("  revoke   - Revoke a certificate")
		fmt.Println("  import   - Import certificate from PEM files (set CERT_IMPORT_CERT, CERT_IMPORT_KEY)")
		fmt.Println("  export   - Export certificate to PEM")
		fmt.Println("  info     - Show certificate details")
		os.Exit(1)
	}
}
