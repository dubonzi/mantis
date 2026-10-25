package app

import (
	"fmt"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
)

func prepareMatcher(tb testing.TB, entries []Mapping) (*Matcher, Mappings) {
	tb.Helper()
	regex := NewRegexCache()
	jsonPath := NewJSONPathCache()
	mappings := make(Mappings)
	for _, mapping := range entries {
		mapping.CalcMaxScoreAndCost()
		require.NoError(tb, regex.AddFromMapping(mapping))
		require.NoError(tb, jsonPath.AddExpressions(mapping.Request.Body.JsonPath))
		require.NoError(tb, mappings.Put(mapping))
	}
	return NewMatcher(regex, jsonPath), mappings
}

func TestMatcherLoadOrderAndPartialScores(t *testing.T) {
	exact := RequestMapping{Method: "POST", Path: CommonMatch{Exact: "/products/42"}}
	contains := RequestMapping{Method: "POST", Path: CommonMatch{Contains: []string{"products"}}}
	regex := RequestMapping{Method: "POST", Path: CommonMatch{Patterns: []string{`^/products/\d+$`}}}
	inactive := &ScenarioMapping{Name: "checkout", State: "ready"}
	tests := []struct {
		name     string
		entries  []Mapping
		request  Request
		wantKind MatchKind
		wantFile string
	}{
		{name: "earlier contains beats exact", entries: []Mapping{{Request: contains}, {Request: exact}}, wantKind: FullMatch, wantFile: "0"},
		{name: "earlier exact beats regex", entries: []Mapping{{Request: exact}, {Request: regex}}, wantKind: FullMatch, wantFile: "0"},
		{name: "earlier regex beats exact", entries: []Mapping{{Request: regex}, {Request: exact}}, wantKind: FullMatch, wantFile: "0"},
		{name: "duplicate exact uses first", entries: []Mapping{{Request: exact}, {Request: exact}}, wantKind: FullMatch, wantFile: "0"},
		{name: "inactive full score skipped for full match", entries: []Mapping{{Request: exact, Scenario: inactive}, {Request: contains}}, wantKind: FullMatch, wantFile: "1"},
		{name: "inactive full score skipped for partial", entries: []Mapping{
			{Request: exact, Scenario: inactive},
			{Request: RequestMapping{Method: "POST", Path: exact.Path, Body: BodyMatch{CommonMatch: CommonMatch{Exact: "different"}}}},
		}, wantKind: PartialMatch, wantFile: "1"},
		{name: "only inactive full score has no partial", entries: []Mapping{{Request: exact, Scenario: inactive}}, wantKind: NoMatch},
		{name: "inactive incomplete score still eligible", entries: []Mapping{
			{Request: RequestMapping{Method: "POST", Path: exact.Path, Body: BodyMatch{CommonMatch: CommonMatch{Exact: "different"}}}, Scenario: inactive},
		}, wantKind: PartialMatch, wantFile: "0"},
		{name: "highest partial includes wrong exact path", entries: []Mapping{
			{Request: RequestMapping{Method: "POST", Path: exact.Path, Body: BodyMatch{CommonMatch: CommonMatch{Exact: "different"}}}},
			{Request: RequestMapping{Method: "POST", Path: CommonMatch{Exact: "/elsewhere"}, Headers: map[string]CommonMatch{"Accept": {Contains: []string{"application", "json"}}}, Body: BodyMatch{CommonMatch: CommonMatch{Exact: "payload"}}}},
		}, wantKind: PartialMatch, wantFile: "1"},
		{name: "partial tie preserves order", entries: []Mapping{
			{Request: RequestMapping{Method: "POST", Path: CommonMatch{Exact: "/elsewhere"}, Body: BodyMatch{CommonMatch: CommonMatch{Exact: "payload"}}}},
			{Request: RequestMapping{Method: "POST", Path: exact.Path, Body: BodyMatch{CommonMatch: CommonMatch{Exact: "different"}}}},
		}, wantKind: PartialMatch, wantFile: "0"},
		{name: "components are all or nothing", entries: []Mapping{
			{Request: RequestMapping{Method: "POST", Path: CommonMatch{Contains: []string{"products", "absent"}}, Headers: map[string]CommonMatch{"accept": {Contains: []string{"application", "absent"}}}, Body: BodyMatch{CommonMatch: CommonMatch{Contains: []string{"pay", "absent"}}}}},
			{Request: RequestMapping{Method: "POST", Path: exact.Path, Body: BodyMatch{CommonMatch: CommonMatch{Exact: "different"}}}},
		}, wantKind: PartialMatch, wantFile: "1"},
		{name: "zero score is no match", entries: []Mapping{{Request: RequestMapping{Method: "POST", Path: CommonMatch{Exact: "/elsewhere"}}}}, wantKind: NoMatch},
		{name: "no mappings for method", entries: []Mapping{{Request: exact}}, request: Request{Method: "DELETE", Path: "/products/42"}, wantKind: NoMatch},
		{name: "empty mappings", wantKind: NoMatch},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := range tt.entries {
				tt.entries[i].FilePath = fmt.Sprint(i)
			}
			matcher, mappings := prepareMatcher(t, tt.entries)
			request := tt.request
			if request.Method == "" {
				request = Request{Method: "POST", Path: "/products/42", Headers: map[string]string{"accept": "application/json"}, Body: "payload"}
			}
			result := matcher.Match(request, mappings, map[string]ScenarioState{"checkout": {CurrentState: "waiting"}})
			require.Equal(t, tt.wantKind, result.Kind)
			require.Equal(t, tt.wantFile, result.Mapping.FilePath)
		})
	}
}

