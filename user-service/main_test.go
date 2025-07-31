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
	r.POST("/users", func(c *gin.Context) {
		var req struct {
			Name     string `json:"name" binding:"required"`
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required,min=8"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"id": "test-user-id"})
	})

	r.GET("/users", func(c *gin.Context) {
		users := []gin.H{
			{"id": "user1", "name": "John Doe", "email": "john@example.com"},
			{"id": "user2", "name": "Jane Smith", "email": "jane@example.com"},
		}
		c.JSON(200, users)
	})

	r.GET("/users/:id", func(c *gin.Context) {
		id := c.Param("id")
		if id == "user1" {
			c.JSON(200, gin.H{"id": "user1", "name": "John Doe", "email": "john@example.com"})
		} else {
			c.JSON(404, gin.H{"error": "User not found"})
		}
	})

	r.PUT("/users/:id", func(c *gin.Context) {
		id := c.Param("id")
		if id == "user1" {
			c.JSON(200, gin.H{"status": "updated"})
		} else {
			c.JSON(404, gin.H{"error": "User not found"})
		}
	})

	r.DELETE("/users/:id", func(c *gin.Context) {
		id := c.Param("id")
		if id == "user1" {
			c.JSON(200, gin.H{"status": "deleted"})
		} else {
			c.JSON(404, gin.H{"error": "User not found"})
		}
	})

	return r
}

func TestCreateUser(t *testing.T) {
	r := setupTestRouter()

	// Test valid user creation
	userData := gin.H{
		"name":     "Test User",
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonData, _ := json.Marshal(userData)

	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response gin.H
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(t, response, "id")
}

func TestCreateUserValidation(t *testing.T) {
	r := setupTestRouter()

	// Test missing required fields
	userData := gin.H{
		"name": "Test User",
		// Missing email and password
	}
	jsonData, _ := json.Marshal(userData)

	req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetUsers(t *testing.T) {
	r := setupTestRouter()

	req, _ := http.NewRequest("GET", "/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var users []gin.H
	json.Unmarshal(w.Body.Bytes(), &users)
	assert.Len(t, users, 2)
}

func TestGetUserByID(t *testing.T) {
	r := setupTestRouter()

	// Test existing user
	req, _ := http.NewRequest("GET", "/users/user1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	
	var user gin.H
	json.Unmarshal(w.Body.Bytes(), &user)
	assert.Equal(t, "user1", user["id"])

	// Test non-existing user
	req, _ = http.NewRequest("GET", "/users/nonexistent", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateUser(t *testing.T) {
	r := setupTestRouter()

	updateData := gin.H{
		"name":  "Updated Name",
		"email": "updated@example.com",
	}
	jsonData, _ := json.Marshal(updateData)

	req, _ := http.NewRequest("PUT", "/users/user1", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteUser(t *testing.T) {
	r := setupTestRouter()

	req, _ := http.NewRequest("DELETE", "/users/user1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPasswordHashing(t *testing.T) {
	password := "testpassword123"
	
	hash, err := hashPassword(password)
	assert.NoError(t, err)
	assert.NotEqual(t, password, hash)
	
	// Test password verification
	assert.True(t, checkPassword(hash, password))
	assert.False(t, checkPassword(hash, "wrongpassword"))
} 