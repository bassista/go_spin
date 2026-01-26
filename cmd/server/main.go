
package main

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

func main() {
    // crea un router con le middleware default (logger, recovery)
    r := gin.Default()

    // route GET semplice
    r.GET("/hello", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "message": "Ciao da Gin!",
        })
    })

    // avvia il server sulla porta 8080
    r.Run(":8084")
}
