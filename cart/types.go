package cart

import "time"

type Item struct {
	ID        string  `json:"id"`
	CartID    string  `json:"cartID"`
	Quantity  int     `json:"quantity"`
	UnitPrice float64 `json:"unitPrice"`
}

type Cart struct {
	ID         string `json:"id"`
	CustomerID string `json:"customerID"`
	Items      []Item `json:"items"`
}

type HealthCheck struct {
	Service string    `json:"service"`
	Status  string    `json:"status"`
	Date    time.Time `json:"date"`
}
