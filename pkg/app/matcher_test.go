package app

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestMatcher(t *testing.T) {
	tests := []struct {
		name  string
		input Request
		want  MappingResponse
	}{
		{
			name:  "Should match GET request",
			input: Request{Method: "GET", Path: "/bears/321", Headers: map[string]string{"Authorization": "Bearer Bear 🐻"}},
			want:  MappingResponse{StatusCode: 200, Headers: map[string]string{"content-type": "text/plain"}, Body: "🐻"},
		},
		{
			name:  "Should match GET request and load body from file",
			input: Request{Method: "GET", Path: "/match/me/123?file=true"},
			want:  MappingResponse{StatusCode: 200, Headers: map[string]string{"content-type": "application/json"}, Body: `{"message": "Hello from the body file"}`},
		},
		{
			name:  "Should return 404 with the closest mapping when no match is found",
			input: Request{Method: "GET", Path: "/bears/321"},
			want: MappingResponse{
				StatusCode: 404,
				Headers:    map[string]string{"content-type": "application/json"},
				Body:       `{"message": "No mapping found for the request","request": {"method": "GET","path": "/bears/123"},"closestMapping": {"method": "GET","path": {"exact": "/bears/123"},"headers": {"Authorization": "Bearer Bear 🐻"}}}`,
			},
		},
	}

	matcher := NewMatcher(getMappings())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := matcher.Match(tt.input)
			assert.Equal(t, response, tt.want)
		})
	}
}

func getMappings() Mappings {
	return Mappings{
		"GET": []Mapping{
			{
				Request: MappingRequest{
					Method:  "GET",
					Path:    PathMapping{Exact: "/bears/321"},
					Headers: map[string]HeaderMapping{"Authorization": {Exact: "Bearer Bear 🐻"}},
				},
				Response: MappingResponse{StatusCode: 200, Headers: map[string]string{"Content-type": "text/plain"}, Body: "🐻"},
			},
			{
				Request:  MappingRequest{Method: "GET", Path: PathMapping{Exact: "/match/me/123?file=true"}},
				Response: MappingResponse{StatusCode: 200, Headers: map[string]string{"Content-type": "application/json"}, BodyFile: "body_from_file.json"},
			},
		},
		"POST": []Mapping{
			{
				Request: MappingRequest{
					Method:  "POST",
					Path:    PathMapping{Exact: "/order"},
					Headers: map[string]HeaderMapping{"Authorization": {Exact: "Bearer ItsMe"}},
					Body:    BodyMapping{Exact: `{"cart": "555"}`},
				},
				Response: MappingResponse{StatusCode: 201, Headers: map[string]string{"location": "12345"}},
			},
			{
				Request:  MappingRequest{Method: "POST", Path: PathMapping{Exact: "/order"}, Body: BodyMapping{Exact: `{"cart": "555"}`}},
				Response: MappingResponse{StatusCode: 401},
			},
		},
		"DELETE": []Mapping{
			{
				Request:  MappingRequest{Method: "DELETE", Path: PathMapping{Exact: "/cart/123"}},
				Response: MappingResponse{StatusCode: 204},
			},
		},
	}
}
