package app

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestMatcher(t *testing.T) {
	tests := []struct {
		name      string
		input     Request
		want      *MappingResponse
		wantMatch bool
	}{
		{
			name:      "Should match simple request",
			input:     Request{Method: "GET", Path: "/simple"},
			want:      &MappingResponse{StatusCode: 200, Headers: map[string]string{"content-type": "text/plain"}, Body: "I'm a simple response"},
			wantMatch: true,
		},
		{
			name:      "Should match GET request with header",
			input:     Request{Method: "GET", Path: "/bears/321", Headers: map[string]string{"Authorization": "Bearer Bear 🐻"}},
			want:      &MappingResponse{StatusCode: 200, Headers: map[string]string{"Content-type": "text/plain"}, Body: "🐻"},
			wantMatch: true,
		},
		{
			name:      "Should match GET request and load body from file",
			input:     Request{Method: "GET", Path: "/match/me/123?file=true"},
			want:      &MappingResponse{StatusCode: 200, Headers: map[string]string{"Content-type": "application/json"}, Body: `{"message": "Hello from the body file"}`},
			wantMatch: true,
		},
		{
			name:      "Should match POST request with body",
			input:     Request{Method: "POST", Path: "/order", Headers: map[string]string{"Authorization": "Bearer ItsMe"}, Body: `{"cart": "555"}`},
			want:      &MappingResponse{StatusCode: 201, Headers: map[string]string{"location": "12345"}},
			wantMatch: true,
		},
		{
			name:  "Should return 404 with the closest mapping when no match is found",
			input: Request{Method: "GET", Path: "/bears/321"},
			want: &MappingResponse{
				StatusCode: 404,
				Headers:    map[string]string{"Content-type": "application/json"},
				Body:       `{"message": "No mapping found for the request","request": {"method": "GET","path": "/bears/123"},"closestMapping": {"method": "GET","path": {"exact": "/bears/123"},"headers": {"Authorization": "Bearer Bear 🐻"}}}`,
			},
			wantMatch: false,
		},
	}

	matcher := NewMatcher(getMappings())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, matched := matcher.Match(tt.input)
			assert.Equal(t, matched, tt.wantMatch)
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
			{
				Request:  MappingRequest{Method: "GET", Path: PathMapping{Exact: "/simple"}},
				Response: MappingResponse{StatusCode: 200, Headers: map[string]string{"content-type": "text/plain"}, Body: "I'm a simple response"},
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
