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
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
)

// Rate limiter per IP - more reasonable limits
var limiter = rate.NewLimiter(10, 20) // 10 req/sec, burst 20

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
	logger.SetFormatter(&logrus.JSONFormatter{})
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)

		// Log with more context
		logger.WithFields(logrus.Fields{
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"status":     c.Writer.Status(),
			"latency_ms": latency.Milliseconds(),
			"user_agent": c.Request.UserAgent(),
			"ip":         c.ClientIP(),
		}).Info("HTTP Request")
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
	r.Use(loggingMiddleware(), rateLimitMiddleware())

	// Health check (no auth)
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Prometheus metrics endpoint (no auth)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Test route (no auth)
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Test endpoint working"})
	})

	// Public routes (no auth)
	r.POST("/login", func(c *gin.Context) {
		log.Printf("Login endpoint hit: %s", c.Request.URL.Path)
		c.JSON(200, gin.H{"message": "Login endpoint working"})
	})

	// Protected routes (with auth)
	userGroup := r.Group("/user")
	userGroup.Use(jwtAuthMiddleware())
	userGroup.Any("/*proxyPath", reverseProxy(userSvc))

	taskGroup := r.Group("/task")
	taskGroup.Use(jwtAuthMiddleware())
	taskGroup.Any("/*proxyPath", reverseProxy(taskSvc))

	analyticsGroup := r.Group("/analytics")
	analyticsGroup.Use(jwtAuthMiddleware())
	analyticsGroup.Any("/*proxyPath", reverseProxy(analyticsSvc))

	notificationGroup := r.Group("/notification")
	notificationGroup.Use(jwtAuthMiddleware())
	notificationGroup.Any("/*proxyPath", reverseProxy(notificationSvc))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("API Gateway listening on :%s", port)
	r.Run(":" + port)
}
