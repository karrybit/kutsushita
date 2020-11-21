package order

import "time"

type Address struct {
	ID       string
	Number   string
	Street   string
	City     string
	Postcode string
	Country  string
}

type Card struct {
	ID      string
	LongNum string
	Expires string
	CCV     string
}

type Cart struct {
	CustomerID string
	ID         string
	Items      []Item
}

type Customer struct {
	ID        string
	FirstName string
	LastName  string
	UserName  string
	Addresses []Address
	Cards     []Card
}

type CustomerOrder struct {
	ID         string
	CustomerID string
	Customer   Customer
	Address    Address
	Card       Card
	Items      []Item
	Shipment   Shipment
	Date       time.Time
	Total      float64
}

type HealthCheck struct {
	Service string
	Status  string
	Date    time.Time
}

type Item struct {
	ID        string
	ItemID    string
	Quantity  int
	UnitPrice float64
}

type Shipment struct {
	ID   string
	Name string
}
