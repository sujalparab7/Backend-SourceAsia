package controllers

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type Incomingrequest struct {
	UserID  string `json:"user_id"`
	Payload any    `json:"payload"`
}

type UserRateLimit struct {
	Accepted    int
	Rejected    int
	WindowStart time.Time
}

var (
	rateLimiter = make(map[string]*UserRateLimit)
	mu          sync.Mutex
)

func RequestAdder(c *gin.Context) {
	var user Incomingrequest
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(400, gin.H{
			"error":   "Bad request",
			"details": err.Error(),
		})
		return
	}
	mu.Lock()
	defer mu.Unlock()
	newuser, exists := rateLimiter[user.UserID]
	if !exists {
		newuser = &UserRateLimit{
			Accepted:    0,
			Rejected:    0,
			WindowStart: time.Now(),
		}
		rateLimiter[user.UserID] = newuser
	}

	if time.Since(newuser.WindowStart) > time.Minute {
		newuser.Accepted = 0
		newuser.WindowStart = time.Now()
	}

	if newuser.Accepted >= 5 {
		newuser.Rejected++
		c.JSON(429, gin.H{"error": "Too many incoming requests"})
		return
	}
	newuser.Accepted++

	c.JSON(200, gin.H{
		"message": "Request Accepted",
		"user_id": user.UserID,
		"payload": user.Payload,
	})
}

func GetStats(c *gin.Context) {
	mu.Lock()
	defer mu.Unlock()
	status := make(map[string]map[string]int)
	for userID, userData := range rateLimiter {
		status[userID] = map[string]int{
			"accepted": userData.Accepted,
			"rejected": userData.Rejected,
		}
	}
	c.JSON(200, status)
}
