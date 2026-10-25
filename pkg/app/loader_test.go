package app

import (
	"io/fs"
	"path/filepath"
	"regexp/syntax"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadMappingsMissingDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing")
	loader := NewLoader(LoaderConfig{MappingPath: path}, NewRegexCache(), NewJSONPathCache(), NewScenarioHandler(nil))

	_, err := loader.GetMappings()

	require.ErrorContains(t, err, "error walking mapping path")
	assert.ErrorContains(t, err, path)
	assert.ErrorIs(t, err, fs.ErrNotExist)
}

func TestProcessMappingBodyFile(t *testing.T) {
	tests := []struct {
		name     string
		bodyFile string
		wantBody string
	}{
		{
			name:     "Should compact JSON",
			bodyFile: "get_product_12345_response.json",
			wantBody: `{"id":"12345","name":"My Product","description":"This is it"}`,
		},
		{
			name:     "Should preserve CSV whitespace",
			bodyFile: "products.csv",
			wantBody: "  id,name\n    1,Frodo\n\t2,Sam\n  3,Pippin",
		},
		{
			name:     "Should preserve invalid JSON unchanged",
			bodyFile: "invalid.json",
			wantBody: "  {\n    \"id\": \"12345\",\n  }\n",
		},
		{
			name:     "Should preserve JSON values while compacting",
			bodyFile: "values.json",
			wantBody: `{"text":"  first\n  second  ","number":1.20e3,"key":1,"key":2}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewLoader(LoaderConfig{}, NewRegexCache(), NewJSONPathCache(), NewScenarioHandler(nil))
			mapping := Mapping{Response: ResponseMapping{BodyFile: tt.bodyFile}}

			err := loader.processMapping(&mapping, "mapping.json", "testdata/load/valid/response")

			require.NoError(t, err)
			assert.Equal(t, tt.wantBody, mapping.Response.Body)
		})
	}
}

func TestProcessMappingErrors(t *testing.T) {
	tests := []struct {
		name    string
		mapping Mapping
		wantErr string
	}{
		{
			name:    "Should describe regex compilation errors",
			mapping: Mapping{Request: RequestMapping{Path: CommonMatch{Patterns: []string{"["}}}},
			wantErr: "error compiling regex patterns",
		},
		{
			name:    "Should describe jsonpath parsing errors",
			mapping: Mapping{Request: RequestMapping{Body: BodyMatch{JsonPath: []string{"$["}}}},
			wantErr: "error parsing jsonpath expressions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewLoader(LoaderConfig{}, NewRegexCache(), NewJSONPathCache(), NewScenarioHandler(nil))

			err := loader.processMapping(&tt.mapping, "mapping.json", "")

			require.ErrorContains(t, err, tt.wantErr)
			if tt.name == "Should describe regex compilation errors" {
				var syntaxErr *syntax.Error
				assert.ErrorAs(t, err, &syntaxErr)
			}
		})
	}
}

var (
	validLoaderMappings = []Mapping{
		{
			Request: RequestMapping{
				Method: "GET",
				Path:   CommonMatch{Exact: "/delay/fixed"},
			},
			Response: ResponseMapping{StatusCode: 204, ResponseDelay: Delay{Fixed: FixedDelay{Duration: Duration(time.Millisecond * 250)}}},
			MaxScore: 1,
			FilePath: "testdata/load/valid/mapping/get_fixed_delay.json",
		},
		{
			Request: RequestMapping{
				Method:  "GET",
				Path:    CommonMatch{Exact: "/product/12345"},
				Headers: map[string]CommonMatch{"accept": {Exact: "application/json"}},
			},
			Response: ResponseMapping{StatusCode: 200, Headers: map[string]string{"content-type": "application/json"}, Body: `{"id":"12345","name":"My Product","description":"This is it"}`, BodyFile: "get_product_12345_response.json"},
			MaxScore: 2,
			FilePath: "testdata/load/valid/mapping/get_product_12345.json",
		},
		{
			Request: RequestMapping{
				Method:  "GET",
				Path:    CommonMatch{Patterns: []string{"/regex/[A-z0-9]+", "/regex/.{1}"}},
				Headers: map[string]CommonMatch{"accept": {Patterns: []string{"application/(json|xml){1}", ".*json.*"}}},
			},
			Response: ResponseMapping{StatusCode: 200, Headers: map[string]string{"content-type": "application/json"}, Body: `{"id": "regex","name": "Regex response"}`},
			MaxScore: 4,
			Cost:     20,
			FilePath: "testdata/load/valid/mapping/get_regex.json",
		},
		{
			Request: RequestMapping{
				Method:  "GET",
				Path:    CommonMatch{Patterns: []string{"/alpha/[A-Za-z]+"}},
				Headers: map[string]CommonMatch{"accept": {Patterns: []string{"application/(json|xml){1}", ".*application.*"}}},
			},
			Response: ResponseMapping{StatusCode: 200, Headers: map[string]string{"content-type": "application/json"}, Body: `{"id": "regex","name": "Regex response"}`},
			MaxScore: 3,
			Cost:     15,
			FilePath: "testdata/load/valid/mapping/multiple.json",
		},
		{
			Request: RequestMapping{
				Method:  "GET",
				Path:    CommonMatch{Patterns: []string{"/search/query/[a-zA-Z0-9]+"}},
				Headers: map[string]CommonMatch{"accept": {Patterns: []string{"video/(mp4|avi)"}}},
			},
			Response: ResponseMapping{StatusCode: 200, Headers: map[string]string{"content-type": "application/json"}, Body: `{"id": "regex","name": "Regex response"}`},
			MaxScore: 2,
			Cost:     10,
			FilePath: "testdata/load/valid/mapping/multiple.json",
		},
		{
			Request: RequestMapping{
				Method: "PUT",
				Path:   CommonMatch{Exact: "/json/path"},
				Body:   BodyMatch{JsonPath: []string{"$[?(@.product.id == '12345')]", "$.person[?(@.age > 21 || @.name == 'John')]"}},
			},
			Response: ResponseMapping{StatusCode: 204},
			MaxScore: 3,
			Cost:     8,
			FilePath: "testdata/load/valid/mapping/put_json_path.json",
		},
		{
			Request: RequestMapping{
				Method:  "POST",
				Path:    CommonMatch{Exact: "/order"},
				Headers: map[string]CommonMatch{"content-type": {Exact: "application/json"}},
				Body:    BodyMatch{CommonMatch: CommonMatch{Contains: []string{"orderId", "999"}}},
			},
			Response: ResponseMapping{StatusCode: 200},
			MaxScore: 4,
			Cost:     4,
			FilePath: "testdata/load/valid/mapping/post_order.json",
		},
	}
	validScenarioMappings = []Mapping{
		{
			Scenario: &ScenarioMapping{
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
			MaxScore: 4,
			Cost:     4,
			FilePath: "testdata/load/valid/mapping/post_scenario_start.json",
		},
		{
			Scenario: &ScenarioMapping{
				Name:  "My Scenario",
				State: "Second state",
			},
			Request: RequestMapping{
				Method:  "POST",
				Path:    CommonMatch{Exact: "/scenario"},
				Headers: map[string]CommonMatch{"content-type": {Exact: "application/json"}},
				Body:    BodyMatch{CommonMatch: CommonMatch{Contains: []string{"scenario", "test"}}},
			},
			Response: ResponseMapping{StatusCode: 400},
			MaxScore: 4,
			Cost:     4,
			FilePath: "testdata/load/valid/mapping/post_scenario_state.json",
		},
	}
)

func TestGetMappings(t *testing.T) {
	wantMappings := make(Mappings)
	for _, mapping := range validLoaderMappings {
		require.NoError(t, wantMappings.Put(mapping))
	}
	wantScenarioMappings := make(Mappings)
	for _, mapping := range validScenarioMappings {
		require.NoError(t, wantScenarioMappings.Put(mapping))
	}

	tests := []struct {
		name                 string
		config               LoaderConfig
		wantMappings         Mappings
		wantScenarioMappings Mappings
		wantErr              string
	}{
		{
			name: "Should return mappings",
			config: LoaderConfig{
				MappingPath:  "testdata/load/valid/mapping",
				ResponsePath: "testdata/load/valid/response",
			},
			wantMappings:         wantMappings,
			wantScenarioMappings: wantScenarioMappings,
		},
		{
			name:                 "Should return empty mappings for an empty directory",
			config:               LoaderConfig{MappingPath: t.TempDir()},
			wantMappings:         make(Mappings),
			wantScenarioMappings: make(Mappings),
		},
		{
			name:    "Should return an error for an empty path",
			wantErr: "error walking mapping path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scHandler := NewScenarioHandler(nil)
			loader := NewLoader(tt.config, NewRegexCache(), NewJSONPathCache(), scHandler)

			gotMappings, err := loader.GetMappings()
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, tt.wantMappings, gotMappings)
			assert.Equal(t, tt.wantScenarioMappings, scHandler.scenarioMappings)
		})
	}
}

func TestDecodeFile(t *testing.T) {

	tests := []struct {
		name        string
		path        string
		wantErr     error
		anyErr      bool
		wantMapping []Mapping
	}{
		{
			name: "Should decode file successfully",
			path: "testdata/decode/get_product_12345.json",
			wantMapping: []Mapping{{
				Request: RequestMapping{
					Method:  "GET",
					Path:    CommonMatch{Exact: "/product/12345"},
					Headers: map[string]CommonMatch{"accept": {Exact: "application/json"}},
				},
				Response: ResponseMapping{StatusCode: 200, Headers: map[string]string{"content-type": "application/json"}, BodyFile: "get_product_12345_response.json"},
			}},
		},
		{
			name:    "Should return an error if file doesn't exist",
			path:    "testdata/decode/you_shall_pass.json",
			wantErr: FileNotFound("testdata/decode/you_shall_pass.json"),
		},
		{
			name:   "Should return an error when unmarshaling an invalid mapping",
			path:   "testdata/decode/invalid_mapping.json",
			anyErr: true,
		},
	}

	loader := NewLoader(LoaderConfig{}, NewRegexCache(), NewJSONPathCache(), NewScenarioHandler(nil))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping, err := loader.decodeMapping(tt.path)

			if err != nil {
				if !tt.anyErr {
					if tt.wantErr == nil {
						t.Log("did not expect an error, but got: ", err)
						t.FailNow()
					}
					require.Equal(t, tt.wantErr.Error(), err.Error())
				}
			}

			require.Equal(t, tt.wantMapping, mapping)
		})
	}
}

func TestLoadMappings(t *testing.T) {
	wantMappings := make(Mappings)
	for _, mapping := range validLoaderMappings {
		require.NoError(t, wantMappings.Put(mapping))
	}

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

	loader := NewLoader(LoaderConfig{}, NewRegexCache(), NewJSONPathCache(), NewScenarioHandler(nil))

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
					require.Equal(t, tt.wantErr, err.Error())
				}
				return
			}

			require.Equal(t, tt.wantMappings, mappings)
		})
	}

}
