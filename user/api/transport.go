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
	app.Get("/customers", getCustomers(service))
	app.Get("/addresses", getAddresses(service))
	app.Get("/cards", getCards(service))
	app.Post("/customers", postCustomers(service))
	app.Post("/addresses", postAddresses(service))
	app.Post("/cards", postCards(service))
	app.Delete("/", delete(service))
	app.Get("/health", health(service))
	return app
}

func register(service Service) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		registerRequest := new(registerRequest)
		if err := json.Unmarshal(c.Body(), registerRequest); err != nil {
			return err
		}
		id, err := service.Register(ctx, registerRequest.Username, registerRequest.Password, registerRequest.Email, registerRequest.FirstName, registerRequest.LastName)
		if err != nil {
			return err
		}
		return c.JSON(postResponse{ID: id})
	}
}

func getCustomers(service Service) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		req := new(GetRequest)
		if err := json.Unmarshal(c.Body(), req); err != nil {
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

func postCustomers(service Service) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		return nil
	}
}

func getAddresses(service Service) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		return nil
	}
}

func postAddresses(service Service) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		return nil
	}
}

func getCards(service Service) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		req := new(GetRequest)
		if err := json.Unmarshal(c.Body(), req); err != nil {
			return err
		}
		cards, err := service.GetCards(ctx, req.ID)
		if err != nil {
			return err
		}
		if req.ID == "" {
			return c.JSON(EmbedStruct{cardsResponse{Cards: cards}})
		}
		if len(cards) == 0 {
			return c.JSON(users.Card{})
		}
		return c.JSON(cards[0])
	}
}

func postCards(service Service) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		return nil
	}
}

func delete(service Service) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		return nil
	}
}

func health(service Service) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		return nil
	}
}

func methodNotAllowed(service Service) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		return fiber.ErrMethodNotAllowed
	}
}
