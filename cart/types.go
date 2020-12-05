package cart

import "time"

type Item struct {
	ID        string
	CartID    string
	Quantity  int
	UnitPrice float64
}

type Cart struct {
	ID         string
	CustomerID string
	Items      []Item
}

type HealthCheck struct {
	Service string
	Status  string
	Date    time.Time
}
