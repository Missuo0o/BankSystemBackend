package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Missuo0o/goBank/model"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

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
			fmt.Println(err)
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

	// Open Account API
	r.PUT("/open", func(c *gin.Context) {
		session := sessions.Default(c)
		username := session.Get("username")

		var openAccountRequest OpenAccountRequest
		err := c.ShouldBindJSON(&openAccountRequest)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid request",
			})
			return
		}

		// select type from account where id = (select id from customer where username = username) and type = typeValue
		var accountType string
		db.Model(model.Account{}).Select("type").
			Where("id = (?)", db.Model(model.Customer{}).
				Select("id").Where("username = ?", username)).
			Where("type = ?", openAccountRequest.Type).Find(&accountType)

		if accountType != "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Account already exists",
			})
			return
		}

		var customer model.Customer
		var account model.Account
		var checking model.Checking
		var saving model.Saving
		var loan model.Loan
		var studentLoan model.StudentLoan
		var homeLoan model.HomeLoan

		_ = copier.Copy(&customer, &openAccountRequest)
		_ = copier.Copy(&account, &openAccountRequest)
		_ = copier.Copy(&checking, &openAccountRequest)
		_ = copier.Copy(&saving, &openAccountRequest)
		_ = copier.Copy(&loan, &openAccountRequest)
		_ = copier.Copy(&studentLoan, &openAccountRequest)
		_ = copier.Copy(&homeLoan, &openAccountRequest)

		// if select id from customer where username = username is null, then insert into customer
		db.Where("username = ?", username).First(&customer)
		if customer.ID == 0 {
			db.Create(&customer)
		}

		uniqueNumberString := generateUniqueRandomNumberString(db)
		account.Number = uniqueNumberString
		checking.Number = uniqueNumberString
		saving.Number = uniqueNumberString
		loan.Number = uniqueNumberString
		studentLoan.Number = uniqueNumberString
		homeLoan.Number = uniqueNumberString
		account.OpenDate = time.Now().Format("2006-01-02")

		switch requestType := openAccountRequest.Type; requestType {
		case "C":
			db.Begin()
			if err := db.Create(&account).Error; err != nil {
				db.Rollback()
			}
			if err := db.Create(&checking).Error; err != nil {
				db.Rollback()
			}
			db.Commit()

		case "S":
			db.Begin()
			if err := db.Create(&account).Error; err != nil {
				db.Rollback()
			}
			if err := db.Create(&saving).Error; err != nil {
				db.Rollback()
			}
			db.Commit()

		case "L":
			// SELECT l.type
			//FROM loan l
			//JOIN account a ON l.number = a.number
			//JOIN customer c ON a.id = c.id
			//WHERE l.type = 'STUDENT' AND c.username = 'vincent';
			var loanType string
			db.Model(model.Loan{}).Select("type").
				Joins("JOIN account a ON loan.number = a.number").
				Joins("JOIN customer c ON a.id = c.id").
				Where("loan.type = '?' AND c.username = ?", openAccountRequest.LoanType, username).Find(&loanType)
			if loanType != "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "LoanAccount already exists",
				})
			}
			switch openAccountRequest.LoanType {
			case "student":
				db.Begin()
				if err := db.Create(&account).Error; err != nil {
					db.Rollback()
				}
				if err := db.Create(&loan).Error; err != nil {
					db.Rollback()
				}
				if err := db.Create(&studentLoan).Error; err != nil {
					db.Rollback()
				}
				db.Commit()
			case "home":
				db.Begin()
				if err := db.Create(&account).Error; err != nil {
					db.Rollback()
				}
				if err := db.Create(&loan).Error; err != nil {
					db.Rollback()
				}
				if err := db.Create(&homeLoan).Error; err != nil {
					db.Rollback()
				}
				db.Commit()
			case "personal":
				db.Begin()
				if err := db.Create(&account).Error; err != nil {
					db.Rollback()
				}
				if err := db.Create(&loan).Error; err != nil {
					db.Rollback()
				}
				db.Commit()

			}
		}
	})

	r.GET("/account", func(c *gin.Context) {
		var accounts []model.Account
		db.Model(model.Account{}).Find(&accounts)
		for i := range accounts {
			fmt.Println(accounts[i].OpenDate)
		}
	})

	// Reset Password API
	r.POST("/reset", func(c *gin.Context) {
		var resetRequest model.User
		err := c.ShouldBind(&resetRequest)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid request",
			})
			return
		}

		// update user set password = hashPassword(resetRequest.Password) where username = resetRequest.Username and keyword = resetRequest.Keyword
		result := db.Model(&model.User{}).Where("username = ? AND keyword = ?", resetRequest.Username, resetRequest.Keyword).Update("password", hashPassword(resetRequest.Password))

		if result.RowsAffected == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid username or keyword",
			})
			return
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message": "Reset password successful",
			})
		}
	})

	//Transfer Fund API
	type TransferRequest struct {
		FromAccountID string  `json:"fromAccountId"`
		ToAccountID   string  `jason:"toAccountId"`
		Amount        float64 `json:"amount"`
	}

	r.POST("/transfer", func(c *gin.Context) {
		var transferReq TransferRequest
		if err := c.ShouldBindJSON(&tranferReq); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"message: Invalid request"})
			return
		}
		// Start a database transaction
		tx := db.Begin()

		//Check if the source account has sufficient fund to transfer
		var sourceBalance float64
		tx.Model(&model.Account{}).Where("id=?", tranferReq.FromAccountID).Select("balance").Row().Scan(&sourceBalance)

		if sourceBalance < transferReq.Amount {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"message: Insufficient fund "})
			return

		}
		//Deduct the amount of transferred fund from source account balance
		if err := tx.Model(&model.Account{}).Where("id = ?"), transferReq.FromAccountID.Update("balance", gorm.Expr("balance - ?", transferReq.Amount)).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Transfer failed"})
			return
		}

		//Add amount to the destination account
		if err := tx.Model(&model.Account{}).Where("id = ?", transferReq.ToAccountID).Update("balance", gorm.Expr("balance + ?", transferReq.Amount)).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Transfer failed"})
			return
		}

		//Commit the transaction
		if err := tx.Commit().Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": "Transfer failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Transfer successful"})

	})

	_ = r.Run(":8080")

}
