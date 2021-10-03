package handler

import "github.com/gofiber/fiber/v2"

func All(c *fiber.Ctx) error {
	return c.SendString("Hello")
}
