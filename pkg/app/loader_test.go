package app

import (
	"errors"
	"testing"
	"time"

	"github.com/americanas-go/config"
	"github.com/stretchr/testify/require"
)

var (
	validLoaderMappings = []*Mapping{
		{
			Request: RequestMapping{
				Method: "GET",
				Path:   CommonMatch{Exact: "/delay/fixed"},
			},
			Response: ResponseMapping{StatusCode: 204, ResponseDelay: Delay{Fixed: FixedDelay{Duration: Duration(time.Millisecond * 250)}}},
		},
		{
			Request: RequestMapping{
				Method:  "GET",
				Path:    CommonMatch{Exact: "/product/12345"},
				Headers: map[string]CommonMatch{"accept": {Exact: "application/json"}},
			},
			Response: ResponseMapping{StatusCode: 200, Headers: map[string]string{"content-type": "application/json"}, Body: `{"id": "12345","name": "My Product","description": "This is it"}`, BodyFile: "get_product_12345_response.json"},
		},
		{
			Request: RequestMapping{
				Method:  "GET",
				Path:    CommonMatch{Patterns: []string{"/regex/[A-z0-9]+", "/regex/.{1}"}},
				Headers: map[string]CommonMatch{"accept": {Patterns: []string{"application/(json|xml){1}", ".*json.*"}}},
			},
			Response: ResponseMapping{StatusCode: 200, Headers: map[string]string{"content-type": "application/json"}, Body: `{"id": "regex","name": "Regex response"}`},
		},
		{
			Request: RequestMapping{
				Method: "PUT",
				Path:   CommonMatch{Exact: "/json/path"},
				Body:   BodyMatch{JsonPath: []string{"$[?(@.product.id == '12345')]", "$.person[?(@.age > 21 || @.name == 'John')]"}},
			},
			Response: ResponseMapping{StatusCode: 204},
		},
		{
			Request: RequestMapping{
				Method:  "POST",
				Path:    CommonMatch{Exact: "/order"},
				Headers: map[string]CommonMatch{"content-type": {Exact: "application/json"}},
				Body:    BodyMatch{CommonMatch: CommonMatch{Contains: []string{"orderId", "999"}}},
			},
			Response: ResponseMapping{StatusCode: 200},
		},
		{
			Scenario: ScenarioMapping{
				Name:          "My Scenario",
				StartingState: true,
				State:         "First state",
				NewState:      "Second state",
			},
			Request: RequestMapping{
				Method:  "POST",
				Path:    CommonMatch{Exact: "/scenario"},
				Headers: map[string]CommonMatch{"content-type": {Exact: "application/json"}},
				Body:    BodyMatch{CommonMatch: CommonMatch{Contains: []string{"scenario", "test"}}},
			},
			Response: ResponseMapping{StatusCode: 200},
		},
		{
			Scenario: ScenarioMapping{
				Name:  "My Scenario",
				State: "Second state",
			},
			Request: RequestMapping{
				Method:  "POST",
				Path:    CommonMatch{Exact: "/scenario"},
				Headers: map[string]CommonMatch{"content-type": {Exact: "application/json"}},
				Body:    BodyMatch{CommonMatch: CommonMatch{Contains: []string{"scenario", "test"}}},
			},
			Response: ResponseMapping{StatusCode: 200},
		},
	}
)

func TestGetMappings(t *testing.T) {
	wantMappings := make(Mappings)
	_ = wantMappings.PutAll(validLoaderMappings)

	tests := []struct {
		name         string
		before       func(t *testing.T)
		wantMappings Mappings
	}{
		{
			name: "Should return mappings",
			before: func(t *testing.T) {
				t.Setenv("LOADER_PATH_MAPPING", "testdata/load/valid/mapping")
				t.Setenv("LOADER_PATH_RESPONSE", "testdata/load/valid/response")
				config.Load()
			},
			wantMappings: wantMappings,
		},
		{
			name: "Should return empty mappings if path is not found",
			before: func(t *testing.T) {
				t.Setenv("LOADER_PATH_MAPPING", "")
				t.Setenv("LOADER_PATH_RESPONSE", "")
				config.Load()
			},
			wantMappings: make(Mappings),
		},
	}

	loader := NewLoader(NewRegexCache(), NewJSONPathCache())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.before(t)

			mappings, err := loader.GetMappings()
			if err != nil {
				t.Log("did not expect an error, but got: ", err)
				t.FailNow()
			}

			require.Equal(t, mappings, tt.wantMappings)
		})
	}
}

func TestDecodeFile(t *testing.T) {

	tests := []struct {
		name        string
		path        string
		wantErr     error
		anyErr      bool
		wantMapping Mapping
	}{
		{
			name: "Should decode file successfully",
			path: "testdata/decode/get_product_12345.json",
			wantMapping: Mapping{
				Request: RequestMapping{
					Method:  "GET",
					Path:    CommonMatch{Exact: "/product/12345"},
					Headers: map[string]CommonMatch{"accept": {Exact: "application/json"}},
				},
				Response: ResponseMapping{StatusCode: 200, Headers: map[string]string{"content-type": "application/json"}, BodyFile: "get_product_12345_response.json"},
			},
		},
		{
			name:    "Should return an error if file doesn't exist",
			path:    "testdata/decode/you_shall_pass.json",
			wantErr: FileNotFound("testdata/decode/you_shall_pass.json"),
		},
		{
			name:    "Should return an error when unmarshaling an invalid mapping",
			path:    "testdata/decode/invalid_mapping.json",
			wantErr: errors.New("json: cannot unmarshal bool into Go struct field RequestMapping.request.method of type string"),
		},
		// TODO: Test to check on the other error path
	}

	loader := NewLoader(NewRegexCache(), NewJSONPathCache())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping, err := loader.decodeFile(tt.path)

			if err != nil {
				if !tt.anyErr {
					if tt.wantErr == nil {
						t.Log("did not expect an error, but got: ", err)
						t.FailNow()
					}
					require.Equal(t, err.Error(), tt.wantErr.Error())
				}
				return
			}

			require.Equal(t, mapping, tt.wantMapping)
		})
	}
}

func TestLoadMappings(t *testing.T) {
	wantMappings := make(Mappings)
	_ = wantMappings.PutAll(validLoaderMappings)

	tests := []struct {
		name          string
		mappingsPath  string
		responsesPath string
		wantErr       string
		anyErr        bool
		wantMappings  Mappings
	}{
		{
			name:          "Should load mappings for each request method",
			mappingsPath:  "testdata/load/valid/mapping",
			responsesPath: "testdata/load/valid/response",
			wantMappings:  wantMappings,
		},
		{
			name:         "Should throw error if mapping is invalid",
			mappingsPath: "testdata/load/invalid",
			wantErr:      `error adding mapping from file [ testdata/load/invalid/invalid_mapping.json ]: mapping definition is invalid: [{"field":"Request.Method","message":"Method is required"},{"field":"Request.Path","message":"Path mapping is required"}]`,
		},
	}

	loader := NewLoader(NewRegexCache(), NewJSONPathCache())

	mappings := make(Mappings)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := loader.loadMappings(tt.mappingsPath, tt.responsesPath, mappings)
			if err != nil {
				if !tt.anyErr {
					if tt.wantErr == "" {
						t.Log("did not expect an error, but got: ", err)
						t.FailNow()
					}
					require.Equal(t, err.Error(), tt.wantErr)
				}
				return
			}

			require.Equal(t, mappings, tt.wantMappings)
		})
	}

}
