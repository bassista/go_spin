package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"syscall"

	"github.com/bassista/go_spin/internal/api/controller"
	route "github.com/bassista/go_spin/internal/api/route"
	appctx "github.com/bassista/go_spin/internal/app"
	"github.com/bassista/go_spin/internal/cache"
	"github.com/bassista/go_spin/internal/config"
	"github.com/bassista/go_spin/internal/logger"
	"github.com/bassista/go_spin/internal/repository"
	"github.com/bassista/go_spin/internal/runtime"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/enrichman/httpgrace"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.WithComponent("main").Fatalf("configuration error: %v", err)
	}

	// Set log level from configuration
	logLevel, err := logrus.ParseLevel(cfg.Misc.LogLevel)
	if err != nil {
		logger.WithComponent("main").Warnf("invalid log level '%s', using 'info': %v", cfg.Misc.LogLevel, err)
		logLevel = logrus.InfoLevel
	}
	logger.Logger.SetLevel(logLevel)
	logger.WithComponent("main").Debugf("log level set to: %s", logLevel.String())

	repo, err := repository.NewJSONRepository(cfg.Data.FilePath)
	if err != nil {
		logger.WithComponent("main").Fatalf("cannot init repository: %v", err)
	}

	jsonDoc, err := repo.Load(context.Background())
	if err != nil {
		logger.WithComponent("main").Fatalf("cannot load data file: %v", err)
	}

	cacheStore := cache.NewStore(*jsonDoc)
	rt, err := runtime.NewRuntimeFromConfig(cfg.Misc.RuntimeType, jsonDoc)
	if err != nil {
		logger.WithComponent("main").Fatalf("cannot init runtime: %v", err)
	}

	app, err := appctx.New(cfg, repo, cacheStore, rt)
	if err != nil {
		logger.WithComponent("main").Fatalf("cannot init app: %v", err)
	}
	defer app.Shutdown()

	app.StartWatchers()

	gin.SetMode(cfg.Misc.GinMode)
	gin.DefaultWriter = logger.Logger.Writer()
	gin.DefaultErrorWriter = logger.Logger.Writer()
	r := gin.Default()

	route.SetupRoutes(r, app)
	srv := createServer(r, app)

	// Setup and start the secondary waiting server
	waitingSrv := createWaitingServer(app)
	go func() {
		logger.WithComponent("main").Infof("Waiting server will run on port: %d", cfg.Server.WaitingServerPort)
		if err := waitingSrv.ListenAndServe(fmt.Sprintf(":%d", cfg.Server.WaitingServerPort)); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.WithComponent("main").Errorf("Waiting server error: %v", err)
		}
	}()

	logger.WithComponent("main").Infof("App will run on port: %d", cfg.Server.Port)
	if err := srv.ListenAndServe(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.WithComponent("main").Fatal(err)
	}

}

func createServer(r *gin.Engine, app *appctx.App) *httpgrace.Server {
	// Create slog logger that delegates to logrus
	slogLogger := slog.New(slog.NewTextHandler(logger.Logger.Writer(), nil))

	srv := httpgrace.NewServer(r,
		httpgrace.WithTimeout(app.Config.Server.ShutDownTimeout),
		httpgrace.WithSignals(syscall.SIGTERM, syscall.SIGINT),
		httpgrace.WithLogger(slogLogger),
		httpgrace.WithBeforeShutdown(func() {
			logger.WithComponent("http").Info("Shutting down http server....")
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
				srv.ErrorLog = log.New(logger.Logger.Writer(), "[http] ", log.LstdFlags)
			},
		),
	)
	return srv
}

// createWaitingServer creates a secondary HTTP server dedicated to serving only the waiting page.
// It exposes a single route GET /:name that triggers RuntimeController.WaitingPage.
func createWaitingServer(app *appctx.App) *httpgrace.Server {
	r := gin.New()
	r.Use(gin.Recovery())

	// Create RuntimeController for the waiting page
	rc := controller.NewRuntimeController(app.BaseCtx, app.Runtime, app.Cache)
	cc := controller.NewContainerController(app.BaseCtx, app.Cache, app.Runtime)

	// Middleware fallback: handle /container/:name/ready explicitly to ensure JSON response
	r.Use(func(c *gin.Context) {
		// If the path matches /container/:name/ready, call the Ready handler directly and abort
		if strings.HasPrefix(c.Request.URL.Path, "/container/") && strings.HasSuffix(c.Request.URL.Path, "/ready") {
			cc.Ready(c)
			c.Abort()
			return
		}
		c.Next()
	})

	r.GET("/:name", rc.WaitingPage)

	slogLogger := slog.New(slog.NewTextHandler(logger.Logger.Writer(), nil))

	srv := httpgrace.NewServer(r,
		httpgrace.WithTimeout(app.Config.Server.ShutDownTimeout),
		httpgrace.WithSignals(syscall.SIGTERM, syscall.SIGINT),
		httpgrace.WithLogger(slogLogger),
		httpgrace.WithBeforeShutdown(func() {
			logger.WithComponent("http").Info("Shutting down waiting server....")
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
				srv.ErrorLog = log.New(logger.Logger.Writer(), "[waiting-http] ", log.LstdFlags)
			},
		),
	)
	return srv
}
