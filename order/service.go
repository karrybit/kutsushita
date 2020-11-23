package order

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type Service interface {
	Ping(ctx context.Context) *HealthCheckResponse
}

type service struct {
	logger *zap.Logger
}

func NewService(logger *zap.Logger) Service {
	return &service{
		logger: logger,
	}
}

func (s *service) Ping(ctx context.Context) *HealthCheckResponse {
	resp := new(HealthCheckResponse)
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

	resp.Health = append(resp.Health, app)
	resp.Health = append(resp.Health, database)

	return resp
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
