package main

import (
	"github.com/Missuo0o/goBank/model"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"math/rand"
	"net/http"
	"time"
)

func hashPassword(password string) string {
	bytes, _ := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes)
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// RoleAuthMiddleware is a middleware that checks if the user is logged in and has the correct role
func RoleAuthMiddleware(allowRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		username := session.Get("username")
		role := session.Get("role")
		if username == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Not logged in",
			})
			c.Abort() // Prevent subsequent handlers from being called
			return
		} else if role != allowRole {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Unauthorized",
			})
			c.Abort() // Prevent subsequent handlers from being called
			return
		}
		// If authentication is passed, continue processing the request
		c.Next()
	}
}

// generateRandomNumber
func generateRandomNumber() int64 {
	rand.Seed(time.Now().UnixNano()) // 设置随机种子

	// 生成一个在1000000000到9999999999之间的随机数字
	minNumber := int64(1000000000)
	maxNumber := int64(9999999999)
	return rand.Int63n(maxNumber-minNumber+1) + minNumber
}

func generateUniqueRandomNumberString(db *gorm.DB) int64 {
	var number []int64
	db.Model(&model.Account{}).Select("number").Find(&number)

	for {
		newNumber := generateRandomNumber()
		isUnique := true
		for _, num := range number {
			if num == newNumber {
				isUnique = false
				break
			}
		}
		if isUnique {
			return newNumber
		}
	}
}
