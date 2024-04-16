package main

import (
	"github.com/Missuo0o/goBank/model"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"net/http"
	"time"
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
		// select * from user where username = loginRequest.Username
		db.Model(model.User{}).Where("username = ?", loginRequest.Username).First(&user)
		if len(user) == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid username or password1",
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
					"error": "Invalid username or password2",
				})
				return
			}
		}
	})

	// User Register API
	r.POST("/register", func(c *gin.Context) {
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
	r.POST("/logout", func(c *gin.Context) {
		session := sessions.Default(c)
		session.Clear()
		_ = session.Save()
		c.JSON(http.StatusOK, gin.H{
			"message": "Logout successful",
		})
	})

	// Open Account API
	r.PUT("/open", RoleAuthMiddleware("C"), func(c *gin.Context) {
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
				"message": "Account already exists1",
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
			customer.Username = username.(string)
			db.Create(&customer)
		}

		uniqueNumberString := generateUniqueRandomNumberString(db)
		account.Number = uniqueNumberString
		checking.Number = uniqueNumberString
		saving.Number = uniqueNumberString
		loan.Number = uniqueNumberString
		studentLoan.Number = uniqueNumberString
		homeLoan.Number = uniqueNumberString
		account.ID = customer.ID
		account.OpenDate = time.Now().Format("2006-01-02 15:04:05")

		switch requestType := openAccountRequest.Type; requestType {
		case "C":
			checking.Charge = 10.00
			tx := db.Begin()
			if err := db.Create(&account).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{
					"message": err.Error(),
				})
				return
			}
			if err := db.Create(&checking).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{
					"message": err.Error(),
				})
				return
			}
			tx.Commit()
			c.JSON(http.StatusOK, gin.H{
				"message": "Account opened successfully",
			})

		case "S":
			saving.Rate = 1.50
			tx1 := db.Begin()
			if err := db.Create(&account).Error; err != nil {
				tx1.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{
					"message": err.Error(),
				})
				return
			}
			if err := db.Create(&saving).Error; err != nil {
				tx1.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{
					"message": err.Error(),
				})
				return
			}
			tx1.Commit()
			c.JSON(http.StatusOK, gin.H{
				"message": "Account opened successfully",
			})

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
				Where("loan.type = ? AND c.username = ?", openAccountRequest.LoanType, username).Find(&loanType)
			if loanType != "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "LoanAccount already exists",
				})
				return
			}

			loan.Rate = 5.00
			interest := loan.Amount * loan.Rate * 0.01 * float64(loan.Months)
			loan.Payment = (loan.Amount + interest) / float64(loan.Months)

			switch openAccountRequest.LoanType {
			case "STUDENT":
				loan.Type = "STUDENT"
				tx2 := db.Begin()
				if err := db.Create(&account).Error; err != nil {
					tx2.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": err.Error(),
					})
					return
				}
				if err := db.Create(&loan).Error; err != nil {
					tx2.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": err.Error(),
					})
					return
				}
				// select * from university where name = openAccountRequest.UniversityName
				var university model.University
				db.Where("name = ?", openAccountRequest.UniversityName).First(&university)
				if university.Id == 0 {
					university.Name = openAccountRequest.UniversityName
					if err := db.Create(&university).Error; err != nil {
						tx2.Rollback()
						c.JSON(http.StatusBadRequest, gin.H{
							"message": err.Error(),
						})
					}
				}
				studentLoan.UniversityID = university.Id
				if err := db.Create(&studentLoan).Error; err != nil {
					tx2.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": err.Error(),
					})
					return
				}
				tx2.Commit()
				c.JSON(http.StatusOK, gin.H{
					"message": "Account opened successfully",
				})
			case "HOME":
				loan.Type = "HOME"
				tx3 := db.Begin()
				if err := db.Create(&account).Error; err != nil {
					tx3.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": err.Error(),
					})
					return
				}
				if err := db.Create(&loan).Error; err != nil {
					tx3.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": err.Error(),
					})
					return
				}
				if err := db.Create(&homeLoan).Error; err != nil {
					tx3.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": err.Error(),
					})
					return
				}
				tx3.Commit()
				c.JSON(http.StatusOK, gin.H{
					"message": "Account opened successfully",
				})
			case "PERSONAL":
				loan.Type = "PERSONAL"
				tx4 := db.Begin()
				if err := db.Create(&account).Error; err != nil {
					tx4.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": err.Error(),
					})
					return
				}
				if err := db.Create(&loan).Error; err != nil {
					tx4.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": err.Error(),
					})
					return
				}
				tx4.Commit()

				c.JSON(http.StatusOK, gin.H{
					"message": "Account opened successfully",
				})
			}
		}
	})

	// Reset Password API
	r.POST("/reset", func(c *gin.Context) {
		var resetRequest model.User
		err := c.ShouldBindJSON(&resetRequest)
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
				"message": "Password reset successful",
			})
		}
	})

	// Transfer API
	r.POST("/transfer", RoleAuthMiddleware("C"), func(c *gin.Context) {
		session := sessions.Default(c)
		username := session.Get("username")

		var transferRequest model.TransferHistory
		err := c.ShouldBindJSON(&transferRequest)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid request",
			})
			return
		}
		// select type from account where number = transferRequest.FromAccount
		var fromAccountType string
		db.Model(model.Account{}).Select("type").Where("number = ?", transferRequest.FromAccountNumber).Find(&fromAccountType)
		// select number from account where id = (select id from customer where username = username) and type = 'C' or type = 'S'
		var fromAccountNumber string
		db.Model(model.Account{}).Select("number").Where("id = (?)", db.Model(model.Customer{}).Select("id").Where("username = ?", username)).Where("type = 'C' OR type = 'S'").Find(&fromAccountNumber)
		if fromAccountType == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "FromAccount does not exist",
			})
		}
		// select type from account where number = transferRequest.ToAccount
		var toAccountType string
		db.Model(model.Account{}).Select("type").Where("number = ?", transferRequest.ToAccountNumber).Find(&toAccountType)
		if fromAccountType == "" || toAccountType == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "FromAccount does not exist",
			})
		}
		switch fromAccountType {
		case "C":
			switch toAccountType {
			case "C":
				tx := db.Begin()
				// update checking set balance = balance - transferRequest.Amount where number = transferRequest.FromAccount and balance >= transferRequest.Amount
				result := db.Model(model.Checking{}).Where("number = ? AND balance >= ?", transferRequest.FromAccountNumber, transferRequest.Amount).Update("balance", gorm.Expr("balance - ?", transferRequest.Amount))
				if result.RowsAffected == 0 {
					tx.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": "Insufficient balance",
					})
					return
				}
				// update checking set balance = balance + transferRequest.Amount where number = transferRequest.ToAccount
				if err := db.Model(model.Checking{}).Where("number = ?", transferRequest.ToAccountNumber).Update("balance", gorm.Expr("balance + ?", transferRequest.Amount)).Error; err != nil {
					tx.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": "ToAccount does not exist",
					})
					return
				}
				// insert into transfer_history
				transferRequest.TransferDate = time.Now()
				transferRequest.AccountType = "C"
				if err := db.Create(&transferRequest).Error; err != nil {
					tx.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": "Transfer failed",
					})
					return
				}
				tx.Commit()
				c.JSON(http.StatusOK, gin.H{
					"message": "Transfer successful",
				})
			case "S":
				tx1 := db.Begin()
				// update checking set balance = balance - transferRequest.Amount where number = transferRequest.FromAccount and balance >= transferRequest.Amount
				result := db.Model(model.Checking{}).Where("number = ? AND balance >= ?", transferRequest.FromAccountNumber, transferRequest.Amount).Update("balance", gorm.Expr("balance - ?", transferRequest.Amount))
				if result.RowsAffected == 0 {
					tx1.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": "Insufficient balance",
					})
					return
				}
				// update saving set balance = balance + transferRequest.Amount where number = transferRequest.ToAccount
				if err := db.Model(model.Saving{}).Where("number = ?", transferRequest.ToAccountNumber).Update("balance", gorm.Expr("balance + ?", transferRequest.Amount)).Error; err != nil {
					tx1.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": "ToAccount does not exist",
					})
					return

				}
				// insert into transfer_history
				transferRequest.TransferDate = time.Now()
				transferRequest.AccountType = "S"
				if err := db.Create(&transferRequest).Error; err != nil {
					tx1.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": "Transfer failed",
					})
					return
				}
				tx1.Commit()
				c.JSON(http.StatusOK, gin.H{
					"message": "Transfer successful",
				})
			}
		case "S":
			switch toAccountType {
			case "C":
				tx2 := db.Begin()
				// update saving set balance = balance - transferRequest.Amount where number = transferRequest.FromAccount and balance >= transferRequest.Amount
				result := db.Model(model.Saving{}).Where("number = ? AND balance >= ?", transferRequest.FromAccountNumber, transferRequest.Amount).Update("balance", gorm.Expr("balance - ?", transferRequest.Amount))
				if result.RowsAffected == 0 {
					tx2.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": "Insufficient balance",
					})
					return
				}
				// update checking set balance = balance + transferRequest.Amount where number = transferRequest.ToAccount
				if err := db.Model(model.Checking{}).Where("number = ?", transferRequest.ToAccountNumber).Update("balance", gorm.Expr("balance + ?", transferRequest.Amount)).Error; err != nil {
					tx2.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": "ToAccount does not exist",
					})
					return
				}
				// insert into transfer_history
				transferRequest.TransferDate = time.Now()
				transferRequest.AccountType = "C"
				if err := db.Create(&transferRequest).Error; err != nil {
					tx2.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": "Transfer failed",
					})
					return
				}
				tx2.Commit()
				c.JSON(http.StatusOK, gin.H{
					"message": "Transfer successful",
				})
			case "S":
				tx3 := db.Begin()
				// update saving set balance = balance - transferRequest.Amount where number = transferRequest.FromAccount and balance >= transferRequest.Amount
				result := db.Model(model.Saving{}).Where("number = ? AND balance >= ?", transferRequest.FromAccountNumber, transferRequest.Amount).Update("balance", gorm.Expr("balance - ?", transferRequest.Amount))
				if result.RowsAffected == 0 {
					tx3.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": "Insufficient balance",
					})
					return
				}
				// update saving set balance = balance + transferRequest.Amount where number = transferRequest.ToAccount
				if err := db.Model(model.Saving{}).Where("number = ?", transferRequest.ToAccountNumber).Update("balance", gorm.Expr("balance + ?", transferRequest.Amount)).Error; err != nil {
					tx3.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": "ToAccount does not exist",
					})
					return
				}
				// insert into transfer_history
				transferRequest.TransferDate = time.Now()
				transferRequest.AccountType = "S"
				if err := db.Create(&transferRequest).Error; err != nil {
					tx3.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": "Transfer failed",
					})
					return
				}
				tx3.Commit()
				c.JSON(http.StatusOK, gin.H{
					"message": "Transfer successful",
				})
			}
		}
	})

	// Get Transfer History API
	r.GET("/transfer", RoleAuthMiddleware("C"), func(c *gin.Context) {
		session := sessions.Default(c)
		username := session.Get("username")
		var transferHistory []model.TransferHistory
		// select * from transfer_history where from_account_number in (select number from account where type = 'C' or type = 'S' and id = (select id from customer where username = username))
		db.Where("from_account_number in (?)", db.Model(model.Account{}).Select("number").Where("type = 'C' OR type = 'S' AND id = (?)", db.Model(model.Customer{}).Select("id").Where("username = ?", username))).Find(&transferHistory)
		c.JSON(http.StatusOK, gin.H{
			"data": transferHistory,
		})
	})

	// Deposit API
	r.POST("/deposit", RoleAuthMiddleware("C"), func(c *gin.Context) {
		session := sessions.Default(c)
		username := session.Get("username")

		var depositRequest Deposit

		err := c.ShouldBindJSON(&depositRequest)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid request",
			})
			return
		}

		// 判断当前username是否有对应的账户
		// select number from account where id in (select id from customer where username = username) and number =depositRequest.Account and (type = 'C' or type = 'S')
		var accountNumber []int64
		db.Model(model.Account{}).Select("number").Where("id in (?)", db.Model(model.Customer{}).Select("id").Where("username = ?", username)).Where("number = ? AND (type = 'C' OR type = 'S')", depositRequest.Account).Find(&accountNumber)

		if len(accountNumber) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Account does not exist1",
			})
			return
		}

		// select type from account where number = depositRequest.Account
		var accountType string
		db.Model(model.Account{}).Select("type").Where("number = ?", depositRequest.Account).Find(&accountType)
		if accountType == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Account does not exist",
			})
			return
		}
		switch accountType {
		case "C":
			// update checking set balance = balance + balance where number = depositRequest.AccountNumber
			if err := db.Model(model.Checking{}).Where("number = ?", depositRequest.Account).Update("balance", gorm.Expr("balance + ?", depositRequest.Balance)).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "Account does not exist",
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"message": "Deposit successful",
			})
		case "S":
			// update saving set balance = balance + balance where number = depositRequest.AccountNumber
			if err := db.Model(model.Saving{}).Where("number = ?", depositRequest.Account).Update("balance", gorm.Expr("balance + ?", depositRequest.Balance)).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "Account does not exist",
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"message": "Deposit successful",
			})
		}
	})

	// Get AccountNumber API
	r.GET("/account", RoleAuthMiddleware("C"), func(c *gin.Context) {
		session := sessions.Default(c)
		username := session.Get("username")
		var accountNumber []string
		// select number from account where id = (select id from customer where username = username)
		db.Model(model.Account{}).Select("number").Where("id = (?)", db.Model(model.Customer{}).Select("id").Where("username = ?", username)).Find(&accountNumber)
		c.JSON(http.StatusOK, gin.H{
			"data": accountNumber,
		})
	})

	// Get AccountBalance API
	r.GET("/balance", RoleAuthMiddleware("C"), func(c *gin.Context) {
		session := sessions.Default(c)
		username := session.Get("username")
		var accountBalance []float64
		// select balance from checking where number in (select number from account where id = (select id from customer where username = username) and type = 'C')
		var checkingBalance float64
		db.Model(model.Checking{}).Select("balance").Where("number in (?)", db.Model(model.Account{}).Select("number").Where("id = (?) AND type = 'C'", db.Model(model.Customer{}).Select("id").Where("username = ?", username))).Find(&checkingBalance)
		accountBalance = append(accountBalance, checkingBalance)
		// select balance from saving where number in (select number from account where id = (select id from customer where username = username) and type = 'S')
		var savingBalance float64
		db.Model(model.Saving{}).Select("balance").Where("number in (?)", db.Model(model.Account{}).Select("number").Where("id = (?) AND type = 'S'", db.Model(model.Customer{}).Select("id").Where("username = ?", username))).Find(&savingBalance)
		accountBalance = append(accountBalance, savingBalance)
		c.JSON(http.StatusOK, gin.H{
			"data": accountBalance,
		})
	})

	_ = r.Run(":8080")

}
