package app

import (
	"github.com/gofiber/fiber/v2"
)

type Request struct {
	Path    string
	Method  string
	Headers map[string]string
	Body    string
}

func RequestFromFiber(r *fiber.Request) Request {
	req := Request{
		Path:    string(r.URI().RequestURI()),
		Body:    string(r.Body()),
		Method:  string(r.Header.Method()),
		Headers: make(map[string]string),
	}
	r.Header.VisitAll(
		func(key, value []byte) {
			req.Headers[string(key)] = string(value)
		},
	)
	return req
}

type Handler struct {
	matcher *Matcher
}

func NewHandler(matcher *Matcher) *Handler {
	return &Handler{matcher}
}

func (h Handler) All(c *fiber.Ctx) error {
	return c.JSON(RequestFromFiber(c.Request()))
}
