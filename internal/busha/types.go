package busha

import "time"

type PayOutType string

const (
	PayOutAddress PayOutType = "address"
)

type CreateQuoteRequest struct {
	SourceCurrency	string	`json:"source_currency"`
	TargetCurrency	string	`json:"target_currency"`
	SourceAmount	string	`json:"source_amount,omitempty"`
	TargetAmount	string	`json:"target_amount,omitempty"`
	PayOut		*PayOut	`json:"pay_out,omitempty"`
}

type PayOut struct {
	Type	PayOutType	`json:"type"`
	Address	string		`json:"address,omitempty"`
	Network	string		`json:"network,omitempty"`
}

type Quote struct {
	ID		string		`json:"id"`
	SourceCurrency	string		`json:"source_currency"`
	TargetCurrency	string		`json:"target_currency"`
	SourceAmount	string		`json:"source_amount"`
	TargetAmount	string		`json:"target_amount"`
	Status		string		`json:"status"`
	ProfileID	string		`json:"profile_id"`
	ExpiresAt	time.Time	`json:"expires_at"`
	CreatedAt	time.Time	`json:"created_at"`
	UpdatedAt	time.Time	`json:"updated_at"`
}

type QuoteResponse struct {
	Status	string	`json:"status"`
	Message	string	`json:"message"`
	Data	Quote	`json:"data"`
}

type Transfer struct {
	ID		string		`json:"id"`
	QuoteID		string		`json:"quote_id"`
	SourceCurrency	string		`json:"source_currency"`
	TargetCurrency	string		`json:"target_currency"`
	SourceAmount	string		`json:"source_amount"`
	TargetAmount	string		`json:"target_amount"`
	Status		string		`json:"status"`
	ProfileID	string		`json:"profile_id"`
	CreatedAt	time.Time	`json:"created_at"`
	UpdatedAt	time.Time	`json:"updated_at"`
}

type TransferResponse struct {
	Status	string		`json:"status"`
	Message	string		`json:"message"`
	Data	Transfer	`json:"data"`
}

type ErrorDetail struct {
	Name	string	`json:"name"`
	Message	string	`json:"message"`
}

type APIError struct {
	Error	ErrorDetail		`json:"error"`
	Fields	map[string]interface{}	`json:"fields,omitempty"`
}
