package payment

import (
	"github.com/gofiber/fiber/v2"
)

func MakeHTTPHandler(service Service) *fiber.App {
	app := fiber.New()
	app.Post("/paymentauth", auth(service))
	app.Get("/health", health(service))
	return app
}

func auth(service Service) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		_ = c.Context()
		amount, err := decodeAuthoriseRequest(c)
		if err != nil {
			return err
		}

		authorisation, err := service.Authorise(amount)
		if err != nil {
			return err
		}

		return c.JSON(authoriseResponse{authorisation, err})
	}
}

func decodeAuthoriseRequest(c *fiber.Ctx) (float32, error) {
	var amount float32
	if err := c.BodyParser(&amount); err != nil {
		return 0.0, err
	}
	if amount == 0.0 {
		return 0.0, &UnmarshalKeyError{Key: "amount", JSON: string(c.Body())}
	}

	return amount, nil
}

func health(service Service) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		_ = c.Context()
		health := service.Health()
		return c.JSON(healthResponse{health})
	}
}
