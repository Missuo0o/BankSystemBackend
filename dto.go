package main

type OpenAccountRequest struct {
	FName                   string  `json:"fName"`
	LName                   string  `json:"lName"`
	State                   string  `json:"state"`
	City                    string  `json:"city"`
	Zip                     string  `json:"zip"`
	Address                 string  `json:"address"`
	Type                    string  `json:"type"`
	Status                  bool    `json:"status"`
	Amount                  float64 `json:"amount"`
	Month                   int     `json:"month"`
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
