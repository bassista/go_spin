package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"syscall"
	"time"

	route "github.com/bassista/go_spin/internal/api/route"
	"github.com/bassista/go_spin/internal/cache"
	"github.com/bassista/go_spin/internal/config"
	"github.com/bassista/go_spin/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"

	"github.com/enrichman/httpgrace"
)

func main() {
	// Load .env file if it exists
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found (that's okay in production)")
	}

	config.LoadConfig("./config")

	validator := validator.New()
	repo, err := repository.NewJSONRepository(viper.GetString("data.file_path"), validator, log.New(os.Stdout, "[json-repo] ", log.LstdFlags))
	if err != nil {
		log.Fatalf("cannot init repository: %v", err)
	}

	cacheStore := cache.NewStore()
	initial, err := repo.Load()
	if err != nil {
		log.Fatalf("cannot load data file: %v", err)
	}
	if err := cacheStore.Replace(*initial); err != nil {
		log.Fatalf("cannot seed cache: %v", err)
	}
	cacheStore.SetLastUpdate(initial.Metadata.LastUpdate)

	ctx, stopWatchers := context.WithCancel(context.Background())
	defer stopWatchers()

	// Start scheduled persistence goroutine
	persistInterval := viper.GetInt("data.persist_interval_secs")
	if persistInterval <= 0 {
		persistInterval = 5 // default 5 seconds
	}
	cache.StartPersistenceScheduler(ctx, cacheStore, repo, time.Duration(persistInterval)*time.Second, log.New(os.Stdout, "[persist] ", log.LstdFlags))

	if err := repo.StartWatcher(ctx, func() {
		diskDoc, loadErr := repo.Load()
		if loadErr != nil {
			log.Printf("watch reload failed: %v", loadErr)
			return
		}
		cacheLastUpdate := cacheStore.GetLastUpdate()
		diskLastUpdate := diskDoc.Metadata.LastUpdate

		// If disk is newer, reload cache regardless of dirty flag
		if diskLastUpdate < cacheLastUpdate {
			log.Println("disk version is not newer than cache: diskLastUpdate =", diskLastUpdate, ", cacheLastUpdate =", cacheLastUpdate)
			return
		}

		if cacheStore.IsDirty() {
			log.Println("Warning: disk data is newer but cache is dirty; skipping reload")
			//the cache content will be write to file soon anyway
			return
		}

		isDiskSameAsCache := false
		if diskLastUpdate == cacheLastUpdate {
			//check if disk content is really the same as cache content
			snapshot, err := cacheStore.Snapshot()
			if err != nil {
				log.Printf("cache reload error: failed to get snapshot: %v", err)
				return
			}
			isDiskSameAsCache = repository.AreDataDocumentsEqual(&snapshot, diskDoc)
		}
		if !isDiskSameAsCache {
			if err := cacheStore.Replace(*diskDoc); err != nil {
				log.Printf("cache reload error: %v", err)
				return
			}
			log.Println("cache reloaded from newer disk version")
		}
	}); err != nil {
		log.Fatalf("cannot start watcher: %v", err)
	}

	port := viper.GetString("server.port")
	fmt.Println("Server running on port:", port)

	// Get environment variables with fallbacks
	appPort := getEnv("PORT", "8084")
	fmt.Printf("App will run on port: %s\n", appPort)

	// Check for required variables
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("API_KEY environment variable is required")
	}
	fmt.Println("All configuration loaded successfully!")

	ginMode := os.Getenv("GIN_MODE")
	gin.SetMode(ginMode)

	// crea un router con le middleware default (logger, recovery)
	r := gin.Default()

	// route GET semplice
	r.GET("/hello", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Ciao da Gin!",
		})
	})

	r.GET("/users/:id", func(c *gin.Context) {
		id := c.Param("id")
		c.JSON(http.StatusOK, gin.H{
			"user_id": id,
			"message": "Found your user!",
		})
	})

	timeout := time.Duration(1) * time.Second
	fmt.Println("DEBUG: About to call SetupRoutes") // <-- Aggiungi questo
	route.SetupRoutes(timeout, r, cacheStore, validator)
	fmt.Println("DEBUG: SetupRoutes finished") // <-- Aggiungi questo

	// Set graceful shutdown timeout (default: 10 seconds)
	httpgrace.WithTimeout(5 * time.Second)

	// Customize shutdown signals (default: SIGINT, SIGTERM)
	httpgrace.WithSignals(syscall.SIGTERM, syscall.SIGINT)

	// Provide custom logger (default: slog.Default())
	//httpgrace.WithLogger(customLogger)

	// Provide a function to run before shutdown
	httpgrace.WithBeforeShutdown(func() {
		fmt.Println("Shoutting down!")
		time.Sleep(3 * time.Second)
	})

	// Server httpgrace che wrappa Gin (che è un http.Handler)

	srv := httpgrace.NewServer(r,
		httpgrace.WithServerOptions(
			httpgrace.WithReadTimeout(10*time.Second),
			httpgrace.WithWriteTimeout(10*time.Second),
			httpgrace.WithIdleTimeout(120*time.Second),
			// or with your custom ServerOption
			func(srv *http.Server) {
				srv.ErrorLog = log.New(os.Stdout, "", 0)
			},
		),
	)

	// Avvio con shutdown gracevole già incluso
	if err := srv.ListenAndServe(":" + appPort); err != nil {
		log.Fatal(err)
	}

}

// Helper function for environment variables with defaults
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
