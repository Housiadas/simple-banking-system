package main

import (
	"context"
	"expvar"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/Housiadas/backend-system/app/api/handler"
	"github.com/Housiadas/backend-system/app/domain/productapp"
	"github.com/Housiadas/backend-system/app/domain/systemapp"
	"github.com/Housiadas/backend-system/app/domain/userapp"
	"github.com/Housiadas/backend-system/business/config"
	"github.com/Housiadas/backend-system/business/data/sqldb"
	"github.com/Housiadas/backend-system/business/domain/authbus"
	"github.com/Housiadas/backend-system/business/domain/productbus"
	"github.com/Housiadas/backend-system/business/domain/productbus/stores/productdb"
	"github.com/Housiadas/backend-system/business/domain/userbus"
	"github.com/Housiadas/backend-system/business/domain/userbus/stores/userdb"
	"github.com/Housiadas/backend-system/business/web"
	"github.com/Housiadas/backend-system/business/web/mid"
	_ "github.com/Housiadas/backend-system/docs"
	"github.com/Housiadas/backend-system/foundation/debug"
	"github.com/Housiadas/backend-system/foundation/kafka"
	"github.com/Housiadas/backend-system/foundation/keystore"
	"github.com/Housiadas/backend-system/foundation/logger"
	"github.com/Housiadas/backend-system/foundation/tracer"
)

/*
	Need to figure out timeouts for http service.
*/

var build = "develop"

