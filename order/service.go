package order

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type Service interface {
	Ping(ctx context.Context) []HealthCheck
}

type service struct {
	logger *zap.Logger
}

func NewService(logger *zap.Logger) Service {
	return &service{
		logger: logger,
	}
}

func (s *service) Ping(ctx context.Context) []HealthCheck {
	now := time.Now()
	app := HealthCheck{
		Service: "orders",
		Status:  "OK",
		Date:    now,
	}

	database := HealthCheck{
		Service: "orders-db",
		Status:  "OK",
		Date:    now,
	}
	if err := db.Ping(ctx); err != nil {
		database.Status = "err"
	}

	return []HealthCheck{app, database}
}

func calculateTotal(items *[]Item) float64 {
	amount := 0.0
	for _, item := range *items {
		amount += float64(item.Quantity) * item.UnitPrice
	}
	shipping := 4.99
	amount += shipping
	return amount
}
