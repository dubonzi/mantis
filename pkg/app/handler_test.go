package app

import (
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestRequest(t *testing.T) {

	tests := []struct {
		name  string
		input *fiber.Request
		want  Request
	}{
		{
			name: "Should build POST request",
			input: func() *fiber.Request {
				r := &fiber.Request{}
				r.SetBodyString("{\"name\": \"gopher 1\"}")
				r.Header.Add("Content-Type", "application/json")
				r.Header.SetMethod("POST")
				r.SetRequestURI("/gopher")
				return r
			}(),
			want: Request{
				Method:  "POST",
				Path:    "/gopher",
				Headers: map[string]string{"content-type": "application/json"},
				Body:    "{\"name\": \"gopher 1\"}",
			},
		},
		{
			name: "Should build GET request",
			input: func() *fiber.Request {
				r := &fiber.Request{}
				r.Header.Add("Accept", "application/json")
				r.Header.SetMethod("GET")
				r.SetRequestURI("/gopher/2")
				return r
			}(),
			want: Request{
				Method:  "GET",
				Path:    "/gopher/2",
				Headers: map[string]string{"accept": "application/json"},
			},
		},
		{
			name: "Should build request with no headers",
			input: func() *fiber.Request {
				r := &fiber.Request{}
				r.Header.SetMethod("GET")
				r.SetRequestURI("/gopher/2")
				return r
			}(),
			want: Request{
				Method:  "GET",
				Path:    "/gopher/2",
				Headers: map[string]string{},
			},
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			got := RequestFromFiber(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}

}

func TestHandleResponse(t *testing.T) {
	tests := []struct {
		name      string
		matchFunc func(Request) MatchResult
		want      func(fiber)
	}{}
}
