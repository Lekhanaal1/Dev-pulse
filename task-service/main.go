package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
	taskOperationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "task_operations_total",
			Help: "Total number of task operations",
		},
		[]string{"operation"},
	)
)

func init() {
	// Register metrics
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(taskOperationsTotal)
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

type Task struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Title       string             `json:"title" bson:"title" binding:"required"`
	Description string             `json:"description" bson:"description"`
	Status      string             `json:"status" bson:"status" binding:"required,oneof=todo in_progress done"`
	Priority    string             `json:"priority" bson:"priority" binding:"required,oneof=low medium high"`
	UserID      string             `json:"user_id" bson:"user_id" binding:"required"`
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

var client *mongo.Client
var collection *mongo.Collection

func connectDB() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://mongo:27017"
	}

	var err error
	client, err = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Ping the database
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	collection = client.Database("devpulse_tasks").Collection("tasks")
	log.Println("Connected to MongoDB successfully")
}

func main() {
	connectDB()
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}()

	r := gin.Default()

	// Add metrics middleware
	r.Use(metricsMiddleware())

	// Health check
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Prometheus metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Create task
	r.POST("/tasks", func(c *gin.Context) {
		taskOperationsTotal.WithLabelValues("create").Inc()
		var task Task
		if err := c.ShouldBindJSON(&task); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		task.CreatedAt = time.Now().UTC()
		task.UpdatedAt = time.Now().UTC()

		result, err := collection.InsertOne(context.Background(), task)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
			return
		}

		task.ID = result.InsertedID.(primitive.ObjectID)
		c.JSON(http.StatusCreated, task)
	})

	// Get all tasks
	r.GET("/tasks", func(c *gin.Context) {
		taskOperationsTotal.WithLabelValues("list").Inc()
		userID := c.Query("user_id")
		filter := bson.M{}
		if userID != "" {
			filter["user_id"] = userID
		}

		cursor, err := collection.Find(context.Background(), filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
			return
		}
		defer cursor.Close(context.Background())

		var tasks []Task
		if err = cursor.All(context.Background(), &tasks); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode tasks"})
			return
		}

		c.JSON(200, tasks)
	})

	// Get task by ID
	r.GET("/tasks/:id", func(c *gin.Context) {
		taskOperationsTotal.WithLabelValues("get").Inc()
		id := c.Param("id")
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
			return
		}

		var task Task
		err = collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&task)
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch task"})
			return
		}

		c.JSON(200, task)
	})

	// Update task
	r.PUT("/tasks/:id", func(c *gin.Context) {
		taskOperationsTotal.WithLabelValues("update").Inc()
		id := c.Param("id")
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
			return
		}

		var updateData map[string]interface{}
		if err := c.ShouldBindJSON(&updateData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		updateData["updated_at"] = time.Now().UTC()
		update := bson.M{"$set": updateData}

		result, err := collection.UpdateOne(
			context.Background(),
			bson.M{"_id": objectID},
			update,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
			return
		}

		if result.MatchedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}

		c.JSON(200, gin.H{"status": "updated"})
	})

	// Delete task
	r.DELETE("/tasks/:id", func(c *gin.Context) {
		taskOperationsTotal.WithLabelValues("delete").Inc()
		id := c.Param("id")
		objectID, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
			return
		}

		result, err := collection.DeleteOne(context.Background(), bson.M{"_id": objectID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
			return
		}

		if result.DeletedCount == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}

		c.JSON(200, gin.H{"status": "deleted"})
	})

	// Get tasks by status
	r.GET("/tasks/status/:status", func(c *gin.Context) {
		taskOperationsTotal.WithLabelValues("list_by_status").Inc()
		status := c.Param("status")
		userID := c.Query("user_id")

		filter := bson.M{"status": status}
		if userID != "" {
			filter["user_id"] = userID
		}

		cursor, err := collection.Find(context.Background(), filter)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
			return
		}
		defer cursor.Close(context.Background())

		var tasks []Task
		if err = cursor.All(context.Background(), &tasks); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode tasks"})
			return
		}

		c.JSON(200, tasks)
	})

	log.Println("Task service starting on :8000")
	r.Run(":8000")
}