// @title           Backend System
// @description     This is a backend system with various technologies.
//
// @contact.name	API Support
// @contact.url		http://www.swagger.io/support
// @contact.email	support@swagger.io
//
// @license.name	Apache 2.0
// @license.url		http://www.apache.org/licenses/LICENSE-2.0.html
//
//	@query.collection.format multi
//
// @externalDocs.description  OpenAPI
//
// @externalDocs.url	https://swagger.io/resources/open-api/
// @host				localhost:4000
func main() {
	var log *logger.Logger

	events := logger.Events{
		Error: func(ctx context.Context, r logger.Record) {
			log.Info(ctx, "******* SEND ALERT *******")
		},
	}
	traceIDFn := func(ctx context.Context) string {
		return web.GetTraceID(ctx)
	}

	log = logger.NewWithEvents(os.Stdout, logger.LevelInfo, "API", traceIDFn, events)

	// -------------------------------------------------------------------------
	// Run the application
	// -------------------------------------------------------------------------
	ctx := context.Background()
	if err := run(ctx, log); err != nil {
		log.Error(ctx, "startup", "msg", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, log *logger.Logger) error {

	// -------------------------------------------------------------------------
	// GOMAXPROCS
	// -------------------------------------------------------------------------
	log.Info(ctx, "startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	// -------------------------------------------------------------------------
	// Configuration
	// -------------------------------------------------------------------------
	cfg, err := config.LoadConfig("../../")
	if err != nil {
		return fmt.Errorf("parsing config: %w", err)
	}
	cfg.Version = config.Version{
		Build: build,
		Desc:  "API",
	}

	// -------------------------------------------------------------------------
	// App Starting
	// -------------------------------------------------------------------------
	log.Info(ctx, "starting service", "version", cfg.Version.Build)
	defer log.Info(ctx, "shutdown complete")
	expvar.NewString("build").Set(cfg.Version.Build)

	// -------------------------------------------------------------------------
	// Database Support
	// -------------------------------------------------------------------------
	log.Info(ctx, "startup", "status", "initializing database support", "hostport", cfg.DB.Host)

	db, err := sqldb.Open(sqldb.Config{
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Host:         cfg.DB.Host,
		Name:         cfg.DB.Name,
		MaxIdleConns: cfg.DB.MaxIdleConns,
		MaxOpenConns: cfg.DB.MaxOpenConns,
		DisableTLS:   cfg.DB.DisableTLS,
	})
	if err != nil {
		return fmt.Errorf("connecting to db: %w", err)
	}
	defer db.Close()

	// -------------------------------------------------------------------------
	// Initialize authentication support
	// -------------------------------------------------------------------------
	log.Info(ctx, "startup", "status", "initializing authentication support")

	// Load the private keys files from disk. We can assume some systemapi like
	// Vault has created these files already. How that happens is not our concern.
	ks := keystore.New()
	if err := ks.LoadRSAKeys(os.DirFS(cfg.Auth.KeysFolder)); err != nil {
		return fmt.Errorf("reading keys: %w", err)
	}

	authCfg := authbus.Config{
		Log:       log,
		DB:        db,
		KeyLookup: ks,
	}

	authSrv, err := authbus.New(authCfg)
	if err != nil {
		return fmt.Errorf("constructing authapi: %w", err)
	}

	// -------------------------------------------------------------------------
	// Initialize Kafka Producer
	// -------------------------------------------------------------------------
	log.Info(ctx, "startup", "status", "initializing kafka support")
	producer, err := kafka.NewProducer(kafka.ProducerConfig{
		Brokers:          cfg.Kafka.Brokers,
		LogLevel:         cfg.Kafka.LogLevel,
		AddressFamily:    cfg.Kafka.AddressFamily,
		MaxMessageBytes:  cfg.Kafka.MaxMessageBytes,
		SecurityProtocol: cfg.Kafka.SecurityProtocol,
	})
	if err != nil {
		return fmt.Errorf("creating kafka producer: %w", err)
	}
	defer producer.Close()

	// -------------------------------------------------------------------------
	// Start Tracing Support
	// -------------------------------------------------------------------------
	log.Info(ctx, "startup", "status", "initializing tracing support")

	traceProvider, err := tracer.InitTracing(tracer.Config{
		Log:         log,
		ServiceName: cfg.App.Name,
		Host:        cfg.Tempo.Host,
		ExcludedRoutes: map[string]struct{}{
			"/v1/liveness":  {},
			"/v1/readiness": {},
		},
		Probability: cfg.Tempo.Probability,
	})
	if err != nil {
		return fmt.Errorf("starting tracing: %w", err)
	}
	defer traceProvider.Shutdown(context.Background())

	tr := traceProvider.Tracer(cfg.App.Name)

	// -------------------------------------------------------------------------
	// Build Business APIs
	// -------------------------------------------------------------------------
	log.Info(ctx, "startup", "status", "initializing business core")

	userBus := userbus.NewBusiness(log, userdb.NewStore(log, db))
	productBus := productbus.NewBusiness(log, userBus, productdb.NewStore(log, db))

	// -------------------------------------------------------------------------
	// Start Debug Service
	// -------------------------------------------------------------------------
	go func() {
		log.Info(ctx, "startup", "status", "debug v1 router started", "host", cfg.Server.Debug)

		if err := http.ListenAndServe(cfg.Server.Debug, debug.Mux()); err != nil {
			log.Error(ctx, "shutdown", "status", "debug v1 router closed", "host", cfg.Server.Debug, "msg", err)
		}
	}()

	// -------------------------------------------------------------------------
	// Start Http Server
	// -------------------------------------------------------------------------
	log.Info(ctx, "startup", "status", "initializing Http Server")

	// Initialize handler
	respond := web.NewRespond(log)
	midBusiness := mid.Business{
		Auth:    authSrv,
		User:    userBus,
		Product: productBus,
	}
	h := handler.Handler{
		AppName: cfg.App.Name,
		Log:     log,
		DB:      db,
		Tracer:  tr,
		Build:   build,
		Cors:    cfg.Cors,
		Web: handler.Web{
			Mid:     mid.New(midBusiness, log),
			Respond: respond,
		},
		App: handler.App{
			User:    userapp.NewApp(userBus, authSrv),
			Product: productapp.NewApp(productBus),
			System:  systemapp.NewApp(cfg.Version.Build, log, db),
		},
		Business: handler.Business{
			Auth:    authSrv,
			User:    userBus,
			Product: productBus,
		},
	}

	api := http.Server{
		Addr:         cfg.Server.Api,
		Handler:      h.Routes(),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
		ErrorLog:     logger.NewStdLogger(log, logger.LevelError),
	}

	serverErrors := make(chan error, 1)
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Info(ctx, "startup", "status", "api router started", "host", api.Addr)

		serverErrors <- api.ListenAndServe()
	}()

	// Shutdown
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Info(ctx, "shutdown", "status", "shutdown started", "signal", sig)
		defer log.Info(ctx, "shutdown", "status", "shutdown complete", "signal", sig)

		ctx, cancel := context.WithTimeout(ctx, cfg.Server.ShutdownTimeout)
		defer cancel()

		if err := api.Shutdown(ctx); err != nil {
			api.Close()
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}
