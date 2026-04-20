package main

import (
	"log"
	"time"

	"github.com/mohammad-niyas/rate-limiter-api/internal/handler"
	"github.com/mohammad-niyas/rate-limiter-api/internal/ratelimiter"
	"github.com/mohammad-niyas/rate-limiter-api/internal/store"
	"github.com/gin-gonic/gin"
)

func main() {

	memStore := store.NewMemoryStore()

	cfg := ratelimiter.Config{
		MaxRequests: 5,
		WindowSize:  1 * time.Minute,
	}
	limiter := ratelimiter.NewRateLimiter(memStore, cfg)

	h := handler.NewHandler(limiter)

	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// Main API endpoints
	router.POST("/request", h.HandleRequest)
	router.GET("/stats", h.HandleStats)

	log.Println("Rate Limiter API starting on port 8080...")

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
