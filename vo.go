package main

type CheckingAccount struct {
	Number   int64   `json:"number"`
	Fname    string  `json:"fname"`
	Lname    string  `json:"lname"`
	State    string  `json:"state"`
	City     string  `json:"city"`
	Zip      string  `json:"zip"`
	OpenDate string  `json:"openDate"`
	Address  string  `json:"address"`
	Id       int64   `json:"id"`
	Balance  float64 `json:"balance"`
	Charge   float64 `json:"charges"`
	Type     string  `json:"type"`
}
type SavingAccount struct {
	Number   int64   `json:"number"`
	Fname    string  `json:"fname"`
	Lname    string  `json:"lname"`
	State    string  `json:"state"`
	City     string  `json:"city"`
	Zip      string  `json:"zip"`
	OpenDate string  `json:"openDate"`
	Address  string  `json:"address"`
	Id       int64   `json:"id"`
	Balance  float64 `json:"balance"`
	Rate     float64 `json:"rate"`
	Type     string  `json:"type"`
}
type StudentLoanAccount struct {
	Number         int64   `json:"number"`
	Fname          string  `json:"fname"`
	Lname          string  `json:"lname"`
	State          string  `json:"state"`
	City           string  `json:"city"`
	Zip            string  `json:"zip"`
	OpenDate       string  `json:"openDate"`
	Address        string  `json:"address"`
	Id             int64   `json:"id"`
	Rate           float64 `json:"rate"`
	Amount         float64 `json:"amount"`
	Months         int32   `json:"months"`
	Payment        float64 `json:"payment"`
	UniversityName string  `json:"universityName"`
	StudentID      string  `json:"studentID"`
	EducationType  string  `json:"educationType"`
	ExpectGradDate string  `json:"expectGradDate"`
}

type HomeLoanAccount struct {
	Number                  int64   `json:"number"`
	Fname                   string  `json:"fname"`
	Lname                   string  `json:"lname"`
	State                   string  `json:"state"`
	City                    string  `json:"city"`
	Zip                     string  `json:"zip"`
	OpenDate                string  `json:"openDate"`
	Address                 string  `json:"address"`
	Id                      int64   `json:"id"`
	Rate                    float64 `json:"rate"`
	Amount                  float64 `json:"amount"`
	Months                  int32   `json:"months"`
	Payment                 float64 `json:"payment"`
	HouseBuildYear          string  `json:"houseBuildYear"`
	InsuranceAccNo          int64   `json:"insuranceAccNo"`
	InsuranceCompanyName    string  `json:"insuranceCompanyName"`
	InsuranceCompanyState   string  `json:"insuranceCompanyState"`
	InsuranceCompanyCity    string  `json:"insuranceCompanyCity"`
	InsuranceCompanyZip     string  `json:"insuranceCompanyZip"`
	InsuranceCompanyAddress string  `json:"insuranceCompanyAddress"`
	InsuranceCompanyPremium float64 `json:"insuranceCompanyPremium"`
}

type PersonalLoanAccount struct {
	Number   int64   `json:"number"`
	Fname    string  `json:"fname"`
	Lname    string  `json:"lname"`
	State    string  `json:"state"`
	City     string  `json:"city"`
	Zip      string  `json:"zip"`
	OpenDate string  `json:"openDate"`
	Address  string  `json:"address"`
	Id       int64   `json:"id"`
	Rate     float64 `json:"rate"`
	Amount   float64 `json:"amount"`
	Months   int32   `json:"months"`
	Payment  float64 `json:"payment"`
}

type NumberType struct {
	Number int64  `json:"number"`
	Type   string `json:"type"`
}

type LoanSummary struct {
	Number          string  `json:"number"`           // 贷款编号
	Type            string  `json:"type"`             // 贷款类型
	OriginalAmount  float64 `json:"original_amount"`  // 原始金额
	RemainingAmount float64 `json:"remaining_amount"` // 剩余金额
	RemainingMonths int     `json:"remaining_months"` // 剩余月数
}
