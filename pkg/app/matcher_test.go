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
			want:  MappingResponse{StatusCode: 200, Headers: map[string]string{"content-type": "text/plain"}},
		},
	}

	matcher := NewMatcher(getMappings())

	for _, tt := range tests {
		response := matcher.Match(tt.input)
		assert.Equal(t, response, tt.want)
	}
}

func getMappings() Mappings {
	return Mappings{
		"GET": []Mapping{
			{
				Request:  MappingRequest{Method: "GET", Path: "/bears/321", Headers: map[string]string{"Authorization": "Bearer Bear 🐻"}},
				Response: MappingResponse{StatusCode: 200, Headers: map[string]string{"Content-type": "text/plain"}, Body: "🐻"},
			},
			{
				Request:  MappingRequest{Method: "GET", Path: "/match/me/123?file=true"},
				Response: MappingResponse{StatusCode: 200, Headers: map[string]string{"Content-type": "application/json"}, BodyFile: "body_from_file.json"},
			},
		},
		"POST": []Mapping{
			{
				Request:  MappingRequest{Method: "POST", Path: "/order", Headers: map[string]string{"Authorization": "Bearer ItsMe"}, Body: "{\"cart\": \"555\"}"},
				Response: MappingResponse{StatusCode: 201, Headers: map[string]string{"location": "12345"}},
			},
			{
				Request:  MappingRequest{Method: "POST", Path: "/order", Body: "{\"cart\": \"555\"}"},
				Response: MappingResponse{StatusCode: 401},
			},
		},
		"DELETE": []Mapping{
			{
				Request:  MappingRequest{Method: "DELETE", Path: "/cart/123"},
				Response: MappingResponse{StatusCode: 204},
			},
		},
	}
}
