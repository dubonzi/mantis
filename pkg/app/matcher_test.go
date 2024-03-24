package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMatcher(t *testing.T) {
	tests := []struct {
		name        string
		input       Request
		want        MatchResult
		wantMatch   bool
		wantPartial bool
	}{
		{
			name:      "Should match simple request",
			input:     Request{Method: "GET", Path: "/simple"},
			want:      MatchResult{StatusCode: 200, Matched: true, Headers: map[string]string{"content-type": "text/plain", "X-Mapping-File": "file_3"}, Body: "I'm a simple response"},
			wantMatch: true,
		},
		{
			name:      "Should match GET request with header",
			input:     Request{Method: "GET", Path: "/bears/321", Headers: map[string]string{"authorization": "Bearer Bear 🐻"}},
			want:      MatchResult{StatusCode: 200, Matched: true, Headers: map[string]string{"content-type": "text/plain", "X-Mapping-File": "file_1"}, Body: "🐻"},
			wantMatch: true,
		},
		{
			name:      "Should match GET request and load body from file",
			input:     Request{Method: "GET", Path: "/match/me/123?file=true"},
			want:      MatchResult{StatusCode: 200, Matched: true, Headers: map[string]string{"content-type": "application/json", "X-Mapping-File": "file_2"}, Body: `{"message": "Hello from the body file"}`},
			wantMatch: true,
		},
		{
			name:      "Should match GET request if path contains request",
			input:     Request{Method: "GET", Path: "/thispath/contains/123"},
			want:      MatchResult{StatusCode: 200, Matched: true, Headers: map[string]string{"content-type": "text/plain", "X-Mapping-File": "file_4"}, Body: `Mapping contains path`},
			wantMatch: true,
		},
		{
			name:      "Should match GET request if path matches regex",
			input:     Request{Method: "GET", Path: "/regex/2"},
			want:      MatchResult{StatusCode: 200, Matched: true, Headers: map[string]string{"content-type": "text/plain", "X-Mapping-File": "file_5"}, Body: `Mapping with regex on path`},
			wantMatch: true,
		},
		{
			name:      "Should match GET request combining path regex and contains",
			input:     Request{Method: "GET", Path: "/combination/123"},
			want:      MatchResult{StatusCode: 200, Matched: true, Headers: map[string]string{"content-type": "text/plain", "X-Mapping-File": "file_6"}, Body: `Mapping combining path regex and contains`},
			wantMatch: true,
		},
		{
			name:      "Should match GET request combining path/headers regex and contains",
			input:     Request{Method: "GET", Path: "/combination/__1234?abc=s2", Headers: map[string]string{"accept": "application/json"}},
			want:      MatchResult{StatusCode: 200, Matched: true, Headers: map[string]string{"content-type": "application/json", "X-Mapping-File": "file_7"}, Body: `{"message": "Mapping combining path/headers regex and contains"}`},
			wantMatch: true,
		},
		{
			name:      "Should match POST request with body",
			input:     Request{Method: "POST", Path: "/order", Headers: map[string]string{"authorization": "Bearer ItsMe"}, Body: `{"cart": "555"}`},
			want:      MatchResult{StatusCode: 201, Matched: true, Headers: map[string]string{"location": "12345", "X-Mapping-File": "file_8"}},
			wantMatch: true,
		},
		{
			name:      "Should match POST request if body and header contain request",
			input:     Request{Method: "POST", Path: "/bears/contains", Headers: map[string]string{"content-type": "application/json"}, Body: `{"name": "Mr Bear", "honey": true}`},
			want:      MatchResult{StatusCode: 201, Matched: true, Headers: map[string]string{"location": "12345", "X-Mapping-File": "file_10"}},
			wantMatch: true,
		},
		{
			name:      "Should match PUT request if body matches single JSON path",
			input:     Request{Method: "PUT", Path: "/json/path", Body: `{"products": [{"id": "12345"}, {"id": "123452"}]}`},
			want:      MatchResult{StatusCode: 204, Matched: true, Headers: map[string]string{"multiple": "false", "X-Mapping-File": "file_12"}},
			wantMatch: true,
		},
		{
			name:      "Should match PUT request if body matches multiple JSON paths",
			input:     Request{Method: "PUT", Path: "/json/path", Body: `{"products": [{"id": "12346"}], "users": [{"name": "Bob"}]}`},
			want:      MatchResult{StatusCode: 204, Matched: true, Headers: map[string]string{"multiple": "true", "X-Mapping-File": "file_13"}},
			wantMatch: true,
		},
		{
			name:      "Should match POST request if body and header match regex",
			input:     Request{Method: "POST", Path: "/gopher/regex", Headers: map[string]string{"content-type": "application/json"}, Body: `{"name": "Mr Gopher", "honey": true}`},
			want:      MatchResult{StatusCode: 201, Matched: true, Headers: map[string]string{"location": "999", "X-Mapping-File": "file_11"}},
			wantMatch: true,
		},
		{
			name:      "Should not match GET request if path does not match regex",
			input:     Request{Method: "GET", Path: "/regex/abc"},
			wantMatch: false,
		},
		{
			name:      "Should not match POST request if header-exact does not match",
			input:     Request{Method: "POST", Path: "/order", Headers: map[string]string{"authorization": "Bearer NotMe"}},
			wantMatch: false,
		},
		{
			name:      "Should not match POST request if header-contains does not match",
			input:     Request{Method: "POST", Path: "/bears/contains", Headers: map[string]string{"content-type": "xml"}},
			wantMatch: false,
		},
		{
			name:      "Should not match POST request if body-contains does not match",
			input:     Request{Method: "POST", Path: "/bears/contains", Headers: map[string]string{"content-type": "json"}, Body: `no match`},
			wantMatch: false,
		},
		{
			name:      "Should not match a request when the method is not mapped",
			input:     Request{Method: "HEAD", Path: "/gopher/regex"},
			wantMatch: false,
		},
		{
			name:        "Should return 404 without closest match when no component matches",
			input:       Request{Method: "Post", Path: "/nomatchhere", Body: `{"message": "I have no matches :("}`},
			wantMatch:   false,
			wantPartial: false,
			want: MatchResult{
				Matched:    false,
				StatusCode: 404,
				Headers:    map[string]string{"Content-type": "application/json"},
				Body: NotFoundResponse{
					Message:        NoMappingFoundMessage,
					Request:        Request{Method: "GET", Path: "/nomatchhere"},
					ClosestMapping: nil,
				},
			},
		},
		{
			name:        "Should return 404 with the closest mapping when no match is found",
			input:       Request{Method: "GET", Path: "/bears/321"},
			wantMatch:   false,
			wantPartial: true,
			want: MatchResult{
				Matched:    false,
				StatusCode: 404,
				Headers:    map[string]string{"Content-type": "application/json", "X-Mapping-File": "file_1"},
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

	matcher := NewMatcher(regexCache, jsonPathCache)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping, matched, partial := matcher.Match(context.Background(), tt.input, mappings, nil)

			result := NewMatchResult(&mapping, tt.input, matched, partial)

			if tt.wantMatch {
				require.Equal(t, tt.wantMatch, matched)
			}
			if tt.wantPartial {
				require.Equal(t, tt.wantPartial, partial)
			}

			if tt.wantMatch || tt.wantPartial {
				require.Equal(t, tt.want, result)
			}
		})
	}
}

