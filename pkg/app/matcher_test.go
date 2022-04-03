package app

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestMatcher(t *testing.T) {
	tests := []struct {
		name      string
		input     Request
		want      MatchResult
		wantMatch bool
	}{
		{
			name:      "Should match simple request",
			input:     Request{Method: "GET", Path: "/simple"},
			want:      MatchResult{StatusCode: 200, Matched: true, Headers: map[string]string{"content-type": "text/plain"}, Body: "I'm a simple response"},
			wantMatch: true,
		},
		{
			name:      "Should match GET request with header",
			input:     Request{Method: "GET", Path: "/bears/321", Headers: map[string]string{"authorization": "Bearer Bear 🐻"}},
			want:      MatchResult{StatusCode: 200, Matched: true, Headers: map[string]string{"content-type": "text/plain"}, Body: "🐻"},
			wantMatch: true,
		},
		{
			name:      "Should match GET request and load body from file",
			input:     Request{Method: "GET", Path: "/match/me/123?file=true"},
			want:      MatchResult{StatusCode: 200, Matched: true, Headers: map[string]string{"content-type": "application/json"}, Body: `{"message": "Hello from the body file"}`},
			wantMatch: true,
		},
		{
			name:      "Should match GET request if path contains request",
			input:     Request{Method: "GET", Path: "/thispath/contains/123"},
			want:      MatchResult{StatusCode: 200, Matched: true, Headers: map[string]string{"content-type": "text/plain"}, Body: `Mapping contains path`},
			wantMatch: true,
		},
		{
			name:      "Should match GET request if path matches regex",
			input:     Request{Method: "GET", Path: "/regex/2"},
			want:      MatchResult{StatusCode: 200, Matched: true, Headers: map[string]string{"content-type": "text/plain"}, Body: `Mapping with regex on path`},
			wantMatch: true,
		},
		{
			name:      "Should match POST request with body",
			input:     Request{Method: "POST", Path: "/order", Headers: map[string]string{"authorization": "Bearer ItsMe"}, Body: `{"cart": "555"}`},
			want:      MatchResult{StatusCode: 201, Matched: true, Headers: map[string]string{"location": "12345"}},
			wantMatch: true,
		},
		{
			name:      "Should match POST request if body and header contain request",
			input:     Request{Method: "POST", Path: "/bears/contains", Headers: map[string]string{"content-type": "application/json"}, Body: `{"name": "Mr Bear", "honey": true}`},
			want:      MatchResult{StatusCode: 201, Matched: true, Headers: map[string]string{"location": "12345"}},
			wantMatch: true,
		},
		{
			name:      "Should match PUT request if body matches single JSON path",
			input:     Request{Method: "PUT", Path: "/json/path", Body: `{"products": [{"id": "12345"}, {"id": "123452"}]}`},
			want:      MatchResult{StatusCode: 204, Matched: true, Headers: map[string]string{"multiple": "false"}},
			wantMatch: true,
		},
		{
			name:      "Should match PUT request if body matches multiple JSON paths",
			input:     Request{Method: "PUT", Path: "/json/path", Body: `{"products": [{"id": "12346"}], "users": [{"name": "Bob"}]}`},
			want:      MatchResult{StatusCode: 204, Matched: true, Headers: map[string]string{"multiple": "true"}},
			wantMatch: true,
		},
		{
			name:      "Should match POST request if body and header match regex",
			input:     Request{Method: "POST", Path: "/gopher/regex", Headers: map[string]string{"content-type": "application/json"}, Body: `{"name": "Mr Gopher", "honey": true}`},
			want:      MatchResult{StatusCode: 201, Matched: true, Headers: map[string]string{"location": "999"}},
			wantMatch: true,
		},
		{
			name:      "Should not match GET request if path does not match regex",
			input:     Request{Method: "GET", Path: "/regex/abc"},
			wantMatch: false,
		},
		{
			name:      "Should not match a request when the method is not mapped",
			input:     Request{Method: "HEAD", Path: "/gopher/regex", Headers: map[string]string{"content-type": "application/json"}, Body: `{"name": "Mr Gopher", "honey": true}`},
			wantMatch: false,
		},
		{
			name:  "Should return 404 with the closest mapping when no match is found",
			input: Request{Method: "GET", Path: "/bears/321"},
			want: MatchResult{
				Matched:    false,
				StatusCode: 404,
				Body: NotFoundResponse{
					Message: NoMappingFoundMessage,
					Request: Request{Method: "GET", Path: "/bears/321"},
					ClosestMapping: &RequestMapping{
						Method:  "GET",
						Path:    CommonMatch{Exact: "/bears/321"},
						Headers: map[string]CommonMatch{"authorization": {Exact: "Bearer Bear 🐻"}},
					},
				},
			},
			wantMatch: false,
		},
	}

	mappings := getMappings()
	regexCache := NewRegexCache()
	jsonPathCache := NewJSONPathCache()
	for _, method := range mappings {
		for _, mapping := range method {
			_ = regexCache.AddFromMapping(mapping)
			_ = jsonPathCache.AddExpressions(mapping.Request.Body.JsonPath)
		}
	}
	matcher := NewMatcher(mappings, regexCache, jsonPathCache)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping, matched := matcher.Match(tt.input)

			result := NewMatchResult(mapping, tt.input, matched)

			if matched != tt.wantMatch {
				t.Logf("Matching conditions differ: got '%t', want '%t'", matched, tt.wantMatch)
			}

			if tt.wantMatch && !assert.IsEqual(result, tt.want) {
				t.Logf("%s doest not equal %s", result, tt.want)
				t.FailNow()
			}
		})
	}
}

