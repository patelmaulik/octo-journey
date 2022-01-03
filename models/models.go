package models

type AuditTrail struct {
	ServedBy string `json:"servedBy"`
}

// Account type
type Account struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	AuditTrail
}
