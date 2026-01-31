package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
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
	rt, err := runtime.NewRuntimeFromConfig(cfg.Misc.RuntimeType, jsonDoc)
	if err != nil {
		log.Fatalf("cannot init runtime: %v", err)
	}

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
	if err := srv.ListenAndServe(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}

}

func createServer(r *gin.Engine, app *appctx.App) *httpgrace.Server {
	srv := httpgrace.NewServer(r,
		httpgrace.WithTimeout(app.Config.Server.ShutDownTimeout),
		httpgrace.WithSignals(syscall.SIGTERM, syscall.SIGINT),
		httpgrace.WithBeforeShutdown(func() {
			fmt.Println("Shutting down....")
		}),
		httpgrace.WithServerOptions(
			httpgrace.WithReadTimeout(app.Config.Server.ReadTimeout),
			httpgrace.WithWriteTimeout(app.Config.Server.WriteTimeout),
			httpgrace.WithIdleTimeout(app.Config.Server.IdleTimeout),
			func(srv *http.Server) {
				srv.BaseContext = func(_ net.Listener) context.Context {
					return app.BaseCtx
				}
			},
			func(srv *http.Server) {
				srv.ErrorLog = log.New(os.Stdout, "", 0)
			},
		),
	)
	return srv
}
