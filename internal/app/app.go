package app

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/bassista/go_spin/internal/cache"
	"github.com/bassista/go_spin/internal/config"
	"github.com/bassista/go_spin/internal/repository"
	"github.com/bassista/go_spin/internal/runtime"
	"github.com/bassista/go_spin/internal/scheduler"
)

// App is the application container (immutable dependencies + lifecycle context).
// It is not a request context; handlers should still use gin's request context.
type App struct {
	Config  *config.Config
	Repo    repository.Repository
	Cache   cache.AppStore
	Runtime runtime.ContainerRuntime

	BaseCtx context.Context
	Cancel  context.CancelFunc
}

func New(cfg *config.Config, repo repository.Repository, store cache.AppStore, rt runtime.ContainerRuntime) (*App, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}
	if repo == nil {
		return nil, errors.New("repo is nil")
	}
	if store == nil {
		return nil, errors.New("cache store is nil")
	}
	if rt == nil {
		return nil, errors.New("runtime is nil")
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &App{
		Config:  cfg,
		Repo:    repo,
		Cache:   store,
		Runtime: rt,
		BaseCtx: ctx,
		Cancel:  cancel,
	}, nil
}

func (a *App) Shutdown() {
	if a == nil || a.Cancel == nil {
		return
	}
	a.Cancel()
}

func (a *App) StartWatchers() {
	if err := a.Repo.StartWatcher(a.BaseCtx, a.Cache); err != nil {
		log.Fatalf("cannot start config file watcher: %v", err)
	}

	// Start scheduled persistence goroutine
	cache.StartPersistenceScheduler(a.BaseCtx, a.Cache, a.Repo, a.Config.Data.PersistInterval)

	if a.Config.Misc.SchedulingEnabled {
		loc := time.Local
		if a.Config.Misc.SchedulingTZ != "" && a.Config.Misc.SchedulingTZ != "Local" {
			l, err := time.LoadLocation(a.Config.Misc.SchedulingTZ)
			if err != nil {
				log.Fatalf("invalid scheduling timezone: %v", err)
			}
			loc = l
		}

		s := scheduler.NewPollingScheduler(a.Cache, a.Runtime, a.Config.Misc.SchedulingPoll, loc)
		s.Start(a.BaseCtx)
	}
}
