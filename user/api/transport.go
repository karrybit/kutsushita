package api

import (
	"encoding/json"
	"errors"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"

	"user/db"
	"user/users"
)

var ErrInvalidRequest = errors.New("Invalid request")

func MakeHTTPHandler(service Service, logger *zap.Logger /*, tracer */) *fiber.App {
	app := fiber.New()
	app.Post("/register", register(service))
	app.Get("/customers", customers(service))
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
		return c.JSON(postResponse{ID: id})
	}
}

func customers(service Service) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		req := new(GetRequest)
		if err := json.Unmarshal(c.Body(), &req); err != nil {
			return err
		}
		if req.ID == "" {
			return c.JSON(EmbedStruct{addressesResponse{Addresses: make([]users.Address, 0)}})
		}

		us, err := service.GetUsers(ctx, req.ID)
		if err != nil {
			return err
		}
		if len(us) == 0 {
			if req.Attr == "addresses" {
				return c.JSON(EmbedStruct{addressesResponse{Addresses: make([]users.Address, 0)}})
			}
			if req.Attr == "cards" {
				return c.JSON(EmbedStruct{addressesResponse{Addresses: make([]users.Address, 0)}})
			}
			return c.JSON(users.User{})
		}

		user := us[0]
		if err := db.GetUserAttributes(ctx, &user); err != nil {
			return err
		}
		if req.Attr == "address" {
			return c.JSON(EmbedStruct{addressesResponse{Addresses: user.Addresses}})
		}
		if req.Attr == "cards" {
			return c.JSON(EmbedStruct{cardsResponse{Cards: user.Cards}})
		}

		return c.JSON(user)
	}
}
