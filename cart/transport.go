package cart

import "github.com/gofiber/fiber/v2"

func MakeHTTPHandler(service Service) *fiber.App {
	app := fiber.New()
	app.Get("/carts/:customerID", getCart(service))
	app.Delete("/carts/:customerID", getCart(service))
	app.Get("/carts/:customerID/merge", mergeCart(service))
	app.Get("/health", health(service))
	return app
}

func getCart(service Service) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		customerID := c.Params("customerID")
		carts, err := service.GetCart(ctx, customerID)
		if err != nil {
			return err
		}
		return c.JSON(carts)
	}
}

func deleteCart(service Service) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		_ = c.Context()
		customerID := c.Params("customerID")
		return nil
	}
}

func mergeCart(service Service) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		_ = c.Context()
		customerID := c.Params("customerID")
		return nil
	}
}

func health(service Service) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		health := service.Ping(ctx)
		response := struct {
			Health []HealthCheck `json:"health"`
		}{
			Health: health,
		}
		return c.JSON(response)
	}
}
