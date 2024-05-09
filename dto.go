package main

type OpenAccountRequest struct {
	Username                string  `json:"username"`
	Fname                   string  `json:"fname"`
	Lname                   string  `json:"lname"`
	State                   string  `json:"state"`
	City                    string  `json:"city"`
	Zip                     string  `json:"zip"`
	Address                 string  `json:"address"`
	Type                    string  `json:"type"`
	Amount                  float64 `json:"amount"`
	Months                  int32   `json:"months"`
	LoanType                string  `json:"loanType"`
	HouseBuildYear          string  `json:"houseBuildYear"`
	InsuranceAccNo          int64   `json:"insuranceAccNo"`
	InsuranceCompanyName    string  `json:"insuranceCompanyName"`
	InsuranceCompanyState   string  `json:"insuranceCompanyState"`
	InsuranceCompanyCity    string  `json:"insuranceCompanyCity"`
	InsuranceCompanyZip     string  `json:"insuranceCompanyZip"`
	InsuranceCompanyAddress string  `json:"insuranceCompanyAddress"`
	InsuranceCompanyPremium float64 `json:"insuranceCompanyPremium"`
	UniversityName          string  `json:"universityName"`
	StudentID               string  `json:"studentID"`
	EducationType           string  `json:"educationType"`
	ExpectGradDate          string  `json:"expectGradDate"`
}
type Deposit struct {
	Balance float64 `json:"balance"` // 金额
	Account int64   `json:"account"` // 账户名或账户ID
}
