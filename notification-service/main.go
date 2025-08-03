package main

import (
	"encoding/json"
	"fmt"
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
	notificationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notifications_total",
			Help: "Total number of notifications",
		},
		[]string{"type"},
	)
)

func init() {
	// Register metrics
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(notificationsTotal)
}

// Metrics middleware
func metricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		status := c.Writer.Status()

		httpRequestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), fmt.Sprintf("%d", status)).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, c.FullPath()).Observe(duration)
	}
}

type NotificationEvent struct {
	EventType string                 `json:"event_type"`
	UserID    string                 `json:"user_id"`
	TaskID    string                 `json:"task_id,omitempty"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

type Notification struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
}

var nc *nats.Conn
var notifications []Notification

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

func handleAnalyticsEvent(msg *nats.Msg) {
	var event NotificationEvent
	if err := json.Unmarshal(msg.Data, &event); err != nil {
		log.Printf("Failed to unmarshal event: %v", err)
		return
	}

	// Create notification based on event type
	var notification Notification
	switch event.EventType {
	case "task_created":
		notification = Notification{
			ID:        generateID(),
			UserID:    event.UserID,
			Type:      "task_created",
			Title:     "New Task Created",
			Message:   "A new task has been created",
			Read:      false,
			CreatedAt: time.Now().UTC(),
		}
	case "task_completed":
		notification = Notification{
			ID:        generateID(),
			UserID:    event.UserID,
			Type:      "task_completed",
			Title:     "Task Completed",
			Message:   "A task has been marked as completed",
			Read:      false,
			CreatedAt: time.Now().UTC(),
		}
	case "task_updated":
		notification = Notification{
			ID:        generateID(),
			UserID:    event.UserID,
			Type:      "task_updated",
			Title:     "Task Updated",
			Message:   "A task has been updated",
			Read:      false,
			CreatedAt: time.Now().UTC(),
		}
	default:
		log.Printf("Unknown event type: %s", event.EventType)
		return
	}

	// Store notification (in memory for now, would be database in production)
	notifications = append(notifications, notification)
	log.Printf("Created notification: %s for user: %s", notification.Type, notification.UserID)

	// In a real implementation, you would send email/SMS/push notification here
	log.Printf("Sending notification to user %s: %s", notification.UserID, notification.Message)
}

func generateID() string {
	return time.Now().Format("20060102150405")
}

func main() {
	connectNATS()
	defer nc.Close()

	// Subscribe to analytics events
	_, err := nc.Subscribe("analytics.events", handleAnalyticsEvent)
	if err != nil {
		log.Fatalf("Failed to subscribe to analytics events: %v", err)
	}
	log.Println("Subscribed to analytics.events")

	r := gin.Default()

	// Add metrics middleware
	r.Use(metricsMiddleware())

	// Health check
	r.GET("/healthz", func(c *gin.Context) {
		// Check NATS connection
		if nc.IsConnected() {
			c.JSON(200, gin.H{"status": "healthy", "service": "notification-service"})
		} else {
			c.JSON(503, gin.H{"status": "unhealthy", "error": "NATS connection failed"})
		}
	})

	// Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Get notifications for a user
	r.GET("/notifications", func(c *gin.Context) {
		userID := c.Query("user_id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
			return
		}

		var userNotifications []Notification
		for _, n := range notifications {
			if n.UserID == userID {
				userNotifications = append(userNotifications, n)
			}
		}

		c.JSON(200, userNotifications)
	})

	// Mark notification as read
	r.PUT("/notifications/:id/read", func(c *gin.Context) {
		id := c.Param("id")
		for i, n := range notifications {
			if n.ID == id {
				notifications[i].Read = true
				notificationsTotal.WithLabelValues("marked_read").Inc()
				c.JSON(200, gin.H{"status": "marked as read"})
				return
			}
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
	})

	// Send manual notification
	r.POST("/notifications", func(c *gin.Context) {
		notificationsTotal.WithLabelValues("manual").Inc()
		var req struct {
			UserID  string `json:"user_id" binding:"required"`
			Type    string `json:"type" binding:"required"`
			Title   string `json:"title" binding:"required"`
			Message string `json:"message" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		notification := Notification{
			ID:        generateID(),
			UserID:    req.UserID,
			Type:      req.Type,
			Title:     req.Title,
			Message:   req.Message,
			Read:      false,
			CreatedAt: time.Now().UTC(),
		}

		notifications = append(notifications, notification)
		log.Printf("Created manual notification: %s for user: %s", notification.Type, notification.UserID)

		c.JSON(http.StatusCreated, notification)
	})

	// Get notification stats
	r.GET("/notifications/stats", func(c *gin.Context) {
		userID := c.Query("user_id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
			return
		}

		total := 0
		unread := 0
		for _, n := range notifications {
			if n.UserID == userID {
				total++
				if !n.Read {
					unread++
				}
			}
		}

		stats := gin.H{
			"user_id":              userID,
			"total_notifications":  total,
			"unread_notifications": unread,
			"read_notifications":   total - unread,
		}

		c.JSON(200, stats)
	})

	log.Println("Notification service starting on :8000")
	r.Run(":8000")
}
