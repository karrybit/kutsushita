package cart

import (
	"context"
	"time"
)

type Service interface {
	GetCart(ctx context.Context, customerID string) (*Cart, error)
	DeleteCart(ctx context.Context, customerID string) error
	MargeCart(ctx context.Context, customerID string, sessionID string) error
	Ping(ctx context.Context) []HealthCheck
}

type service struct{}

func (s *service) GetCart(ctx context.Context, customerID string) (*Cart, error) {
	return db.GetCart(ctx, customerID)
}

func (s *service) DeleteCart(ctx context.Context, customerID string) error {
	return db.DeleteCart(ctx, customerID)
}

func (s *service) MargeCart(ctx context.Context, customerID string, sessionID string) error {
	return db.MargeCart(ctx, customerID, sessionID)
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
