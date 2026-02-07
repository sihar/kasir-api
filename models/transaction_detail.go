package models

type TransactionDetail struct {
	ID				int		`json:"id"`
	TransactionID	int		`json:"transaction_id"`
	ProductID		int		`json:"product_id"`
	ProductName		string	`json:"product_name,omitempty"`
	Quantity		int		`json:"quantity"`
	Subtotal		int		`json:"subtotal"`
}