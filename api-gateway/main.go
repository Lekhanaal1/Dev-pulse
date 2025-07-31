package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

// Rate limiter per IP
var limiter = rate.NewLimiter(2, 5) // 2 req/sec, burst 5

// JWT secret from env
var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// Reverse proxy setup
func reverseProxy(target string) gin.HandlerFunc {
	url, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Invalid proxy target: %v", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(url)
	return func(c *gin.Context) {
		c.Request.URL.Path = strings.TrimPrefix(c.Request.URL.Path, c.FullPath())
		proxy.ServeHTTP(c.Writer, c.Request)
	}
}

// JWT Auth middleware
func jwtAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.FullPath() == "/healthz" {
			c.Next()
			return
		}
		// Expect Authorization: Bearer <token>
		tokenString := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
			return
		}
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}
		c.Next()
	}
}

// Rate limiting middleware
func rateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			return
		}
		c.Next()
	}
}

// Logging middleware
func loggingMiddleware() gin.HandlerFunc {
	logger := logrus.New()
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)
		logger.Infof("%s %s %d %s", c.Request.Method, c.Request.URL.Path, c.Writer.Status(), latency)
	}
}

func main() {
	// Load backend URLs from env
	userSvc := os.Getenv("USER_SERVICE_URL")
	taskSvc := os.Getenv("TASK_SERVICE_URL")
	analyticsSvc := os.Getenv("ANALYTICS_SERVICE_URL")
	notificationSvc := os.Getenv("NOTIFICATION_SERVICE_URL")

	if userSvc == "" || taskSvc == "" || analyticsSvc == "" || notificationSvc == "" {
		log.Fatal("One or more backend service URLs are not set in env")
	}

	r := gin.Default()
	r.Use(loggingMiddleware(), rateLimitMiddleware(), jwtAuthMiddleware())

	// Health check (no auth)
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Proxy routes
	r.Any("/user/*proxyPath", reverseProxy(userSvc))
	r.Any("/task/*proxyPath", reverseProxy(taskSvc))
	r.Any("/analytics/*proxyPath", reverseProxy(analyticsSvc))
	r.Any("/notification/*proxyPath", reverseProxy(notificationSvc))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("API Gateway listening on :%s", port)
	r.Run(":" + port)
}