func getMappings() Mappings {
	return Mappings{
		"GET": []Mapping{
			{
				Request: RequestMapping{
					Method:  "GET",
					Path:    CommonMatch{Exact: "/bears/321"},
					Headers: map[string]CommonMatch{"authorization": {Exact: "Bearer Bear 🐻"}},
				},
				Response: ResponseMapping{StatusCode: 200, Headers: map[string]string{"content-type": "text/plain"}, Body: "🐻"},
			},
			{
				Request:  RequestMapping{Method: "GET", Path: CommonMatch{Exact: "/match/me/123?file=true"}},
				Response: ResponseMapping{StatusCode: 200, Headers: map[string]string{"content-type": "application/json"}, Body: `{"message": "Hello from the body file"}`},
			},
			{
				Request:  RequestMapping{Method: "GET", Path: CommonMatch{Exact: "/simple"}},
				Response: ResponseMapping{StatusCode: 200, Headers: map[string]string{"content-type": "text/plain"}, Body: "I'm a simple response"},
			},
			{
				Request:  RequestMapping{Method: "GET", Path: CommonMatch{Contains: []string{"contains/123"}}},
				Response: ResponseMapping{StatusCode: 200, Headers: map[string]string{"content-type": "text/plain"}, Body: "Mapping contains path"},
			},
			{
				Request:  RequestMapping{Method: "GET", Path: CommonMatch{Patterns: []string{"regex/[0-9]+$", `regex/\d{1}`}}},
				Response: ResponseMapping{StatusCode: 200, Headers: map[string]string{"content-type": "text/plain"}, Body: "Mapping with regex on path"},
			},
		},
		"POST": []Mapping{
			{
				Request: RequestMapping{
					Method:  "POST",
					Path:    CommonMatch{Exact: "/order"},
					Headers: map[string]CommonMatch{"Authorization": {Exact: "Bearer ItsMe"}},
					Body:    BodyMatch{CommonMatch: CommonMatch{Exact: `{"cart": "555"}`}},
				},
				Response: ResponseMapping{StatusCode: 201, Headers: map[string]string{"location": "12345"}},
			},
			{
				Request:  RequestMapping{Method: "POST", Path: CommonMatch{Exact: "/order"}, Body: BodyMatch{CommonMatch: CommonMatch{Exact: `{"cart": "555"}`}}},
				Response: ResponseMapping{StatusCode: 401},
			},
			{
				Request: RequestMapping{
					Method:  "POST",
					Headers: map[string]CommonMatch{"content-type": {Contains: []string{"json"}}},
					Path:    CommonMatch{Contains: []string{"bears", "contains"}},
					Body:    BodyMatch{CommonMatch: CommonMatch{Contains: []string{"name", `"honey": true`}}},
				},
				Response: ResponseMapping{StatusCode: 201, Headers: map[string]string{"location": "12345"}},
			},
			{
				Request: RequestMapping{
					Method:  "POST",
					Headers: map[string]CommonMatch{"content-type": {Patterns: []string{"^application/(json|xml){1}$"}}},
					Path:    CommonMatch{Exact: "/gopher/regex"},
					Body:    BodyMatch{CommonMatch: CommonMatch{Patterns: []string{`"name":\s*"[A-z\s]+"`}}},
				},
				Response: ResponseMapping{StatusCode: 201, Headers: map[string]string{"location": "999"}},
			},
		},
		"PUT": []Mapping{
			{
				Request:  RequestMapping{Method: "PUT", Path: CommonMatch{Exact: "/json/path"}, Body: BodyMatch{JsonPath: []string{"$.products[?(@.id == '12345')]"}}},
				Response: ResponseMapping{StatusCode: 204, Headers: map[string]string{"multiple": "false"}},
			},
			{
				Request: RequestMapping{
					Method: "PUT",
					Path:   CommonMatch{Exact: "/json/path"},
					Body:   BodyMatch{JsonPath: []string{"$.products[?(@.id == '12346')]", "$.users[?(@.name == 'Bob')]"}},
				},
				Response: ResponseMapping{StatusCode: 204, Headers: map[string]string{"multiple": "true"}},
			},
		},
		"DELETE": []Mapping{
			{
				Request:  RequestMapping{Method: "DELETE", Path: CommonMatch{Exact: "/cart/123"}},
				Response: ResponseMapping{StatusCode: 204},
			},
		},
	}
}
