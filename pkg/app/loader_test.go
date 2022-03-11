package app

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestDecodeFile(t *testing.T) {

	tests := []struct {
		name        string
		path        string
		wantErr     error
		anyErr      bool
		wantMapping Mapping
	}{
		{
			name: "Should decode file successfuly",
			path: "testdata/decode/get_product_12345.json",
			wantMapping: Mapping{
				Request:  MappingRequest{Method: "GET", URL: "/product/12345", Headers: map[string]string{"accept": "application/json"}},
				Response: MappingResponse{StatusCode: 200, Headers: map[string]string{"content-type": "application/json"}, BodyFile: "get_product_12345_response.json"},
			},
		},
		{
			name:    "Should return an error if file doesn't exist",
			path:    "testdata/decode/you_shall_pass.json",
			wantErr: FileNotFound("testdata/decode/you_shall_pass.json"),
		},
		// TODO: Test to check on the other error path
	}

	loader := NewFileLoader()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping, err := loader.decodeFile(tt.path)

			if err != nil {
				if !tt.anyErr {
					if tt.wantErr == nil {
						t.FailNow()
					}
					assert.Equal(t, err.Error(), tt.wantErr.Error())
				}
				return
			}

			assert.Equal(t, mapping, tt.wantMapping)
		})
	}
}

func TestLoadMappings(t *testing.T) {
	tests := []struct {
		name         string
		rootPath     string
		wantErr      error
		anyErr       bool
		wantMappings Mappings
	}{
		{
			name:     "Should load mappings for each request method",
			rootPath: "testdata/load/valid",
			wantMappings: Mappings{
				"GET": []Mapping{{
					Request:  MappingRequest{Method: "GET", URL: "/product/12345", Headers: map[string]string{"accept": "application/json"}},
					Response: MappingResponse{StatusCode: 200, Headers: map[string]string{"content-type": "application/json"}, BodyFile: "get_product_12345_response.json"},
				}},
				"POST": []Mapping{{
					Request:  MappingRequest{Method: "POST", URL: "/order", Headers: map[string]string{"content-type": "application/json"}},
					Response: MappingResponse{StatusCode: 200},
				}},
			},
		},
		{
			name:     "Should throw error if mapping is invalid",
			rootPath: "testdata/load/invalid",
			wantErr:  ValidationErrors{ValidationError{"Request.Method", "Method is required"}, ValidationError{"Request.URL", "URL is required"}},
		},
	}

	loader := NewFileLoader()

	mappings := make(Mappings)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := loader.loadMappings(tt.rootPath, mappings)
			if err != nil {
				if !tt.anyErr {
					if tt.wantErr == nil {
						t.FailNow()
					}
					assert.Equal(t, err.Error(), tt.wantErr.Error())
				}
				return
			}

			assert.Equal(t, mappings, tt.wantMappings)
		})
	}

}
