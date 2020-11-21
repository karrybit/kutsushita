package cart

import "time"

type Item struct {
	ID        string
	ItemID    string
	Quantity  int
	UnitPrice float64
}

type Cart struct {
	CustomerID string
	ID         string
	Items      []Item
}

type HealthCheck struct {
	Service string
	Status  string
	Date    time.Time
}
