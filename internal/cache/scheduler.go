package cache

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/bassista/go_spin/internal/repository"
)

// StartPersistenceScheduler runs a goroutine that periodically flushes dirty cache to disk.
// On ctx.Done, it performs a final flush before returning.
func StartPersistenceScheduler(
	ctx context.Context,
	store PersistableStore,
	repo repository.Saver,
	interval time.Duration,
) {
	logger := log.New(os.Stdout, "[persist] ", log.LstdFlags)
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				// Final flush on shutdown
				flushCache(store, repo, logger)
				logger.Println("persistence scheduler stopped after final flush")
				return
			case <-ticker.C:
				flushCache(store, repo, logger)
			}
		}
	}()
}

// flushCache persists the cache to disk if dirty, using optimistic locking.
func flushCache(store PersistableStore, repo repository.Saver, logger *log.Logger) {
	if !store.IsDirty() {
		return
	}

	// Cache is dirty → persist
	snapshot, err := store.Snapshot()
	if err != nil {
		logger.Printf("persist error: failed to get snapshot: %v", err)
		return
	}

	snapshot.Metadata.LastUpdate = time.Now().UnixMilli()

	if err := repo.Save(&snapshot); err != nil {
		logger.Printf("persist error: failed to save: %v", err)
		return
	}

	store.ClearDirty()
	store.SetLastUpdate(snapshot.Metadata.LastUpdate)
	logger.Println("cache persisted to disk")
}
