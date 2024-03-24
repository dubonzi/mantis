package app

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/ohler55/ojg/alt"
	"github.com/ohler55/ojg/oj"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			assert.Equal(t, tt.want.Method, got.Method)
			assert.Equal(t, tt.want.Path, got.Path)
			assert.Equal(t, tt.want.Headers, got.Headers)
			assert.Equal(t, tt.want.Body, got.Body)
			_, err := uuid.Parse(got.ID)
			assert.NoError(t, err)
			_, err = time.Parse(time.RFC3339Nano, got.Date)
			assert.NoError(t, err)
		})
	}

}

type mockService struct {
	mockMatchFunc func(context.Context, Request) MatchResult
}

func (m mockService) MatchRequest(ctx context.Context, r Request) MatchResult {
	return m.mockMatchFunc(ctx, r)
}

func TestHandleResponse(t *testing.T) {

	tests := []struct {
		name       string
		matchFunc  func(context.Context, Request) MatchResult
		assertFunc func(*testing.T, *http.Response)
	}{
		{
			name: "Should send not found response",
			matchFunc: func(ctx context.Context, r Request) MatchResult {
				return MatchResult{
					StatusCode: http.StatusNotFound,
					Headers:    map[string]string{"Content-type": "application/json"},
					Body:       buildNotFoundResponse(Request{Path: "/test", Method: "GET"}, nil),
				}
			},
			assertFunc: func(t *testing.T, r *http.Response) {
				body, err := oj.Load(r.Body)
				require.NoError(t, err)
				var nf NotFoundResponse
				_, err = alt.Recompose(body, &nf)
				require.NoError(t, err)
				assert.Equal(t, http.StatusNotFound, r.StatusCode)
				assert.Equal(t, "application/json", r.Header.Get("Content-type"))
				assert.Equal(t, buildNotFoundResponse(Request{Path: "/test", Method: "GET", Headers: make(map[string]string)}, nil), nf)
			},
		},
		{
			name: "Should send non json response",
			matchFunc: func(ctx context.Context, r Request) MatchResult {
				return MatchResult{
					StatusCode: http.StatusOK,
					Headers:    map[string]string{"Content-type": "application/xml"},
					Body:       "<name>Bilbo</name>",
				}
			},
			assertFunc: func(t *testing.T, r *http.Response) {
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				assert.Equal(t, http.StatusOK, r.StatusCode)
				assert.Equal(t, "application/xml", r.Header.Get("Content-type"))
				assert.Equal(t, "<name>Bilbo</name>", string(body))
			},
		},
		{
			name: "Should send response without body",
			matchFunc: func(ctx context.Context, r Request) MatchResult {
				return MatchResult{
					StatusCode: http.StatusCreated,
					Headers:    map[string]string{"Location": "/users/123"},
				}
			},
			assertFunc: func(t *testing.T, r *http.Response) {
				assert.Equal(t, http.StatusCreated, r.StatusCode)
				assert.Equal(t, "/users/123", r.Header.Get("Location"))
			},
		},
	}

	for _, tt := range tests {
		app := fiber.New()
		hand := NewHandler(mockService{tt.matchFunc})
		app.All("/", hand.All)

		res, err := app.Test(httptest.NewRequest("GET", "/", nil))
		require.NoError(t, err)

		tt.assertFunc(t, res)

	}

}

func TestHealth(t *testing.T) {
	app := fiber.New()
	hand := NewHandler(nil)
	app.Get("/health", hand.Health)

	res, err := app.Test(httptest.NewRequest("GET", "/health", nil))
	require.NoError(t, err)

	r, err := io.ReadAll(res.Body)
	require.NoError(t, err)

	assert.Equal(t, `{"status": "ok"}`, string(r))
}
