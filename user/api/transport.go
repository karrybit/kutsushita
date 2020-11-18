package api

import (
	"encoding/json"
	"errors"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

var ErrInvalidRequest = errors.New("Invalid request")

func MakeHTTPHandler(service Service, logger *zap.Logger /*, tracer */) *fiber.App {
	app := fiber.New()
	app.Post("/register", register(service))
	app.Get("/customers", register(service))
	app.Get("/addresses", register(service))
	app.Get("/cards", register(service))
	app.Post("/customers", register(service))
	app.Post("/addresses", register(service))
	app.Post("/cards", register(service))
	app.Delete("/", register(service))
	app.Get("/health", register(service))
	return app
}

func register(service Service) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		registerRequest := new(registerRequest)
		if err := json.Unmarshal(c.Body(), &registerRequest); err != nil {
			return err
		}
		id, err := service.Register(ctx, registerRequest.Username, registerRequest.Password, registerRequest.Email, registerRequest.FirstName, registerRequest.LastName)
		if err != nil {
			return err
		}
		postResponse := postResponse{ID: id}
		b, err := json.Marshal(postResponse)
		if err != nil {
			return err
		}
		return c.Send(b)
	}
}
