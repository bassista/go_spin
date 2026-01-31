package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"syscall"

	route "github.com/bassista/go_spin/internal/api/route"
	"github.com/bassista/go_spin/internal/cache"
	"github.com/bassista/go_spin/internal/config"
	"github.com/bassista/go_spin/internal/repository"
	"github.com/gin-gonic/gin"

	"github.com/enrichman/httpgrace"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	repo, err := repository.NewJSONRepository(cfg.Data.FilePath)
	if err != nil {
		log.Fatalf("cannot init repository: %v", err)
	}

	jsonDoc, err := repo.Load()
	if err != nil {
		log.Fatalf("cannot load data file: %v", err)
	}

	cacheStore := cache.NewStore(*jsonDoc)

	ctx, stopWatchers := context.WithCancel(context.Background())
	defer stopWatchers()

	if err := repo.StartWatcher(ctx, cacheStore); err != nil {
		log.Fatalf("cannot start watcher: %v", err)
	}

	// Start scheduled persistence goroutine
	cache.StartPersistenceScheduler(ctx, cacheStore, repo, cfg.Data.PersistInterval)

	gin.SetMode(cfg.Security.GinMode)
	r := gin.Default()

	route.SetupRoutes(r, cacheStore)

	srv := createServer(r, cfg)

	fmt.Printf("App will run on port: %s\n", cfg.Server.Port)
	if err := srv.ListenAndServe(":" + cfg.Server.Port); err != nil {
		log.Fatal(err)
	}

}

func createServer(r *gin.Engine, cfg *config.Config) *httpgrace.Server {
	// Set graceful shutdown timeout (default: 10 seconds)
	httpgrace.WithTimeout(cfg.Server.ShutDownTimeout)
	// Customize shutdown signals (default: SIGINT, SIGTERM)
	httpgrace.WithSignals(syscall.SIGTERM, syscall.SIGINT)
	// Provide custom logger (default: slog.Default())
	//httpgrace.WithLogger(customLogger)
	// Provide a function to run before shutdown
	httpgrace.WithBeforeShutdown(func() {
		fmt.Println("Shoutting down!")
	})
	srv := httpgrace.NewServer(r,
		httpgrace.WithServerOptions(
			httpgrace.WithReadTimeout(cfg.Server.ReadTimeout),
			httpgrace.WithWriteTimeout(cfg.Server.WriteTimeout),
			httpgrace.WithIdleTimeout(cfg.Server.IdleTimeout),
			// or with your custom ServerOption
			func(srv *http.Server) {
				srv.ErrorLog = log.New(os.Stdout, "", 0)
			},
		),
	)
	return srv
}
