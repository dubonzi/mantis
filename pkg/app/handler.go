package app

import "github.com/gofiber/fiber/v2"

type Handler struct {
	matcher *Matcher
}

func NewHandler(matcher *Matcher) *Handler {
	return &Handler{matcher}
}

func (h Handler) All(c *fiber.Ctx) error {
	return c.SendString("Hello")
}
