// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

const TableNameStudentLoan = "student_loan"

// StudentLoan mapped from table <student_loan>
type StudentLoan struct {
	Number         int64  `gorm:"column:number;primaryKey" json:"number"`
	UniversityID   int64  `gorm:"column:university_id;not null" json:"university_id"`
	StudentID      string `gorm:"column:student_id;not null" json:"student_id"`
	EducationType  string `gorm:"column:education_type;not null" json:"education_type"`
	ExpectGradDate string `gorm:"column:expect_grad_date;not null" json:"expect_grad_date"`
}

// TableName StudentLoan's table name
func (*StudentLoan) TableName() string {
	return TableNameStudentLoan
}
