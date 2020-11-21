package catalogue

import (
	"github.com/gofiber/fiber/v2"
)

func MakeHTTPHandler(service Service, imagePath string) *fiber.App {
	app := fiber.New()
	catalogue := app.Group("/catalogue")
	catalogue.Get("/", list(service))
	catalogue.Get("/size", size(service))
	catalogue.Get("/:id", id(service))
	app.Get("/tags", tags(service))
	app.Get("/health", health(service))
	return app
}

func list(service Service) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		_ = c.Context()
		req, _ := decodeListRequest(c)
		socks, err := service.List(req.Tags, req.Order, req.PageNum, req.PageSize)
		return c.JSON(listResponse{socks, err})
	}
}

func size(service Service) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		_ = c.Context()
		req, _ := decodeCountRequest(c)
		n, err := service.Count(req.Tags)
		return c.JSON(countResponse{n, err})
	}
}

func id(service Service) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		id, ok := ctx.Value("id").(string)
		if !ok {
			c.Context().NotFound()
			return nil
		}
		sock, err := service.Get(id)
		return c.JSON(getResponse{sock, err})
	}
}

func tags(service Service) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		_ = c.Context()
		tags, err := service.Tags()
		return c.JSON(tagsResponse{Tags: tags, Err: err})
	}
}

func health(service Service) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		_ = c.Context()
		health := service.Health()
		return c.JSON(healthResponse{health})
	}
}
