// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package model

const TableNameHomeLoan = "home_loan"

// HomeLoan mapped from table <home_loan>
type HomeLoan struct {
	Number                  int64   `gorm:"column:number;primaryKey" json:"number"`
	HouseBuiltYear          string  `gorm:"column:house_built_year;not null" json:"house_built_year"`
	InsuranceAccNo          int64   `gorm:"column:insurance_acc_no;not null" json:"insurance_acc_no"`
	InsuranceCompanyName    string  `gorm:"column:insurance_company_name;not null" json:"insurance_company_name"`
	InsuranceCompanyState   string  `gorm:"column:insurance_company_state;not null" json:"insurance_company_state"`
	InsuranceCompanyCity    string  `gorm:"column:insurance_company_city;not null" json:"insurance_company_city"`
	InsuranceCompanyAddress string  `gorm:"column:insurance_company_address;not null" json:"insurance_company_address"`
	InsuranceCompanyZip     string  `gorm:"column:insurance_company_zip;not null" json:"insurance_company_zip"`
	InsuranceCompanyPremium float64 `gorm:"column:insurance_company_premium;not null" json:"insurance_company_premium"`
}

// TableName HomeLoan's table name
func (*HomeLoan) TableName() string {
	return TableNameHomeLoan
}
