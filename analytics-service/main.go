package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Prometheus metrics
var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
	analyticsEventsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "analytics_events_total",
			Help: "Total number of analytics events",
		},
		[]string{"event_type"},
	)
)

func init() {
	// Register metrics
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(analyticsEventsTotal)
}

// Metrics middleware
func metricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		status := c.Writer.Status()

		httpRequestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), string(rune(status))).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, c.FullPath()).Observe(duration)
	}
}

type AnalyticsEvent struct {
	EventType string                 `json:"event_type"`
	UserID    string                 `json:"user_id"`
	TaskID    string                 `json:"task_id,omitempty"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

var nc *nats.Conn

func connectNATS() {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://nats:4222"
	}

	var err error
	nc, err = nats.Connect(natsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	log.Println("Connected to NATS successfully")
}

func publishEvent(eventType, userID, taskID string, data map[string]interface{}) {
	event := AnalyticsEvent{
		EventType: eventType,
		UserID:    userID,
		TaskID:    taskID,
		Data:      data,
		Timestamp: time.Now().UTC(),
	}

	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal event: %v", err)
		return
	}

	err = nc.Publish("analytics.events", eventJSON)
	if err != nil {
		log.Printf("Failed to publish event: %v", err)
	} else {
		log.Printf("Published event: %s for user: %s", eventType, userID)
	}
}

func main() {
	connectNATS()
	defer nc.Close()

	r := gin.Default()

	// Add metrics middleware
	r.Use(metricsMiddleware())

	// Health check
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Analytics endpoint - called by other services
	r.POST("/analytics/event", func(c *gin.Context) {
		analyticsEventsTotal.WithLabelValues("received").Inc()
		var event AnalyticsEvent
		if err := c.ShouldBindJSON(&event); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		event.Timestamp = time.Now().UTC()
		eventJSON, err := json.Marshal(event)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal event"})
			return
		}

		err = nc.Publish("analytics.events", eventJSON)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish event"})
			return
		}

		analyticsEventsTotal.WithLabelValues("published").Inc()
		c.JSON(200, gin.H{"status": "event published"})
	})

	// Get analytics summary
	r.GET("/analytics/summary", func(c *gin.Context) {
		userID := c.Query("user_id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
			return
		}

		// In a real implementation, you would query a database
		// For now, return mock data
		summary := gin.H{
			"user_id":                 userID,
			"total_tasks":             15,
			"completed_tasks":         8,
			"in_progress_tasks":       4,
			"todo_tasks":              3,
			"completion_rate":         53.3,
			"average_completion_time": "2.5 days",
		}

		c.JSON(200, summary)
	})

	// Get user activity
	r.GET("/analytics/activity", func(c *gin.Context) {
		userID := c.Query("user_id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
			return
		}

		// Mock activity data
		activity := []gin.H{
			{
				"event_type": "task_created",
				"timestamp":  time.Now().Add(-2 * time.Hour),
				"task_id":    "task123",
				"title":      "Implement user authentication",
			},
			{
				"event_type": "task_completed",
				"timestamp":  time.Now().Add(-1 * time.Hour),
				"task_id":    "task122",
				"title":      "Setup database schema",
			},
		}

		c.JSON(200, activity)
	})

	log.Println("Analytics service starting on :8000")
	r.Run(":8000")
}
