// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

const TableNameUser = "user"

// User mapped from table <user>
type User struct {
	Username string `gorm:"column:username;primaryKey" binding:"required,min=4,max=20" json:"username"`
	Password string `gorm:"column:password;not null" binding:"required,min=4,max=20" json:"password"`
	Role     string `gorm:"column:role;not null"  json:"role"`
	Keyword  string `gorm:"column:keyword;not null"  json:"keyword"`
}

// TableName User's table name
func (*User) TableName() string {
	return TableNameUser
}
