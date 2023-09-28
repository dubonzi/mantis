package app

import (
	"strings"

	"github.com/americanas-go/log"
	"github.com/gofiber/fiber/v2"
	"github.com/ohler55/ojg/oj"
)

type ServiceMatcher interface {
	MatchRequest(Request) MatchResult
}

type Request struct {
	Path    string            `json:"path"`
	Method  string            `json:"method"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
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
			req.Headers[strings.ToLower(string(key))] = string(value)
		},
	)
	return req
}

type Handler struct {
	service ServiceMatcher
}

func NewHandler(service ServiceMatcher) *Handler {
	return &Handler{service}
}

func (h Handler) All(c *fiber.Ctx) error {
	req := RequestFromFiber(c.Request())
	res := h.service.MatchRequest(req)

	if !res.Matched {
		log.WithFields(log.Fields{
			"request": req,
			"result":  res,
		}).Warn("no match found")
	}

	for k, v := range res.Headers {
		c.Response().Header.Add(k, v)
	}

	c.Status(res.StatusCode)

	if res.Body != nil {
		switch b := res.Body.(type) {
		case string:
			return c.SendString(b)
		default:
			return c.SendString(oj.JSON(b))
		}
	}

	return c.Send(nil)
}
