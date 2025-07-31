package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Add routes for testing
	r.POST("/tasks", func(c *gin.Context) {
		var req struct {
			Title       string `json:"title" binding:"required"`
			Description string `json:"description"`
			Status      string `json:"status" binding:"required,oneof=todo in_progress done"`
			Priority    string `json:"priority" binding:"required,oneof=low medium high"`
			UserID      string `json:"user_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"id": "test-task-id"})
	})

	r.GET("/tasks", func(c *gin.Context) {
		tasks := []gin.H{
			{"id": "task1", "title": "Test Task 1", "status": "todo", "priority": "high", "user_id": "user1"},
			{"id": "task2", "title": "Test Task 2", "status": "in_progress", "priority": "medium", "user_id": "user1"},
		}
		c.JSON(200, tasks)
	})

	r.GET("/tasks/:id", func(c *gin.Context) {
		id := c.Param("id")
		if id == "task1" {
			c.JSON(200, gin.H{"id": "task1", "title": "Test Task 1", "status": "todo", "priority": "high", "user_id": "user1"})
		} else {
			c.JSON(404, gin.H{"error": "Task not found"})
		}
	})

	r.PUT("/tasks/:id", func(c *gin.Context) {
		id := c.Param("id")
		if id == "task1" {
			c.JSON(200, gin.H{"status": "updated"})
		} else {
			c.JSON(404, gin.H{"error": "Task not found"})
		}
	})

	r.DELETE("/tasks/:id", func(c *gin.Context) {
		id := c.Param("id")
		if id == "task1" {
			c.JSON(200, gin.H{"status": "deleted"})
		} else {
			c.JSON(404, gin.H{"error": "Task not found"})
		}
	})

	r.GET("/tasks/status/:status", func(c *gin.Context) {
		status := c.Param("status")
		tasks := []gin.H{
			{"id": "task1", "title": "Test Task 1", "status": status, "priority": "high", "user_id": "user1"},
		}
		c.JSON(200, tasks)
	})

	return r
}

func TestCreateTask(t *testing.T) {
	r := setupTestRouter()

	// Test valid task creation
	taskData := gin.H{
		"title":       "Test Task",
		"description": "Test Description",
		"status":      "todo",
		"priority":    "high",
		"user_id":     "user1",
	}
	jsonData, _ := json.Marshal(taskData)

	req, _ := http.NewRequest("POST", "/tasks", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response gin.H
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response, "id")
}

func TestCreateTaskValidation(t *testing.T) {
	r := setupTestRouter()

	// Test missing required fields
	taskData := gin.H{
		"title": "Test Task",
		// Missing status, priority, user_id
	}
	jsonData, _ := json.Marshal(taskData)

	req, _ := http.NewRequest("POST", "/tasks", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateTaskInvalidStatus(t *testing.T) {
	r := setupTestRouter()

	// Test invalid status
	taskData := gin.H{
		"title":    "Test Task",
		"status":   "invalid_status",
		"priority": "high",
		"user_id":  "user1",
	}
	jsonData, _ := json.Marshal(taskData)

	req, _ := http.NewRequest("POST", "/tasks", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetTasks(t *testing.T) {
	r := setupTestRouter()

	req, _ := http.NewRequest("GET", "/tasks", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var tasks []gin.H
	json.Unmarshal(w.Body.Bytes(), &tasks)
	assert.Len(t, tasks, 2)
}

func TestGetTaskByID(t *testing.T) {
	r := setupTestRouter()

	// Test existing task
	req, _ := http.NewRequest("GET", "/tasks/task1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var task gin.H
	json.Unmarshal(w.Body.Bytes(), &task)
	assert.Equal(t, "task1", task["id"])

	// Test non-existing task
	req, _ = http.NewRequest("GET", "/tasks/nonexistent", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateTask(t *testing.T) {
	r := setupTestRouter()

	updateData := gin.H{
		"title":  "Updated Task",
		"status": "in_progress",
	}
	jsonData, _ := json.Marshal(updateData)

	req, _ := http.NewRequest("PUT", "/tasks/task1", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteTask(t *testing.T) {
	r := setupTestRouter()

	req, _ := http.NewRequest("DELETE", "/tasks/task1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetTasksByStatus(t *testing.T) {
	r := setupTestRouter()

	req, _ := http.NewRequest("GET", "/tasks/status/todo", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var tasks []gin.H
	json.Unmarshal(w.Body.Bytes(), &tasks)
	assert.Len(t, tasks, 1)
	assert.Equal(t, "todo", tasks[0]["status"])
}
