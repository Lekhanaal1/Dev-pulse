package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name" binding:"required"`
	Email        string    `json:"email" binding:"required,email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

var db *sql.DB

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}

func checkPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

func initDB() {
	connStr := os.Getenv("DB_CONN")
	if connStr == "" {
		connStr = "host=postgres user=postgres password=postgres dbname=devpulse_users sslmode=disable"
	}
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalf("DB not reachable: %v", err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL,
		updated_at TIMESTAMPTZ NOT NULL
	)`)
	if err != nil {
		log.Fatalf("Failed to migrate users table: %v", err)
	}
}

func main() {
	initDB()
	r := gin.Default()

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

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
		hash, err := hashPassword(req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		id := uuid.New()
		now := time.Now().UTC()
		_, err = db.Exec(`INSERT INTO users (id, name, email, password_hash, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
			id, req.Name, req.Email, hash, now, now)
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists or DB error"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"id": id})
	})

	r.GET("/users", func(c *gin.Context) {
		rows, err := db.Query(`SELECT id, name, email, created_at, updated_at FROM users`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
			return
		}
		defer rows.Close()
		var users []User
		for rows.Next() {
			var u User
			if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt); err == nil {
				users = append(users, u)
			}
		}
		c.JSON(200, users)
	})

	r.GET("/users/:id", func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID"})
			return
		}
		var u User
		err = db.QueryRow(`SELECT id, name, email, created_at, updated_at FROM users WHERE id=$1`, id).Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
			return
		}
		c.JSON(200, u)
	})

	r.PUT("/users/:id", func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID"})
			return
		}
		var req struct {
			Name     string `json:"name"`
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var set []string
		var args []interface{}
		idx := 1
		if req.Name != "" {
			set = append(set, "name=$"+string(rune(idx)))
			args = append(args, req.Name)
			idx++
		}
		if req.Email != "" {
			set = append(set, "email=$"+string(rune(idx)))
			args = append(args, req.Email)
			idx++
		}
		if req.Password != "" {
			hash, err := hashPassword(req.Password)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
				return
			}
			set = append(set, "password_hash=$"+string(rune(idx)))
			args = append(args, hash)
			idx++
		}
		if len(set) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
			return
		}
		set = append(set, "updated_at=$"+string(rune(idx)))
		args = append(args, time.Now().UTC())
		idx++
		args = append(args, id)
		_, err = db.Exec(`UPDATE users SET `+joinComma(set)+` WHERE id=$`+string(rune(idx)), args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
			return
		}
		c.JSON(200, gin.H{"status": "updated"})
	})

	r.DELETE("/users/:id", func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID"})
			return
		}
		_, err = db.Exec(`DELETE FROM users WHERE id=$1`, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
			return
		}
		c.JSON(200, gin.H{"status": "deleted"})
	})

	r.Run(":8000")
}

func joinComma(fields []string) string {
	return string([]byte(strings.Join(fields, ", ")))
}
