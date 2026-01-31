package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"syscall"

	route "github.com/bassista/go_spin/internal/api/route"
	appctx "github.com/bassista/go_spin/internal/app"
	"github.com/bassista/go_spin/internal/cache"
	"github.com/bassista/go_spin/internal/config"
	"github.com/bassista/go_spin/internal/repository"
	"github.com/bassista/go_spin/internal/runtime"
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
	rt := runtime.NewMemoryRuntimeFromDocument(*jsonDoc)

	app, err := appctx.New(cfg, repo, cacheStore, rt)
	if err != nil {
		log.Fatalf("cannot init app: %v", err)
	}
	defer app.Shutdown()

	app.StartWatchers()

	gin.SetMode(cfg.Misc.GinMode)
	r := gin.Default()

	route.SetupRoutes(r, app)
	srv := createServer(r, app)

	fmt.Printf("App will run on port: %d\n", cfg.Server.Port)
	if err := srv.ListenAndServe(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
		log.Fatal(err)
	}

}

func createServer(r *gin.Engine, app *appctx.App) *httpgrace.Server {
	// Set graceful shutdown timeout (default: 10 seconds)
	httpgrace.WithTimeout(app.Config.Server.ShutDownTimeout)
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
			httpgrace.WithReadTimeout(app.Config.Server.ReadTimeout),
			httpgrace.WithWriteTimeout(app.Config.Server.WriteTimeout),
			httpgrace.WithIdleTimeout(app.Config.Server.IdleTimeout),
			// or with your custom ServerOption
			func(srv *http.Server) {
				srv.ErrorLog = log.New(os.Stdout, "", 0)
			},
		),
	)
	return srv
}
