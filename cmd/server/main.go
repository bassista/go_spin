
package main

import (
    "fmt"
    "log"
    "net/http"
    "os"
    "syscall"
    "time"
    "github.com/bassista/go_spin/internal/config"
    "github.com/gin-gonic/gin"
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

	// Set graceful shutdown timeout (default: 10 seconds)
	httpgrace.WithTimeout(5*time.Second)

	// Customize shutdown signals (default: SIGINT, SIGTERM)
	httpgrace.WithSignals(syscall.SIGTERM, syscall.SIGUSR1)

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
    if err := srv.ListenAndServe(":"+appPort); err != nil {
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
