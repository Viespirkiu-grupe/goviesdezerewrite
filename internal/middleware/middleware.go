package middleware

import (
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger logs all requests and their duration
func RequestLogger(debug bool) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		if debug {
			query := ""
			if param.Request.URL.RawQuery != "" {
				query = "?" + param.Request.URL.RawQuery
			}

			return fmt.Sprintf("[%s] %s %s%s - %d - %dms - ip=%s ua=%q in=%d out=%d err=%q\n",
				param.TimeStamp.Format(time.RFC3339),
				param.Method,
				param.Path,
				query,
				param.StatusCode,
				param.Latency.Milliseconds(),
				param.ClientIP,
				param.Request.UserAgent(),
				param.Request.ContentLength,
				param.BodySize,
				param.ErrorMessage,
			)
		}

		return fmt.Sprintf("[%s] %s %s - %d - %dms\n",
			param.TimeStamp.Format(time.RFC3339),
			param.Method,
			param.Path,
			param.StatusCode,
			param.Latency.Milliseconds(),
		)
	})
}

// APIKeyAuth middleware for API key authentication
func APIKeyAuth(apiKey string, debug bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if debug {
			log.Printf("debug auth: method=%s path=%s ip=%s", c.Request.Method, c.Request.URL.String(), c.ClientIP())
		}

		// Allow GET requests without API key
		if c.Request.Method == "GET" {
			c.Next()
			return
		}

		providedKey := c.GetHeader("X-API-Key")
		if providedKey != apiKey {
			if debug {
				log.Printf("debug auth denied: method=%s path=%s ip=%s", c.Request.Method, c.Request.URL.String(), c.ClientIP())
			}
			c.JSON(403, gin.H{"error": "Forbidden"})
			c.Abort()
			return
		}

		if debug {
			log.Printf("debug auth passed: method=%s path=%s ip=%s", c.Request.Method, c.Request.URL.String(), c.ClientIP())
		}

		c.Next()
	}
}
