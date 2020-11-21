package cart

import "github.com/gofiber/fiber/v2"

func MakeHTTPHandler() *fiber.App {
	app := fiber.New()
	return app
}