func getMappings() Mappings {
	mappings := []Mapping{
		{
			Request: RequestMapping{
				Method:  "GET",
				Path:    CommonMatch{Exact: "/bears/321"},
				Headers: map[string]CommonMatch{"authorization": {Exact: "Bearer Bear 🐻"}},
			},
			Response: ResponseMapping{StatusCode: 200, Headers: map[string]string{"content-type": "text/plain"}, Body: "🐻"},
			MaxScore: 2,
			Cost:     0,
			FilePath: "file_1",
		},
		{
			Request:  RequestMapping{Method: "GET", Path: CommonMatch{Exact: "/match/me/123?file=true"}},
			Response: ResponseMapping{StatusCode: 200, Headers: map[string]string{"content-type": "application/json"}, Body: `{"message": "Hello from the body file"}`},
			MaxScore: 1,
			Cost:     0,
			FilePath: "file_2",
		},
		{
			Request:  RequestMapping{Method: "GET", Path: CommonMatch{Exact: "/simple"}},
			Response: ResponseMapping{StatusCode: 200, Headers: map[string]string{"content-type": "text/plain"}, Body: "I'm a simple response"},
			MaxScore: 1,
			Cost:     0,
			FilePath: "file_3",
		},
		{
			Request:  RequestMapping{Method: "GET", Path: CommonMatch{Contains: []string{"contains/123"}}},
			Response: ResponseMapping{StatusCode: 200, Headers: map[string]string{"content-type": "text/plain"}, Body: "Mapping contains path"},
			MaxScore: 1,
			Cost:     2,
			FilePath: "file_4",
		},
		{
			Request:  RequestMapping{Method: "GET", Path: CommonMatch{Patterns: []string{"regex/[0-9]+$", `regex/\d{1}$`}}},
			Response: ResponseMapping{StatusCode: 200, Headers: map[string]string{"content-type": "text/plain"}, Body: "Mapping with regex on path"},
			MaxScore: 2,
			Cost:     10,
			FilePath: "file_5",
		},
		{
			Request:  RequestMapping{Method: "GET", Path: CommonMatch{Patterns: []string{"combination/[0-9]+$"}, Contains: []string{"123"}}},
			Response: ResponseMapping{StatusCode: 200, Headers: map[string]string{"content-type": "text/plain"}, Body: "Mapping combining path regex and contains"},
			MaxScore: 2,
			Cost:     7,
			FilePath: "file_6",
		},
		{
			Request: RequestMapping{
				Method:  "GET",
				Path:    CommonMatch{Contains: []string{"1234", "abc"}, Patterns: []string{`[_]{2}`}},
				Headers: map[string]CommonMatch{"accept": {Contains: []string{"application"}, Patterns: []string{"application/(json|xml){1}$"}}},
			},
			Response: ResponseMapping{StatusCode: 200, Headers: map[string]string{"content-type": "application/json"}, Body: `{"message": "Mapping combining path/headers regex and contains"}`},
			MaxScore: 5,
			Cost:     16,
			FilePath: "file_7",
		},
		{
			Request: RequestMapping{
				Method:  "POST",
				Path:    CommonMatch{Exact: "/order"},
				Headers: map[string]CommonMatch{"Authorization": {Exact: "Bearer ItsMe"}},
				Body:    BodyMatch{CommonMatch: CommonMatch{Exact: `{"cart": "555"}`}},
			},
			Response: ResponseMapping{StatusCode: 201, Headers: map[string]string{"location": "12345"}},
			MaxScore: 3,
			Cost:     0,
			FilePath: "file_8",
		},
		{
			Request:  RequestMapping{Method: "POST", Path: CommonMatch{Exact: "/order"}, Body: BodyMatch{CommonMatch: CommonMatch{Exact: `{"cart": "555"}`}}},
			Response: ResponseMapping{StatusCode: 401},
			MaxScore: 2,
			Cost:     0,
			FilePath: "file_9",
		},
		{
			Request: RequestMapping{
				Method:  "POST",
				Headers: map[string]CommonMatch{"content-type": {Contains: []string{"json"}}},
				Path:    CommonMatch{Contains: []string{"bears", "contains"}},
				Body:    BodyMatch{CommonMatch: CommonMatch{Contains: []string{"name", `"honey": true`}}},
			},
			Response: ResponseMapping{StatusCode: 201, Headers: map[string]string{"location": "12345"}},
			MaxScore: 5,
			Cost:     10,
			FilePath: "file_10",
		},
		{
			Request: RequestMapping{
				Method:  "POST",
				Headers: map[string]CommonMatch{"content-type": {Patterns: []string{"^application/(json|xml){1}$"}}},
				Path:    CommonMatch{Exact: "/gopher/regex"},
				Body:    BodyMatch{CommonMatch: CommonMatch{Patterns: []string{`"name":\s*"[A-z\s]+"`}}},
			},
			Response: ResponseMapping{StatusCode: 201, Headers: map[string]string{"location": "999"}},
			MaxScore: 3,
			Cost:     10,
			FilePath: "file_11",
		},
		{
			Request:  RequestMapping{Method: "PUT", Path: CommonMatch{Exact: "/json/path"}, Body: BodyMatch{JsonPath: []string{"$.products[?(@.id == '12345')]"}}},
			Response: ResponseMapping{StatusCode: 204, Headers: map[string]string{"multiple": "false"}},
			MaxScore: 2,
			Cost:     4,
			FilePath: "file_12",
		},
		{
			Request: RequestMapping{
				Method: "PUT",
				Path:   CommonMatch{Exact: "/json/path"},
				Body:   BodyMatch{JsonPath: []string{"$.products[?(@.id == '12346')]", "$.users[?(@.name == 'Bob')]"}},
			},
			Response: ResponseMapping{StatusCode: 204, Headers: map[string]string{"multiple": "true"}},
			MaxScore: 3,
			Cost:     8,
			FilePath: "file_13",
		},
		{
			Request:  RequestMapping{Method: "DELETE", Path: CommonMatch{Exact: "/cart/123"}},
			Response: ResponseMapping{StatusCode: 204},
			MaxScore: 1,
			Cost:     0,
			FilePath: "file_14",
		},
	}

	ms := make(Mappings)
	_ = ms.PutAll(mappings)
	return ms
}
