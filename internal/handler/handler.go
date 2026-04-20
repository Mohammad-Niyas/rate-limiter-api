package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mohammad-niyas/rate-limiter-api/internal/ratelimiter"
)

type RequestBody struct {
	UserID  string `json:"user_id" binding:"required"`
	Payload string `json:"payload"`
}

type Handler struct {
	limiter *ratelimiter.RateLimiter
}

func NewHandler(limiter *ratelimiter.RateLimiter) *Handler {
	return &Handler{
		limiter: limiter,
	}
}

func (h *Handler) HandleRequest(c *gin.Context) {
	var body RequestBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "user_id is required and must be a non-empty string",
		})
		return
	}

	result := h.limiter.Allow(body.UserID)

	if !result.Allowed {
		c.Header("Retry-After", fmt.Sprintf("%.0f", result.RetryAfter.Seconds()))
		c.JSON(http.StatusTooManyRequests, gin.H{
			"status":              "rejected",
			"message":             "Rate limit exceeded. Maximum 5 requests per minute allowed.",
			"retry_after_seconds": result.RetryAfter.Seconds(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "accepted",
		"message":   "Request processed successfully",
		"remaining": result.Remaining,
	})
}

func (h *Handler) HandleStats(c *gin.Context) {
	stats := h.limiter.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}
