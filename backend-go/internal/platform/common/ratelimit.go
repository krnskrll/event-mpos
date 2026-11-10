package common

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimit allows n requests per window from one client IP; the rest get 429.
func RateLimit(n int, window time.Duration) gin.HandlerFunc {
	var mu sync.Mutex
	start, counts := time.Now(), map[string]int{}
	return func(c *gin.Context) {
		mu.Lock()
		if time.Since(start) >= window {
			start, counts = time.Now(), map[string]int{}
		}
		counts[c.ClientIP()]++
		over := counts[c.ClientIP()] > n
		mu.Unlock()
		if over {
			AbortError(c, http.StatusTooManyRequests, "Too many attempts, try again in a minute")
			return
		}
		c.Next()
	}
}
