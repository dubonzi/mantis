package app

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
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
			assert.Empty(t, got.ID)
			assert.Empty(t, got.Date)
		})
	}

}

type mockService struct {
	mockMatchFunc func(Request) MatchResult
}

func (m mockService) MatchRequest(r Request) MatchResult {
	return m.mockMatchFunc(r)
}

func TestHandleResponse(t *testing.T) {

	tests := []struct {
		name       string
		matchFunc  func(Request) MatchResult
		assertFunc func(*testing.T, *http.Response)
	}{
		{
			name: "Should send not found response",
			matchFunc: func(r Request) MatchResult {
				return MatchResult{
					StatusCode: http.StatusNotFound,
					Headers:    map[string]string{"Content-type": "application/json"},
					Body:       buildNotFoundResponse(Request{Path: "/test", Method: "GET"}, nil),
				}
			},
			assertFunc: func(t *testing.T, r *http.Response) {
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				var nf NotFoundResponse
				err = json.Unmarshal(body, &nf)
				require.NoError(t, err)
				assert.NotContains(t, string(body), `"closestMapping"`)
				assert.Equal(t, http.StatusNotFound, r.StatusCode)
				assert.Equal(t, "application/json", r.Header.Get("Content-type"))
				_, err = uuid.Parse(nf.Request.ID)
				require.NoError(t, err)
				_, err = time.Parse(time.RFC3339Nano, nf.Request.Date)
				require.NoError(t, err)
				nf.Request.ID, nf.Request.Date = "", ""
				assert.Equal(t, buildNotFoundResponse(Request{Path: "/test", Method: "GET", Headers: nil}, nil), nf)
			},
		},
		{
			name: "Should send non json response",
			matchFunc: func(r Request) MatchResult {
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
			matchFunc: func(r Request) MatchResult {
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

func TestRequestCopiesFiberBuffers(t *testing.T) {
	var input fiber.Request
	input.Header.SetMethod("POST")
	input.SetRequestURI("/original")
	input.Header.Set("X-Token", "original")
	input.SetBodyString("original")
	request := RequestFromFiber(&input)
	input.Body()[0] = 'x'
	input.URI().RequestURI()[1] = 'x'
	input.Header.Method()[0] = 'X'
	input.Header.Peek("X-Token")[0] = 'x'
	require.Equal(t, "original", request.Body)
	require.Equal(t, "/original", request.Path)
	require.Equal(t, "POST", request.Method)
	require.Equal(t, "original", request.Headers["x-token"])
}

func TestMatchedHandlerDoesNotGenerateRequestMetadata(t *testing.T) {
	app := fiber.New()
	handler := NewHandler(mockService{func(request Request) MatchResult {
		require.Empty(t, request.ID)
		require.Empty(t, request.Date)
		return MatchResult{Matched: true, StatusCode: http.StatusOK, Body: "matched"}
	}})
	app.All("/*", handler.All)
	response, err := app.Test(httptest.NewRequest("GET", "/matched", nil))
	require.NoError(t, err)
	defer response.Body.Close()
	require.Equal(t, http.StatusOK, response.StatusCode)
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.Equal(t, "matched", string(body))
}

func TestNotFoundClosestMappingJSON(t *testing.T) {
	tests := []struct {
		name string
		body BodyMatch
		want string
	}{
		{
			name: "Should omit absent body conditions",
			want: `{"method":"GET","path":{"pattern":["^/expected/<tag>&$"]},"headers":{"accept":{"exact":"application/json"}}}`,
		},
		{
			name: "Should omit allocated empty body conditions",
			body: BodyMatch{CommonMatch: CommonMatch{Contains: []string{}, Patterns: []string{}}, JsonPath: []string{}},
			want: `{"method":"GET","path":{"pattern":["^/expected/<tag>&$"]},"headers":{"accept":{"exact":"application/json"}}}`,
		},
		{
			name: "Should flatten embedded body conditions",
			body: BodyMatch{CommonMatch: CommonMatch{Contains: []string{"<tag>&"}, Patterns: []string{"<tag>&"}}, JsonPath: []string{"$.marker"}},
			want: `{"method":"GET","path":{"pattern":["^/expected/<tag>&$"]},"headers":{"accept":{"exact":"application/json"}},"body":{"contains":["<tag>&"],"pattern":["<tag>&"],"jsonPath":["$.marker"]}}`,
		},
		{
			name: "Should retain exact body conditions",
			body: BodyMatch{CommonMatch: CommonMatch{Exact: "<tag>&"}},
			want: `{"method":"GET","path":{"pattern":["^/expected/<tag>&$"]},"headers":{"accept":{"exact":"application/json"}},"body":{"exact":"<tag>&"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping := Mapping{Request: RequestMapping{
				Method:  "GET",
				Path:    CommonMatch{Patterns: []string{"^/expected/<tag>&$"}},
				Headers: map[string]CommonMatch{"accept": {Exact: "application/json"}},
				Body:    tt.body,
			}}
			regex, jsonPath := NewRegexCache(), NewJSONPathCache()
			require.NoError(t, regex.AddFromMapping(mapping))
			require.NoError(t, jsonPath.AddExpressions(mapping.Request.Body.JsonPath))
			mappings := make(Mappings)
			require.NoError(t, mappings.Put(mapping))
			matcher := NewMatcher(regex, jsonPath)
			service := NewService(mappings, matcher, NewScenarioHandler(matcher), NewResponseDelayer())
			app := fiber.New()
			app.All("/*", NewHandler(service).All)
			request := httptest.NewRequest("GET", "/missing", nil)
			request.Header.Set("Accept", "application/json")

			response, err := app.Test(request)
			require.NoError(t, err)
			defer response.Body.Close()
			require.Equal(t, http.StatusNotFound, response.StatusCode)
			assert.Equal(t, "application/json", response.Header.Get("Content-Type"))
			body, err := io.ReadAll(response.Body)
			require.NoError(t, err)
			var output struct {
				ClosestMapping json.RawMessage            `json:"closestMapping"`
				Request        map[string]json.RawMessage `json:"request"`
			}
			require.NoError(t, json.Unmarshal(body, &output))
			assert.JSONEq(t, tt.want, string(output.ClosestMapping))
			assert.NotContains(t, output.Request, "body")
			assert.Contains(t, string(body), "<tag>&")
			assert.NotContains(t, string(body), `\u003c`)
			assert.NotContains(t, string(body), `\u003e`)
			assert.NotContains(t, string(body), `\u0026`)
		})
	}
}
