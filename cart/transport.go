package cart

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
)

func MakeHTTPHandler(service Service) *fiber.App {
	app := fiber.New()
	carts := app.Group("/carts/:customerID")
	carts.Get("/", getCart(service))
	carts.Delete("/", getCart(service))
	carts.Get("/merge", mergeCart(service))
	items := carts.Group("/items")
	items.Get("/:itemID", getItem(service))
	items.Get("/", getItems(service))
	items.Post("/", createItem(service))
	items.Delete("/:itemID", deleteItem(service))
	items.Patch("/", updateItem(service))
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
		ctx := c.Context()
		customerID := c.Params("customerID")
		return service.DeleteCart(ctx, customerID)
	}
}

func mergeCart(service Service) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		customerID := c.Params("customerID")
		sessionID := c.Query("sessionId")
		return service.MargeCart(ctx, customerID, sessionID)
	}
}

func getItem(service Service) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		customerID := c.Params("customerID")
		itemID := c.Params("itemID")
		item, err := service.GetItem(ctx, customerID, itemID)
		if err != nil {
			return err
		}
		return c.JSON(item)
	}
}

func getItems(service Service) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		customerID := c.Params("customerID")
		items, err := service.GetItems(ctx, customerID)
		if err != nil {
			return err
		}
		return c.JSON(items)
	}
}

func createItem(service Service) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		customerID := c.Params("customerID")
		requestItem := new(Item)
		if err := json.Unmarshal(c.Body(), requestItem); err != nil {
			return err
		}
		return service.CreateItem(ctx, customerID, requestItem)
	}
}

func deleteItem(service Service) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		customerID := c.Params("customerID")
		itemID := c.Params("itemID")
		return service.DeleteItem(ctx, customerID, itemID)
	}
}

func updateItem(service Service) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		customerID := c.Params("customerID")
		requestItem := new(Item)
		if err := json.Unmarshal(c.Body(), requestItem); err != nil {
			return err
		}
		return service.UpdateItem(ctx, customerID, requestItem)
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
