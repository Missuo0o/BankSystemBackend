package main

import (
	"fmt"
	"net/http"
	"strings"
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

		if loginRequest.Username == "" || loginRequest.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid r1equest",
			})
			return
		}

		var user []model.User
		// Check loginRequest.Username and loginRequest.Password
		// select * from user where username = loginRequest.Username
		db.Model(model.User{}).Where("username = ?", loginRequest.Username).First(&user)
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
	r.POST("/user/register", func(c *gin.Context) {
		var registerRequest model.User

		if err := c.ShouldBindJSON(&registerRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid request",
			})
			return
		}

		if registerRequest.Username == "" || registerRequest.Password == "" || registerRequest.Keyword == "" {
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
	r.POST("/user/open", RoleAuthMiddleware("C"), func(c *gin.Context) {
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
		fmt.Print(openAccountRequest)

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

		if account.Fname == "" || account.Lname == "" || account.State == "" || account.City == "" || account.Zip == "" || account.Address == "" || (account.Type != "C" && account.Type != "S" && account.Type != "L") {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid request",
			})
			return
		}

		switch requestType := openAccountRequest.Type; requestType {
		case "C":
			checking.Charge = 10.00
			tx := db.Begin()
			if err := tx.Create(&account).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{
					"message": err.Error(),
				})
				return
			}
			if err := tx.Create(&checking).Error; err != nil {
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
			if err := tx1.Create(&account).Error; err != nil {
				tx1.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{
					"message": err.Error(),
				})
				return
			}
			if err := tx1.Create(&saving).Error; err != nil {
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
			if (openAccountRequest.LoanType != "STUDENT" && openAccountRequest.LoanType != "HOME" && openAccountRequest.LoanType != "PERSONAL") || loan.Amount <= 0 || loan.Months <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "Invalid request",
				})
				return
			}
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
				if openAccountRequest.UniversityName == "" || studentLoan.StudentID == "" || studentLoan.EducationType == "" || studentLoan.ExpectGradDate == "" {
					c.JSON(http.StatusBadRequest, gin.H{
						"message": "Invalid request",
					})
					return

				}
				loan.Type = "STUDENT"
				tx2 := db.Begin()
				if err := tx2.Create(&account).Error; err != nil {
					tx2.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": err.Error(),
					})
					return
				}
				if err := tx2.Create(&loan).Error; err != nil {
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
					if err := tx2.Create(&university).Error; err != nil {
						tx2.Rollback()
						c.JSON(http.StatusBadRequest, gin.H{
							"message": err.Error(),
						})
						return
					}
				}
				studentLoan.UniversityID = university.Id
				if err := tx2.Create(&studentLoan).Error; err != nil {
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
				if homeLoan.HouseBuildYear == "" || homeLoan.InsuranceAccNo == 0 || homeLoan.InsuranceCompanyName == "" || homeLoan.InsuranceCompanyState == "" || homeLoan.InsuranceCompanyCity == "" || homeLoan.InsuranceCompanyZip == "" || homeLoan.InsuranceCompanyAddress == "" || homeLoan.InsuranceCompanyPremium == 0 {

					c.JSON(http.StatusBadRequest, gin.H{
						"message": "Invalid request2",
					})
					return
				}
				loan.Type = "HOME"
				tx3 := db.Begin()
				if err := tx3.Create(&account).Error; err != nil {
					tx3.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": err.Error(),
					})
					return
				}
				if err := tx3.Create(&loan).Error; err != nil {
					tx3.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": err.Error(),
					})
					return
				}
				if err := tx3.Create(&homeLoan).Error; err != nil {
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
				if err := tx4.Create(&account).Error; err != nil {
					tx4.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": err.Error(),
					})
					return
				}
				if err := tx4.Create(&loan).Error; err != nil {
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
	r.PUT("/user/reset", func(c *gin.Context) {
		var resetRequest model.User

		err := c.ShouldBindJSON(&resetRequest)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid request",
			})
			return
		}

		if resetRequest.Username == "" || resetRequest.Password == "" || resetRequest.Keyword == "" {
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
	r.PUT("/user/transfer", RoleAuthMiddleware("C"), func(c *gin.Context) {
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
		if transferRequest.FromAccountNumber == 0 || transferRequest.ToAccountNumber == 0 || transferRequest.Amount == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid request",
			})
			return
		}
		// select type from account where number = transferRequest.FromAccount
		var fromAccountType string
		db.Model(model.Account{}).Select("type").Where("number = ?", transferRequest.FromAccountNumber).Find(&fromAccountType)
		// select number from account where id = (select id from customer where username = username) and (type = 'C' or type = 'S')
		var fromAccountNumber int64
		db.Model(model.Account{}).Select("number").Where("id = (?) AND (type = 'C' OR type = 'S')", db.Model(model.Customer{}).Select("id").Where("username = ?", username)).Find(&fromAccountNumber)
		if fromAccountType == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "FromAccount does not exist",
			})
		}
		// select type from account where number = transferRequest.ToAccount
		var toAccountType string
		db.Model(model.Account{}).Select("type").Where("number = ?", transferRequest.ToAccountNumber).Find(&toAccountType)
		if toAccountType == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "ToAccount does not exist",
			})
		}
		switch fromAccountType {
		case "C":
			switch toAccountType {
			case "C":
				tx := db.Begin()
				// update checking set balance = balance - transferRequest.Amount where number = transferRequest.FromAccount and
				// balance >= transferRequest.Amount
				result := tx.Model(model.Checking{}).Where("number = ? AND balance >= ?", transferRequest.FromAccountNumber,
					transferRequest.Amount).Update("balance", gorm.Expr("balance - ?", transferRequest.Amount))
				if result.RowsAffected == 0 {
					tx.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": "Insufficient balance",
					})
					return
				}
				// update checking set balance = balance + transferRequest.Amount where number = transferRequest.ToAccount
				if err := tx.Model(model.Checking{}).Where("number = ?", transferRequest.ToAccountNumber).Update("balance", gorm.Expr("balance + ?", transferRequest.Amount)).Error; err != nil {
					tx.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": "ToAccount does not exist",
					})
					return
				}
				// insert into transfer_history
				transferRequest.TransferDate = time.Now().Format("2006-01-02 15:04:05")
				transferRequest.AccountType = "C"
				if err := tx.Create(&transferRequest).Error; err != nil {
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
				result := tx1.Model(model.Checking{}).Where("number = ? AND balance >= ?", transferRequest.FromAccountNumber, transferRequest.Amount).Update("balance", gorm.Expr("balance - ?", transferRequest.Amount))
				if result.RowsAffected == 0 {
					tx1.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": "Insufficient balance",
					})
					return
				}
				// update saving set balance = balance + transferRequest.Amount where number = transferRequest.ToAccount
				if err := tx1.Model(model.Saving{}).Where("number = ?", transferRequest.ToAccountNumber).Update("balance", gorm.Expr("balance + ?", transferRequest.Amount)).Error; err != nil {
					tx1.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": "ToAccount does not exist",
					})
					return

				}
				// insert into transfer_history
				transferRequest.TransferDate = time.Now().Format("2006-01-02 15:04:05")
				transferRequest.AccountType = "S"
				if err := tx1.Create(&transferRequest).Error; err != nil {
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
				result := tx2.Model(model.Saving{}).Where("number = ? AND balance >= ?", transferRequest.FromAccountNumber, transferRequest.Amount).Update("balance", gorm.Expr("balance - ?", transferRequest.Amount))
				if result.RowsAffected == 0 {
					tx2.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": "Insufficient balance",
					})
					return
				}
				// update checking set balance = balance + transferRequest.Amount where number = transferRequest.ToAccount
				if err := tx2.Model(model.Checking{}).Where("number = ?", transferRequest.ToAccountNumber).Update("balance", gorm.Expr("balance + ?", transferRequest.Amount)).Error; err != nil {
					tx2.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": "ToAccount does not exist",
					})
					return
				}
				// insert into transfer_history
				transferRequest.TransferDate = time.Now().Format("2006-01-02 15:04:05")
				transferRequest.AccountType = "C"
				if err := tx2.Create(&transferRequest).Error; err != nil {
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
				result := tx3.Model(model.Saving{}).Where("number = ? AND balance >= ?", transferRequest.FromAccountNumber, transferRequest.Amount).Update("balance", gorm.Expr("balance - ?", transferRequest.Amount))
				if result.RowsAffected == 0 {
					tx3.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": "Insufficient balance",
					})
					return
				}
				// update saving set balance = balance + transferRequest.Amount where number = transferRequest.ToAccount
				if err := tx3.Model(model.Saving{}).Where("number = ?", transferRequest.ToAccountNumber).Update("balance", gorm.Expr("balance + ?", transferRequest.Amount)).Error; err != nil {
					tx3.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": "ToAccount does not exist",
					})
					return
				}
				// insert into transfer_history
				transferRequest.TransferDate = time.Now().Format("2006-01-02 15:04:05")
				transferRequest.AccountType = "S"
				if err := tx3.Create(&transferRequest).Error; err != nil {
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
	r.GET("/user/transfer", RoleAuthMiddleware("C"), func(c *gin.Context) {
		session := sessions.Default(c)
		username := session.Get("username")
		var transferHistory []model.TransferHistory
		// select * from transfer_history where from_account_number in (select number from account where (type = 'C' or type = 'S') and id = (select id from customer where username = username))
		db.Where("from_account_number in (?)", db.Model(model.Account{}).Select("number").Where("(type = 'C' OR type = 'S') AND id = (?)", db.Model(model.Customer{}).Select("id").Where("username = ?", username))).Find(&transferHistory)
		for i := range transferHistory {
			date, _ := time.Parse(time.RFC3339, transferHistory[i].TransferDate)
			transferHistory[i].TransferDate = date.Format("2006-01-02 15:04:05")
		}
		c.JSON(http.StatusOK, gin.H{
			"data": transferHistory,
		})
	})

	// Deposit API
	r.PUT("/user/deposit", RoleAuthMiddleware("C"), func(c *gin.Context) {
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

		if depositRequest.Account == 0 || depositRequest.Balance == 0 {
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
				"message": "Account does not exist",
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
	r.GET("/user/account", RoleAuthMiddleware("C"), func(c *gin.Context) {
		session := sessions.Default(c)
		username := session.Get("username")
		var accountNumber []NumberType
		// select number,type from account where id = (select id from customer where username = username)
		db.Model(model.Account{}).Select("number", "type").Where("id = (?)", db.Model(model.Customer{}).Select("id").Where("username = ?", username)).Find(&accountNumber)
		c.JSON(http.StatusOK, gin.H{
			"data": accountNumber,
		})
	})

	// Get AccountBalance API
	r.GET("/user/balance", RoleAuthMiddleware("C"), func(c *gin.Context) {
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

	// Get UserInfo API
	r.GET("/user/user", RoleAuthMiddleware("C"), func(c *gin.Context) {
		session := sessions.Default(c)
		username := session.Get("username")
		// select * form customer where username = username
		var customer model.Customer
		db.Where("username = ?", username).First(&customer)
		c.JSON(http.StatusOK, gin.H{
			"data": customer,
		})
	})

	// Update UserInfo API
	r.PUT("/user/user", RoleAuthMiddleware("C"), func(c *gin.Context) {
		session := sessions.Default(c)
		username := session.Get("username")

		var updateRequest model.Customer

		err := c.ShouldBindJSON(&updateRequest)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid request",
			})
			return
		}
		if updateRequest.Fname == "" || updateRequest.Lname == "" || updateRequest.State == "" || updateRequest.City == "" || updateRequest.Zip == "" || updateRequest.Address == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid request",
			})
			return
		}
		// select id from customer where username = username
		var id int64
		db.Model(model.Customer{}).Select("id").Where("username = ?", username).Find(&id)

		tx := db.Begin()
		// update customer set fname = updateRequest.Fname, lname = updateRequest.Lname, state = updateRequest.State, city = updateRequest.City, zip = updateRequest.Zip, address = updateRequest.Address where username = username
		if err := tx.Model(model.Customer{}).Where("username = ?", username).Omit("username").Omit("id").Updates(&updateRequest).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}
		// update account set fname = updateRequest.Fname, lname = updateRequest.Lname, state = updateRequest.State, city = updateRequest.City, zip = updateRequest.Zip, address = updateRequest.Address where id = id
		if err := tx.Model(model.Account{}).Where("id = ?", id).Omit("id").Updates(&updateRequest).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}
		tx.Commit()
		c.JSON(http.StatusOK, gin.H{
			"message": "Update successful",
		})
	})

	// Get LoanInfo API
	r.GET("/user/loan", RoleAuthMiddleware("C"), func(c *gin.Context) {
		session := sessions.Default(c)
		username := session.Get("username")

		var loanPaymentCount int64
		// select * from loan where number in (select number from account where id = (select id from customer where username = username))
		var loan []model.Loan
		db.Model(model.Loan{}).Where("number in (?)", db.Model(model.Account{}).Select("number").Where("id = (?)", db.Model(model.Customer{}).Select("id").Where("username = ?", username))).Find(&loan)
		if len(loan) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Loan does not exist",
			})
			return
		}

		// select count(*) from loan_payment where number = (select number from account where id = (select id from customer where username = username))
		db.Raw("select count(*) from loan_payment where number = (select number from account where  type = 'L' and id = (select id from customer where username = ?))", username).Scan(&loanPaymentCount)
		// select months from loan where number = (select number from account where id = (select id from customer where username = username))
		var months int64
		db.Raw("select months from loan where number = (select number from account where  type = 'L' and id = (select id from customer where username = ?))", username).Scan(&months)
		if loanPaymentCount == months {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Loan already paid off",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"data": loan,
		})

	})

	// Pay Loan API
	r.POST("/user/pay", RoleAuthMiddleware("C"), func(c *gin.Context) {
		session := sessions.Default(c)
		username := session.Get("username")

		var jsonMap map[string]interface{}

		err := c.ShouldBindJSON(&jsonMap)
		var accountNumber = int64(jsonMap["account"].(float64))
		if accountNumber == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid request",
			})
			return
		}

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}

		// select * from loan where number in (select number from account where id = (select id from customer where username = username))
		var loan []model.Loan
		db.Model(model.Loan{}).Where("number in (?)", db.Model(model.Account{}).Select("number").Where("id = (?)", db.Model(model.Customer{}).Select("id").Where("username = ?", username))).Find(&loan)
		if len(loan) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Loan does not exist",
			})
			return
		}

		// select * from loan_payment where number = (select number from account where type = 'L' and id = (select id from customer where username = username))
		var loanPayment []model.LoanPayment
		db.Model(model.LoanPayment{}).Where("number = (?)", db.Model(model.Account{}).Select("number").Where("type = 'L' AND id = (?)", db.Model(model.Customer{}).Select("id").Where("username = ?", username))).Find(&loanPayment)
		for _, payment := range loanPayment {

			var currentMonth bool
			currentMonth = strings.Contains(payment.PaymentDate, time.Now().Format("2006-01"))

			if currentMonth {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "Payment already made this month",
				})
				return
			}
		}

		// select payment from loan where number in  (select number from account where id = (select id from customer where username = username))
		var payment float64
		db.Model(model.Loan{}).Select("payment").Where("number in (?)", db.Model(model.Account{}).Select("number").Where("id = (?)", db.Model(model.Customer{}).Select("id").Where("username = ?", username))).Find(&payment)

		// select type from account where number = accountNumber
		var accountType string
		db.Model(model.Account{}).Select("type").Where("number = ?", accountNumber).Find(&accountType)

		// select number from account where id = (select id from customer where username = username) and (type = 'C' or type = 'S')
		var AccountNumber []int64
		db.Model(model.Account{}).Select("number").Where("id = (?) AND (type = 'C' OR type = 'S')", db.Model(model.Customer{}).Select("id").Where("username = ?", username)).Find(&AccountNumber)

		found := false

		for _, account := range AccountNumber {
			if account == accountNumber {
				found = true
				break
			}
		}

		if !found {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Account does not exist",
			})
		}

		var loanPaymentRecord model.LoanPayment
		loanPaymentRecord.Number = loan[0].Number
		loanPaymentRecord.PaymentDate = time.Now().Format("2006-01-02")
		loanPaymentRecord.PaymentAmount = payment
		switch accountType {
		case "C":
			// select balance from checking where number = accountNumber
			var balance float64
			db.Model(model.Checking{}).Select("balance").Where("number = ?", accountNumber).Find(&balance)
			if balance < payment {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "Insufficient balance",
				})
				return
			}
			tx := db.Begin()
			// update checking set balance = balance - payment where number = accountNumber
			if err := tx.Model(model.Checking{}).Where("number = ?", accountNumber).Update("balance", gorm.Expr("balance - ?", payment)).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{
					"message": err.Error(),
				})
				return
			}
			// insert into loan_payment
			if err := tx.Create(&loanPaymentRecord).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{
					"message": err.Error(),
				})
				return
			}
			tx.Commit()
			c.JSON(http.StatusOK, gin.H{
				"message": "Payment successful",
			})
		case "S":
			// select balance from saving where number = accountNumber
			var balance float64
			db.Model(model.Saving{}).Select("balance").Where("number = ?", accountNumber).Find(&balance)
			if balance < payment {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "Insufficient balance",
				})
				return
			}
			tx := db.Begin()
			// update saving set balance = balance - payment where number = accountNumber
			if err := tx.Model(model.Saving{}).Where("number = ?", accountNumber).Update("balance", gorm.Expr("balance - ?", payment)).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{
					"message": err.Error(),
				})
				return
			}
			// insert into loan_payment
			if err := tx.Create(&loanPaymentRecord).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{
					"message": err.Error(),
				})
				return
			}
			tx.Commit()

			c.JSON(http.StatusOK, gin.H{
				"message": "Payment successful",
			})
		}
	})

	// Admin register API
	r.POST("/admin/register", RoleAuthMiddleware("A"), func(c *gin.Context) {
		var registerRequest model.User

		if err := c.ShouldBindJSON(&registerRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid request",
			})
			return
		}

		if registerRequest.Username == "" || registerRequest.Password == "" || registerRequest.Role == "" || registerRequest.Keyword == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid request",
			})
			return
		}

		registerRequest.Password = hashPassword(registerRequest.Password)
		if registerRequest.Role != "A" && registerRequest.Role != "C" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid role",
			})
			return
		}
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

	// Admin OpenAccount API
	r.POST("/admin/open", RoleAuthMiddleware("A"), func(c *gin.Context) {
		var openAccountRequest OpenAccountRequest

		err := c.ShouldBindJSON(&openAccountRequest)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid request",
			})
			return
		}

		if openAccountRequest.Username == "" || openAccountRequest.Fname == "" || openAccountRequest.Lname == "" || openAccountRequest.State == "" || openAccountRequest.City == "" || openAccountRequest.Zip == "" || openAccountRequest.Address == "" || (openAccountRequest.Type != "C" && openAccountRequest.Type != "S" && openAccountRequest.Type != "L") {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid request",
			})
			return
		}
		var username = openAccountRequest.Username

		// select username from user where username = username
		var user model.User
		db.Where("username = ?", username).First(&user)
		if user.Username == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Username does not exist",
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
			customer.Username = username
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
			if err := tx.Create(&account).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{
					"message": err.Error(),
				})
				return
			}
			if err := tx.Create(&checking).Error; err != nil {
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
			if err := tx1.Create(&account).Error; err != nil {
				tx1.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{
					"message": err.Error(),
				})
				return
			}
			if err := tx1.Create(&saving).Error; err != nil {
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
			if (openAccountRequest.LoanType != "STUDENT" && openAccountRequest.LoanType != "HOME" && openAccountRequest.LoanType != "PERSONAL") || loan.Amount <= 0 || loan.Months <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": "Invalid request",
				})
				return
			}
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
				if openAccountRequest.UniversityName == "" || studentLoan.StudentID == "" || studentLoan.EducationType == "" || studentLoan.ExpectGradDate == "" {
					c.JSON(http.StatusBadRequest, gin.H{
						"message": "Invalid request2",
					})
					return

				}
				loan.Type = "STUDENT"
				tx2 := db.Begin()
				if err := tx2.Create(&account).Error; err != nil {
					tx2.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": err.Error(),
					})
					return
				}
				if err := tx2.Create(&loan).Error; err != nil {
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
					if err := tx2.Create(&university).Error; err != nil {
						tx2.Rollback()
						c.JSON(http.StatusBadRequest, gin.H{
							"message": err.Error(),
						})
						return
					}
				}
				studentLoan.UniversityID = university.Id
				if err := tx2.Create(&studentLoan).Error; err != nil {
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
				if homeLoan.HouseBuildYear == "" || homeLoan.InsuranceAccNo == 0 || homeLoan.InsuranceCompanyName == "" || homeLoan.InsuranceCompanyState == "" || homeLoan.InsuranceCompanyCity == "" || homeLoan.InsuranceCompanyZip == "" || homeLoan.InsuranceCompanyAddress == "" || homeLoan.InsuranceCompanyPremium == 0 {
					c.JSON(http.StatusBadRequest, gin.H{
						"message": "Invalid request3",
					})
					return
				}
				loan.Type = "HOME"
				tx3 := db.Begin()
				if err := tx3.Create(&account).Error; err != nil {
					tx3.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": err.Error(),
					})
					return
				}
				if err := tx3.Create(&loan).Error; err != nil {
					tx3.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": err.Error(),
					})
					return
				}
				if err := tx3.Create(&homeLoan).Error; err != nil {
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
				if err := tx4.Create(&account).Error; err != nil {
					tx4.Rollback()
					c.JSON(http.StatusBadRequest, gin.H{
						"message": err.Error(),
					})
					return
				}
				if err := tx4.Create(&loan).Error; err != nil {
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

	// Admin GetUserInfo APi
	r.GET("/admin/user", RoleAuthMiddleware("A"), func(c *gin.Context) {
		// select * from customer
		var customer []model.Customer
		db.Find(&customer)
		c.JSON(http.StatusOK, gin.H{
			"data": customer,
		})
	})

	// Admin Delete User API
	r.DELETE("/admin/user/:username", RoleAuthMiddleware("A"), func(c *gin.Context) {
		username := c.Param("username")
		// delete from user where username = username
		db.Where("username = ?", username).Delete(&model.User{})
		c.JSON(http.StatusOK, gin.H{
			"message": "Delete successful",
		})
	})

	// Admin GetUserInfoByUsername API
	r.GET("/admin/user/:username", RoleAuthMiddleware("A"), func(c *gin.Context) {
		username := c.Param("username")
		// select * from customer where username = username
		var customer model.Customer
		db.Where("username = ?", username).First(&customer)
		c.JSON(http.StatusOK, gin.H{
			"data": customer,
		})
	})

	// Admin UpdateUserInfo API
	r.PUT("/admin/user/:username", RoleAuthMiddleware("A"), func(c *gin.Context) {
		username := c.Param("username")
		var updateRequest model.Customer

		err := c.ShouldBindJSON(&updateRequest)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid request",
			})
			return
		}

		if updateRequest.Fname == "" || updateRequest.Lname == "" || updateRequest.State == "" || updateRequest.City == "" || updateRequest.Zip == "" || updateRequest.Address == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid request",
			})
			return
		}
		// select id from customer where username = username
		var id int64
		db.Model(model.Customer{}).Select("id").Where("username = ?", username).Find(&id)

		if id == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "User does not exist",
			})
			return
		}
		tx := db.Begin()
		// update customer set fname = updateRequest.Fname, lname = updateRequest.Lname, state = updateRequest.State, city = updateRequest.City, zip = updateRequest.Zip, address = updateRequest.Address where username = username
		if err := tx.Model(model.Customer{}).Where("username = ?", username).Omit("username").Omit("id").Updates(&updateRequest).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}
		// update account set fname = updateRequest.Fname, lname = updateRequest.Lname, state = updateRequest.State, city = updateRequest.City, zip = updateRequest.Zip, address = updateRequest.Address where id = id
		if err := tx.Model(model.Account{}).Where("id = ?", id).Omit("id").Updates(&updateRequest).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}
		tx.Commit()
		c.JSON(http.StatusOK, gin.H{
			"message": "Update successful",
		})
	})

	// Admin GetAccountByType API
	r.GET("/admin/account", RoleAuthMiddleware("A"), func(c *gin.Context) {
		var accountType string
		accountType = c.Query("type")
		if accountType != "S" && accountType != "C" && accountType != "L" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid request",
			})
			return
		}
		switch accountType {
		case "C":
			var checkingAccount []CheckingAccount
			// select account.*,checking.* from account,checking where type = 'C' and account.number = checking.number
			db.Raw("select account.*,checking.* from account,checking where type = 'C' and account.number = checking.number").Scan(&checkingAccount)
			for i := range checkingAccount {
				date, _ := time.Parse(time.RFC3339, checkingAccount[i].OpenDate)
				checkingAccount[i].OpenDate = date.Format("2006-01-02 15:04:05")
			}
			c.JSON(http.StatusOK, gin.H{
				"data": checkingAccount,
			})
		case "S":
			var savingAccount []SavingAccount
			// select account.*,saving.* from account,saving where type = 'S' and account.number = saving.number
			db.Raw("select account.*,savings.* from account,savings where type = 'S' and account.number = savings.number").Scan(&savingAccount)
			for i := range savingAccount {
				date, _ := time.Parse(time.RFC3339, savingAccount[i].OpenDate)
				savingAccount[i].OpenDate = date.Format("2006-01-02 15:04:05")
			}
			c.JSON(http.StatusOK, gin.H{
				"data": savingAccount,
			})
		case "L":
			var studentLoan []StudentLoanAccount
			var homeLoan []HomeLoanAccount
			var personalLoan []PersonalLoanAccount
			//  select account.*,loan.*,student_loan.*,university.name as university_name from account,loan,student_loan,university
			//                                                       where account.number = loan.number
			//                                                         and loan.number = student_loan.number and student_loan.university_id = university.id
			db.Raw("select account.*,loan.*,student_loan.*,university.name as university_name from account,loan,student_loan,university where account.number = loan.number and loan.number = student_loan.number and student_loan.university_id = university.id").Scan(&studentLoan)
			db.Raw("select account.*,loan.*,home_loan.* from account,loan,home_loan where account.number = loan.number and loan.number = home_loan.number").Scan(&homeLoan)
			db.Raw("select account.*,loan.* from account,loan where loan.type ='PERSONAL' and account.number = loan.number").Scan(&personalLoan)
			for i := range studentLoan {
				date, _ := time.Parse(time.RFC3339, studentLoan[i].OpenDate)
				studentLoan[i].OpenDate = date.Format("2006-01-02 15:04:05")
			}
			for i := range homeLoan {
				date, _ := time.Parse(time.RFC3339, homeLoan[i].OpenDate)
				homeLoan[i].OpenDate = date.Format("2006-01-02 15:04:05")
			}
			for i := range personalLoan {
				date, _ := time.Parse(time.RFC3339, personalLoan[i].OpenDate)
				personalLoan[i].OpenDate = date.Format("2006-01-02 15:04:05")
			}
			c.JSON(http.StatusOK, gin.H{
				"data": gin.H{
					"studentLoan":  studentLoan,
					"homeLoan":     homeLoan,
					"personalLoan": personalLoan,
				},
			})
		}
	})

	// Admin GetAccountByNumber API
	r.GET("/admin/account/:number", RoleAuthMiddleware("A"), func(c *gin.Context) {
		number := c.Param("number")
		// select type from account where number = number
		var accountType string
		db.Model(model.Account{}).Select("type").Where("number = ?", number).Find(&accountType)
		if accountType == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Account does not exist",
			})
			return
		}
		switch accountType {
		case "C":
			var checkingAccount CheckingAccount
			//SELECT account.*, checking.*
			//FROM account
			//JOIN checking ON account.number = checking.number
			//WHERE account.type = 'C' AND account.number = 1234567809;
			db.Raw("select account.*,checking.* from account,checking where type = 'C' and account.number = checking.number and account.number = ?", number).Scan(&checkingAccount)
			date, _ := time.Parse(time.RFC3339, checkingAccount.OpenDate)
			checkingAccount.OpenDate = date.Format("2006-01-02 15:04:05")

			c.JSON(http.StatusOK, gin.H{
				"data": checkingAccount,
			})
		case "S":
			var savingAccount SavingAccount
			//SELECT account.*, savings.*
			//FROM account
			//JOIN savings ON account.number = savings.number
			//WHERE account.type = 'C' AND savings.number = 1234567809;
			db.Raw("select account.*,savings.* from account,savings where type = 'S' and account.number = savings.number and account.number = ?", number).Scan(&savingAccount)
			date, _ := time.Parse(time.RFC3339, savingAccount.OpenDate)
			savingAccount.OpenDate = date.Format("2006-01-02 15:04:05")

			c.JSON(http.StatusOK, gin.H{
				"data": savingAccount,
			})
		case "L":
			var studentLoan StudentLoanAccount
			var homeLoan HomeLoanAccount
			var personalLoan PersonalLoanAccount
			//  select account.*,loan.*,student_loan.*,university.name as university_name from account,loan,student_loan,university
			//                                                       where account.number = loan.number
			//                                                         and loan.number = student_loan.number and student_loan.university_id = university.id and account.number = 1234567809
			db.Raw("select account.*,loan.*,student_loan.*,university.name as university_name from account,loan,student_loan,university where account.number = loan.number and loan.number = student_loan.number and student_loan.university_id = university.id and account.number = ?", number).Scan(&studentLoan)
			db.Raw("select account.*,loan.*,home_loan.* from account,loan,home_loan where account.number = loan.number and loan.number = home_loan.number and account.number = ?", number).Scan(&homeLoan)
			db.Raw("select account.*,loan.* from account,loan where loan.type ='PERSONAL' and account.number = loan.number and account.number = ?", number).Scan(&personalLoan)
			date, _ := time.Parse(time.RFC3339, studentLoan.OpenDate)
			studentLoan.OpenDate = date.Format("2006-01-02 15:04:05")
			date1, _ := time.Parse(time.RFC3339, homeLoan.OpenDate)
			homeLoan.OpenDate = date1.Format("2006-01-02 15:04:05")
			date2, _ := time.Parse(time.RFC3339, personalLoan.OpenDate)
			personalLoan.OpenDate = date2.Format("2006-01-02 15:04:05")

			c.JSON(http.StatusOK, gin.H{
				"data": gin.H{
					"studentLoan":  studentLoan,
					"homeLoan":     homeLoan,
					"personalLoan": personalLoan,
				},
			})
		}
	})

	// Admin DeleteAccount API
	r.DELETE("/admin/account/:number", RoleAuthMiddleware("A"), func(c *gin.Context) {
		number := c.Param("number")
		// select type from account where number = number
		var accountType string
		db.Model(model.Account{}).Select("type").Where("number = ?", number).Find(&accountType)
		if accountType == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Account does not exist",
			})
			return
		}
		// delete from account where number = number
		db.Where("number = ?", number).Delete(&model.Account{})
		c.JSON(http.StatusOK, gin.H{
			"message": "Delete successful",
		})
	})

	// Admin Update Account Balance APi
	r.PUT("/admin/balance", RoleAuthMiddleware("A"), func(c *gin.Context) {
		var updateRequest Deposit

		err := c.ShouldBindJSON(&updateRequest)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid request",
			})
			return
		}

		if updateRequest.Balance <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid request",
			})
			return
		}

		// select type from account where number = number
		var accountType string
		db.Model(model.Account{}).Select("type").Where("number = ?", updateRequest.Account).Find(&accountType)
		if accountType == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Account does not exist",
			})
			return
		}

		switch accountType {
		case "C":
			// update checking set balance = updateRequest.Balance where number = number
			if err := db.Model(model.Checking{}).Where("number = ?", updateRequest.Account).Update("balance", updateRequest.Balance).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"message": "Update successful",
			})
		case "S":
			// update saving set balance = updateRequest.Balance where number = number
			if err := db.Model(model.Saving{}).Where("number = ?", updateRequest.Account).Update("balance", updateRequest.Balance).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"message": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"message": "Update successful",
			})
		}
	})

	// Admin GetLoanSummary API
	r.GET("/admin/loan", RoleAuthMiddleware("A"), func(c *gin.Context) {
		var loanSummary []LoanSummary
		//SELECT
		//    l.number,
		//    l.type,
		//    l.amount as Original_Amount,
		//    (l.amount - IFNULL(SUM(lp.payment_amount), 0)) as Remaining_Amount,
		//    (l.months - COUNT(lp.id)) as Remaining_Months
		//FROM
		//    loan l
		//LEFT JOIN
		//    loan_payment lp ON l.number = lp.number
		//GROUP BY
		//    l.number;
		db.Raw("SELECT l.number, l.type, l.amount as original_amount, (l.amount - IFNULL(SUM(lp.payment_amount), 0)) as remaining_amount, (l.months - COUNT(lp.id)) as remaining_months FROM loan l LEFT JOIN loan_payment lp ON l.number = lp.number GROUP BY l.number").Scan(&loanSummary)
		c.JSON(http.StatusOK, gin.H{
			"data": loanSummary,
		})
	})

	_ = r.Run(":80")

}
