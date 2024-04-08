package main

import (
	"encoding/json"
	"github.com/Missuo0o/goBank/model"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"net/http"
)

func RoleAuthMiddleware(allowRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		username := session.Get("username")
		role := session.Get("role")
		if username == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Not logged in",
			})
			c.Abort() // 防止调用后续的处理器
			return
		} else if role != allowRole {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "Unauthorized",
			})
			c.Abort() // 防止调用后续的处理器
			return
		}
		// 如果通过认证，则继续处理请求
		c.Next()
	}
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

		err := c.ShouldBindJSON(&loginRequest)
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
				_ = session.Save()

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
	r.PUT("/register", func(c *gin.Context) {
		var registerRequest model.User
		if err := c.ShouldBindJSON(&registerRequest); err != nil {
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
		_ = session.Save()
		c.JSON(http.StatusOK, gin.H{
			"message": "Logout successful",
		})
	})

	//isExist Account API
	r.GET("/open", RoleAuthMiddleware("C"), func(c *gin.Context) {
		session := sessions.Default(c)
		username := session.Get("username")

		data, _ := c.GetRawData()
		var jsonData map[string]interface{}
		_ = json.Unmarshal(data, &jsonData)

		typeValue := jsonData["type"].(string)

		var customer model.Customer
		var account []model.Account
		db.Select("id").Where("username = ?", username).First(&customer)
		db.Select("type").Where("id = ?", customer.ID).Find(&account)

		for _, v := range account {
			if v.Type == typeValue {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "Account already exists",
				})
				return
			}
		}

	})

	_ = r.Run(":8080")

}
