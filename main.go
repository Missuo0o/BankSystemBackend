package main

import (
	"github.com/Missuo0o/goBank/model"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"net/http"
)

func getSession(c *gin.Context) (bool, string, string) {
	session := sessions.Default(c)
	username := session.Get("username")
	role := session.Get("role")
	if username == nil {
		return false, "", ""
	}
	return true, username.(string), role.(string)
}

func main() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	//r.Use(cors.Default())
	db, _ := connectDB()
	r.Static("/static", "./static")

	store := cookie.NewStore([]byte("secret123"))
	r.Use(sessions.Sessions("session", store))

	// Login API
	r.POST("/login", func(c *gin.Context) {
		var loginRequest model.User

		err := c.ShouldBind(&loginRequest)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid request",
			})
			return
		}

		var user []model.User
		// Check loginRequest.Username and loginRequest.Password
		db.Where("username = ?", loginRequest.Username).First(&user)
		if len(user) == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid username or password",
			})
			return
		} else {
			if checkPasswordHash(loginRequest.Password, user[0].Password) {

				session := sessions.Default(c)
				session.Set("username", user[0].Username)
				session.Set("role", user[0].Role)
				session.Save()

				c.JSON(http.StatusOK, gin.H{
					"message": "Login successful",
					"data": gin.H{
						"username": user[0].Username,
						"role":     user[0].Role},
				})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": "Invalid username or password",
				})
				return
			}
		}
	})

	// User Register API
	r.POST("/user/register", func(c *gin.Context) {
		var registerRequest model.User
		if err := c.ShouldBind(&registerRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid request",
			})
			return
		}
		registerRequest.Password = hashPassword(registerRequest.Password)
		registerRequest.Role = "C"
		result := db.Create(registerRequest)
		if result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Username already exists",
			})
			return
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message": "Register successful",
			})
		}

	})

	// Logout API
	r.GET("/logout", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()
		session.Save()
		c.JSON(http.StatusOK, gin.H{
			"message": "Logout successful",
		})
	})

	err := r.Run(":8080")
	if err != nil {
		return
	}

}
