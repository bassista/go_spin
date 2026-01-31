package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-playground/validator/v10"
)

// CacheStore defines the interface for cache operations needed by the watcher callback.
type CacheStore interface {
	GetLastUpdate() int64
	IsDirty() bool
	Snapshot() (DataDocument, error)
	Replace(doc DataDocument) error
}

// JSONRepository handles disk persistence and watching of the data file.
type JSONRepository struct {
	path      string
	dir       string
	base      string
	validator *validator.Validate
	logger    *log.Logger
	mu        sync.Mutex
}

// NewJSONRepository creates a repository for the given JSON file path.
// It returns the repository interface to avoid leaking implementation details.
func NewJSONRepository(path string) (Repository, error) {
	if path == "" {
		return nil, errors.New("data file path is required")
	}

	dir := filepath.Dir(path)
	base := filepath.Base(path)
	if dir == "" || dir == "." {
		dir = "."
	}

	logger := log.New(os.Stdout, "[json-repo] ", log.LstdFlags)
	v := validator.New()
	return &JSONRepository{path: path, dir: dir, base: base, validator: v, logger: logger}, nil
}

// Load reads the JSON file, parses and validates it.
func (r *JSONRepository) Load() (*DataDocument, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.loadUnlocked()
}

// loadUnlocked reads the JSON file without acquiring the lock (caller must hold it).
func (r *JSONRepository) loadUnlocked() (*DataDocument, error) {
	file, err := os.Open(r.path)
	if err != nil {
		return nil, fmt.Errorf("open data file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	var doc DataDocument
	if err := decoder.Decode(&doc); err != nil {
		return nil, fmt.Errorf("decode data file: %w", err)
	}

	doc.ApplyDefaults()

	if r.validator != nil {
		if err := r.validator.Struct(&doc); err != nil {
			return nil, fmt.Errorf("validate data file: %w", err)
		}
	}

	return &doc, nil
}

// Save validates and writes the document atomically to disk.
func (r *JSONRepository) Save(doc *DataDocument) error {
	if doc == nil {
		return errors.New("document is nil")
	}
	if r.validator != nil {
		if err := r.validator.Struct(doc); err != nil {
			return fmt.Errorf("validate before save: %w", err)
		}
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	return r.saveUnlocked(doc)
}

// saveUnlocked writes the document without acquiring the lock (caller must hold it).
func (r *JSONRepository) saveUnlocked(doc *DataDocument) error {
	payload, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal data: %w", err)
	}

	tmpFile, err := os.CreateTemp(r.dir, r.base+".tmp-")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer func() {
		tmpFile.Close()
		os.Remove(tmpFile.Name())
	}()

	if _, err := tmpFile.Write(payload); err != nil {
		return fmt.Errorf("write temp file: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("sync temp file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Rename(tmpFile.Name(), r.path); err != nil {
		return fmt.Errorf("replace data file: %w", err)
	}

	return nil
}

// StartWatcher listens for changes to the data file and calls onChange after debounce.
// It watches the parent directory (not the file) so atomic replace sequences (temp+rename)
// are still observed on Linux and Windows. Events are filtered by basename and
// debounced to avoid double reloads on write+chmod/rename cycles. The caller owns the
// provided context: cancel it to stop the goroutine and close the watcher cleanly.
func (r *JSONRepository) StartWatcher(ctx context.Context, cacheStore CacheStore) error {
	onChange := r.MakeWatcherCallback(cacheStore)
	if onChange == nil {
		return errors.New("onChange callback is required")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create watcher: %w", err)
	}

	if err := watcher.Add(r.dir); err != nil {
		watcher.Close()
		return fmt.Errorf("watch dir: %w", err)
	}

	// Run watcher loop in the background; it exits when ctx is canceled or channels close.
	go func() {
		defer watcher.Close()

		// debounce coalesces bursty fsnotify events (write+chmod/rename) into a single reload.
		// If the timer is stopped before it fires, the scheduled onChange will not run.
		var debounce *time.Timer
		schedule := func() {
			if debounce != nil {
				if !debounce.Stop() {
					select {
					case <-debounce.C:
					default:
					}
				}
			}
			debounce = time.AfterFunc(200*time.Millisecond, onChange)
		}

		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if filepath.Base(event.Name) != r.base {
					continue
				}
				// Writes/Create/Chmod cover normal edits and atomic replace; trigger reload.
				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Chmod) != 0 {
					schedule()
					continue
				}
				// Remove/Rename indicates the file was moved or replaced; wait for next Create.
				if event.Op&(fsnotify.Remove|fsnotify.Rename) != 0 {
					schedule()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				r.logger.Printf("watcher error: %v", err)
			}
		}
	}()

	return nil
}

// MakeWatcherCallback returns a callback for file watcher that reloads cache from disk if needed.
func (r *JSONRepository) MakeWatcherCallback(cacheStore CacheStore) func() {
	return func() {
		diskDoc, loadErr := r.Load()
		if loadErr != nil {
			r.logger.Printf("watch reload failed: %v", loadErr)
			return
		}
		cacheLastUpdate := cacheStore.GetLastUpdate()
		diskLastUpdate := diskDoc.Metadata.LastUpdate

		// If disk is not newer, skip reload
		if diskLastUpdate < cacheLastUpdate {
			r.logger.Println("disk version is not newer than cache: diskLastUpdate =", diskLastUpdate, ", cacheLastUpdate =", cacheLastUpdate)
			return
		}

		if cacheStore.IsDirty() {
			r.logger.Println("Warning: disk data is newer but cache is dirty; skipping reload")
			// the cache content will be written to file soon anyway
			return
		}

		isDiskSameAsCache := false
		if diskLastUpdate == cacheLastUpdate {
			// check if disk content is really the same as cache content
			snapshot, err := cacheStore.Snapshot()
			if err != nil {
				r.logger.Printf("cache reload error: failed to get snapshot: %v", err)
				return
			}
			isDiskSameAsCache = AreDataDocumentsEqual(&snapshot, diskDoc)
		}
		if !isDiskSameAsCache {
			if err := cacheStore.Replace(*diskDoc); err != nil {
				r.logger.Printf("cache reload error: %v", err)
				return
			}
			r.logger.Println("cache reloaded from newer disk version")
		}
	}
}