func benchmarkMappings() []Mapping {
	entries := make([]Mapping, 200)
	for i := range entries {
		request := RequestMapping{Method: "POST"}
		switch i % 5 {
		case 0, 1:
			request.Path.Exact = fmt.Sprintf("/products/%d", i)
		case 2:
			request.Path.Contains = []string{fmt.Sprintf("/catalog/%d/", i)}
		case 3:
			request.Path.Patterns = []string{fmt.Sprintf(`^/orders/%d/\d+$`, i)}
		case 4:
			request.Path.Exact = "/search"
			request.Body.JsonPath = []string{fmt.Sprintf("$.products[?(@.id == %d)]", i), "$.customer.name"}
		}
		if i%5 != 0 {
			request.Headers = map[string]CommonMatch{"Content-Type": {Exact: "application/json"}, "Authorization": {Patterns: []string{`^Bearer [a-z0-9]+$`}}}
		}
		if i%5 == 2 {
			request.Body.Patterns = []string{`"customer"\s*:`}
		}
		entries[i] = Mapping{Request: request, Response: ResponseMapping{StatusCode: 200, Body: `{"ok":true}`}, FilePath: fmt.Sprint(i)}
	}
	return entries
}

var benchmarkOutcome MatchOutcome
var benchmarkRequest Request

func BenchmarkMatcher(b *testing.B) {
	matcher, mappings := prepareMatcher(b, benchmarkMappings())
	for _, tt := range []struct {
		name string
		path string
		kind MatchKind
		file string
	}{
		{"Exact", "/products/190", FullMatch, "190"},
		{"Regex", "/orders/193/42", FullMatch, "193"},
		{"JSONPath", "/search", FullMatch, "194"},
		{"Miss", "/unmapped", PartialMatch, "194"},
	} {
		b.Run(tt.name, func(b *testing.B) {
			r := Request{Method: "POST", Path: tt.path, Headers: map[string]string{"content-type": "application/json", "authorization": "Bearer abc123"}, Body: `{"products":[{"id":194}],"customer":{"name":"Ada"}}`}
			result := matcher.Match(r, mappings, nil)
			require.Equal(b, tt.kind, result.Kind)
			require.Equal(b, tt.file, result.Mapping.FilePath)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				benchmarkOutcome = matcher.Match(r, mappings, nil)
			}
		})
	}
}

func BenchmarkRequestFromFiber(b *testing.B) {
	r := &fiber.Request{}
	r.Header.SetMethod("POST")
	r.SetRequestURI("/products/190")
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", "Bearer abc123")
	r.SetBodyString(`{"products":[{"id":194}],"customer":{"name":"Ada"}}`)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarkRequest = RequestFromFiber(r)
	}
}

func BenchmarkHandlerAll(b *testing.B) {
	matcher, mappings := prepareMatcher(b, benchmarkMappings())
	service := NewService(mappings, matcher, NewScenarioHandler(matcher), &mockDelayer{})
	app := fiber.New()
	app.All("/*", NewHandler(service).All)
	handler := app.Handler()
	var request fasthttp.Request
	request.Header.SetMethod("POST")
	request.SetRequestURI("/products/190")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer abc123")
	request.SetBodyString(`{"products":[{"id":194}],"customer":{"name":"Ada"}}`)
	var ctx fasthttp.RequestCtx
	ctx.Init(&request, nil, nil)
	handler(&ctx)
	require.Equal(b, 200, ctx.Response.StatusCode())
	require.Equal(b, `{"ok":true}`, string(ctx.Response.Body()))
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler(&ctx)
	}
}
