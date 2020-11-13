package catalogue

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"
)

type fiberHandlerFunc func(c *fiber.Ctx) error

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

func list(service Service) fiberHandlerFunc {
	return func(c *fiber.Ctx) error {
		_ = c.Context()
		req, _ := decodeListRequest(c)
		socks, err := service.List(req.Tags, req.Order, req.PageNum, req.PageSize)
		resp := listResponse{socks, err}
		b, _ := json.Marshal(resp)
		return c.Send(b)
	}
}

func size(service Service) fiberHandlerFunc {
	return func(c *fiber.Ctx) error {
		_ = c.Context()
		req, _ := decodeCountRequest(c)
		n, err := service.Count(req.Tags)
		resp := countResponse{n, err}
		b, _ := json.Marshal(resp)
		return c.Send(b)
	}
}

func id(service Service) fiberHandlerFunc {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		id, ok := ctx.Value("id").(string)
		if !ok {
			c.Context().NotFound()
			return nil
		}
		sock, err := service.Get(id)
		resp := getResponse{sock, err}
		b, _ := json.Marshal(resp)
		return c.Send(b)
	}
}

func tags(service Service) fiberHandlerFunc {
	return func(c *fiber.Ctx) error {
		_ = c.Context()
		tags, err := service.Tags()
		resp := tagsResponse{Tags: tags, Err: err}
		b, _ := json.Marshal(resp)
		return c.Send(b)
	}
}

func health(service Service) fiberHandlerFunc {
	return func(c *fiber.Ctx) error {
		_ = c.Context()
		health := service.Health()
		resp := healthResponse{health}
		b, _ := json.Marshal(resp)
		return c.Send(b)
	}
}
