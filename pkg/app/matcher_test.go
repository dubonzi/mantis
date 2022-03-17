package app

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestMatcher(t *testing.T) {
	tests := []struct {
		name      string
		input     Request
		want      MatcherResult
		wantMatch bool
	}{
		{
			name:      "Should match simple request",
			input:     Request{Method: "GET", Path: "/simple"},
			want:      MatcherResult{StatusCode: 200, Matched: true, Headers: map[string]string{"content-type": "text/plain"}, Body: "I'm a simple response"},
			wantMatch: true,
		},
		{
			name:      "Should match GET request with header",
			input:     Request{Method: "GET", Path: "/bears/321", Headers: map[string]string{"authorization": "Bearer Bear 🐻"}},
			want:      MatcherResult{StatusCode: 200, Matched: true, Headers: map[string]string{"content-type": "text/plain"}, Body: "🐻"},
			wantMatch: true,
		},
		{
			name:      "Should match GET request and load body from file",
			input:     Request{Method: "GET", Path: "/match/me/123?file=true"},
			want:      MatcherResult{StatusCode: 200, Matched: true, Headers: map[string]string{"content-type": "application/json"}, Body: `{"message": "Hello from the body file"}`},
			wantMatch: true,
		},
		{
			name:      "Should match POST request with body",
			input:     Request{Method: "POST", Path: "/order", Headers: map[string]string{"authorization": "Bearer ItsMe"}, Body: `{"cart": "555"}`},
			want:      MatcherResult{StatusCode: 201, Matched: true, Headers: map[string]string{"location": "12345"}},
			wantMatch: true,
		},
		{
			name:  "Should return 404 with the closest mapping when no match is found",
			input: Request{Method: "GET", Path: "/bears/321"},
			want: MatcherResult{
				Matched:    false,
				StatusCode: 404,
				Body: NotFoundResponse{
					Message: NoMappingFoundMessage,
					Request: Request{Method: "GET", Path: "/bears/321"},
					ClosestMapping: &RequestMapping{
						Method:  "GET",
						Path:    PathMapping{Exact: "/bears/321"},
						Headers: map[string]HeaderMapping{"authorization": {Exact: "Bearer Bear 🐻"}},
					},
				},
			},
			wantMatch: false,
		},
	}

	matcher := NewMatcher(getMappings())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := matcher.Match(tt.input)
			if !assert.IsEqual(response, tt.want) {
				t.Logf("%s doest not equal %s", response, tt.want)
				t.FailNow()
			}
			assert.Equal(t, response, tt.want)
		})
	}
}

func getMappings() Mappings {
	return Mappings{
		"GET": []Mapping{
			{
				Request: RequestMapping{
					Method:  "GET",
					Path:    PathMapping{Exact: "/bears/321"},
					Headers: map[string]HeaderMapping{"authorization": {Exact: "Bearer Bear 🐻"}},
				},
				Response: ResponseMapping{StatusCode: 200, Headers: map[string]string{"content-type": "text/plain"}, Body: "🐻"},
			},
			{
				Request:  RequestMapping{Method: "GET", Path: PathMapping{Exact: "/match/me/123?file=true"}},
				Response: ResponseMapping{StatusCode: 200, Headers: map[string]string{"content-type": "application/json"}, Body: `{"message": "Hello from the body file"}`},
			},
			{
				Request:  RequestMapping{Method: "GET", Path: PathMapping{Exact: "/simple"}},
				Response: ResponseMapping{StatusCode: 200, Headers: map[string]string{"content-type": "text/plain"}, Body: "I'm a simple response"},
			},
		},
		"POST": []Mapping{
			{
				Request: RequestMapping{
					Method:  "POST",
					Path:    PathMapping{Exact: "/order"},
					Headers: map[string]HeaderMapping{"Authorization": {Exact: "Bearer ItsMe"}},
					Body:    BodyMapping{Exact: `{"cart": "555"}`},
				},
				Response: ResponseMapping{StatusCode: 201, Headers: map[string]string{"location": "12345"}},
			},
			{
				Request:  RequestMapping{Method: "POST", Path: PathMapping{Exact: "/order"}, Body: BodyMapping{Exact: `{"cart": "555"}`}},
				Response: ResponseMapping{StatusCode: 401},
			},
		},
		"DELETE": []Mapping{
			{
				Request:  RequestMapping{Method: "DELETE", Path: PathMapping{Exact: "/cart/123"}},
				Response: ResponseMapping{StatusCode: 204},
			},
		},
	}
}
