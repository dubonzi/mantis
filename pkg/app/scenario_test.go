package app

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddValidScenarios(t *testing.T) {
	tests := []struct {
		name  string
		input []Mapping
		want  map[string]ScenarioState
	}{
		{
			name:  "loads first scenario with 3 states",
			input: validScenarios["firstScenario"],
			want: map[string]ScenarioState{
				"First Scenario": {
					CurrentState: "Object Exists",
					States:       getMappingsMap(validScenarios["firstScenario"]),
				},
			},
		},
		{
			name:  "loads second scenario with 2 states",
			input: validScenarios["secondScenario"],
			want: map[string]ScenarioState{
				"Second Scenario": {
					CurrentState: "Create Object",
					States:       getMappingsMap(validScenarios["secondScenario"]),
				},
			},
		},
		{
			name:  "loads both scenarios",
			input: append(validScenarios["firstScenario"], validScenarios["secondScenario"]...),
			want: map[string]ScenarioState{
				"First Scenario": {
					CurrentState: "Object Exists",
					States:       getMappingsMap(validScenarios["firstScenario"]),
				},
				"Second Scenario": {
					CurrentState: "Create Object",
					States:       getMappingsMap(validScenarios["secondScenario"]),
				},
			},
		},
	}

	for _, tt := range tests {
		handler := NewScenarioHandler(nil)
		t.Run(tt.name, func(t *testing.T) {
			for _, m := range tt.input {
				require.NoError(t, handler.AddScenario(m))
			}

			assert.Equal(t, tt.want, handler.scenarios)
			assert.NoError(t, handler.ValidateScenarioStates())
		})
	}
}

func TestValidateScenarios(t *testing.T) {
	tests := []struct {
		name  string
		input []Mapping
		want  ScenarioValidationErrors
	}{
		{
			name:  "validates scenario with no starting state",
			input: invalidScenarios["noStartingState"],
			want: ScenarioValidationErrors{
				{
					ScenarioName: "No Start",
					Message:      "the scenario has no starting state defined",
				},
			},
		},
		{
			name:  "validates scenario with multiple starting states",
			input: invalidScenarios["multipleStartingStates"],
			want: ScenarioValidationErrors{
				{
					ScenarioName: "Multiple Start",
					Message:      "the scenario has multiple starting states defined",
				},
			},
		},
		{
			name:  "validates scenario with a single state",
			input: invalidScenarios["singleState"],
			want: ScenarioValidationErrors{
				{
					ScenarioName: "Single State",
					Message:      "the scenario must have at least 2 defined states",
				},
			},
		},
		{
			name:  "validates scenario with invalid state names",
			input: invalidScenarios["invalidStateName"],
			want: ScenarioValidationErrors{
				{
					ScenarioName: "State Not Found",
					Message:      "the scenario has a state pointing to a new state that is not defined in the scenario: [First -> Non existent]",
				},
			},
		},
		{
			name:  "validates scenario with multiple errors",
			input: invalidScenarios["multipleErrors"],
			want: ScenarioValidationErrors{
				{
					ScenarioName: "Multiple Errors",
					Message:      "the scenario has a state pointing to a new state that is not defined in the scenario: [First -> Non existent]",
				}, {
					ScenarioName: "Multiple Errors",
					Message:      "the scenario has no starting state defined",
				}, {
					ScenarioName: "Multiple Errors",
					Message:      "the scenario must have at least 2 defined states",
				},
			},
		},
	}

	for _, tt := range tests {
		handler := NewScenarioHandler(nil)
		t.Run(tt.name, func(t *testing.T) {
			for _, m := range tt.input {
				require.NoError(t, handler.AddScenario(m))
			}

			err := handler.ValidateScenarioStates()
			require.Error(t, err)
			assert.Equal(t, tt.want, err)
		})
	}

}

func TestScenarioMatching(t *testing.T) {
	type scenarioCase struct {
		request  Request
		expected MatchResult
	}

	tests := []struct {
		name     string
		mappings []Mapping
		cases    []scenarioCase
	}{
		{
			name:     "should follow state progression for first scenario",
			mappings: validScenarios["firstScenario"],
			cases: []scenarioCase{
				{
					request:  Request{Method: "DELETE", Path: "/scenario/123"},
					expected: MatchResult{StatusCode: 204, Headers: map[string]string{"X-Mapping-File": "scenario1_1"}, Matched: true},
				},
				{
					request:  Request{Method: "DELETE", Path: "/scenario/123"},
					expected: MatchResult{StatusCode: 404, Headers: map[string]string{"X-Mapping-File": "scenario1_2"}, Matched: true},
				},
				{
					request:  Request{Method: "GET", Path: "/scenario/123"},
					expected: MatchResult{StatusCode: 404, Headers: map[string]string{"X-Mapping-File": "scenario1_3"}, Matched: true},
				},
			},
		},
		{
			name:     "should follow state progression for second scenario",
			mappings: validScenarios["secondScenario"],
			cases: []scenarioCase{
				{
					request: Request{Method: "GET", Path: "/objects/123"},
					expected: MatchResult{
						StatusCode: 404,
						Matched:    false,
						Headers:    map[string]string{"Content-type": "application/json"},
						Body:       NotFoundResponse{Message: "No mapping found for the request", Request: Request{Path: "/objects/123", Method: "GET"}},
					},
				},
				{
					request:  Request{Method: "POST", Path: "/objects"},
					expected: MatchResult{StatusCode: 201, Headers: map[string]string{"Location": "/objects/123", "X-Mapping-File": "scenario2_1"}, Matched: true},
				},
				{
					request:  Request{Method: "GET", Path: "/objects/123"},
					expected: MatchResult{StatusCode: 200, Body: "{\"id\": 123}", Headers: map[string]string{"X-Mapping-File": "scenario2_2"}, Matched: true},
				},
			},
		},
	}

	for _, tt := range tests {
		matcher := NewMatcher(NewRegexCache(), NewJSONPathCache())
		handler := NewScenarioHandler(matcher)
		for _, m := range tt.mappings {
			require.NoError(t, handler.AddScenario(m))
		}

		t.Run(tt.name, func(t *testing.T) {
			err := handler.ValidateScenarioStates()
			require.NoError(t, err)

			for i, c := range tt.cases {
				t.Run(fmt.Sprintf("%d", i+1), func(t *testing.T) {
					match := handler.MatchScenario(c.request)
					got := NewMatchResult(match, c.request)
					assert.Equal(t, c.expected.Matched, match.Kind == FullMatch)
					assert.Equal(t, c.expected, got)
				})
			}
		})
	}
}

func getMappingsMap(mappings []Mapping) map[string][]Mapping {
	res := make(map[string][]Mapping)
	for _, m := range mappings {
		res[m.Scenario.State] = append(res[m.Scenario.State], m)
	}
	return res
}

var invalidScenarios = map[string][]Mapping{
	"noStartingState": {
		{
			Scenario: &ScenarioMapping{Name: "No Start", State: "First", NewState: "Second"},
			Request:  RequestMapping{Method: "GET", Path: CommonMatch{Exact: "/first"}},
			Response: ResponseMapping{StatusCode: 200},
		}, {
			Scenario: &ScenarioMapping{Name: "No Start", State: "Second"},
			Request:  RequestMapping{Method: "GET", Path: CommonMatch{Exact: "/first"}},
			Response: ResponseMapping{StatusCode: 404},
		},
	},
	"multipleStartingStates": {
		{
			Scenario: &ScenarioMapping{Name: "Multiple Start", StartingState: true, State: "First", NewState: "Second"},
			Request:  RequestMapping{Method: "GET", Path: CommonMatch{Exact: "/first"}},
			Response: ResponseMapping{StatusCode: 200},
		}, {
			Scenario: &ScenarioMapping{Name: "Multiple Start", StartingState: true, State: "Second"},
			Request:  RequestMapping{Method: "GET", Path: CommonMatch{Exact: "/first"}},
			Response: ResponseMapping{StatusCode: 404},
		},
	},
	"singleState": {
		{
			Scenario: &ScenarioMapping{Name: "Single State", StartingState: true, State: "First"},
			Request:  RequestMapping{Method: "GET", Path: CommonMatch{Exact: "/first"}},
			Response: ResponseMapping{StatusCode: 200},
		},
	},
	"invalidStateName": {
		{
			Scenario: &ScenarioMapping{Name: "State Not Found", StartingState: true, State: "First", NewState: "Non existent"},
			Request:  RequestMapping{Method: "GET", Path: CommonMatch{Exact: "/first"}},
			Response: ResponseMapping{StatusCode: 200},
		}, {
			Scenario: &ScenarioMapping{Name: "State Not Found", State: "Second"},
			Request:  RequestMapping{Method: "GET", Path: CommonMatch{Exact: "/first"}},
			Response: ResponseMapping{StatusCode: 200},
		},
	},
	"multipleErrors": {
		{
			Scenario: &ScenarioMapping{Name: "Multiple Errors", State: "First", NewState: "Non existent"},
			Request:  RequestMapping{Method: "GET", Path: CommonMatch{Exact: "/first"}},
			Response: ResponseMapping{StatusCode: 200},
		},
	},
}

var validScenarios = map[string][]Mapping{
	"firstScenario": {
		{
			Scenario: &ScenarioMapping{Name: "First Scenario", StartingState: true, State: "Object Exists", NewState: "Object Deleted"},
			Request:  RequestMapping{Method: "DELETE", Path: CommonMatch{Exact: "/scenario/123"}},
			Response: ResponseMapping{StatusCode: 204},
			MaxScore: 1,
			Cost:     0,
			FilePath: "scenario1_1",
		}, {
			Scenario: &ScenarioMapping{Name: "First Scenario", State: "Object Deleted", NewState: "Get Deleted Object"},
			Request:  RequestMapping{Method: "DELETE", Path: CommonMatch{Exact: "/scenario/123"}},
			Response: ResponseMapping{StatusCode: 404},
			MaxScore: 1,
			Cost:     0,
			FilePath: "scenario1_2",
		}, {
			Scenario: &ScenarioMapping{Name: "First Scenario", State: "Get Deleted Object"},
			Request:  RequestMapping{Method: "GET", Path: CommonMatch{Exact: "/scenario/123"}},
			Response: ResponseMapping{StatusCode: 404},
			MaxScore: 1,
			Cost:     0,
			FilePath: "scenario1_3",
		},
	},
	"secondScenario": {
		{
			Scenario: &ScenarioMapping{Name: "Second Scenario", StartingState: true, State: "Create Object", NewState: "Object Created"},
			Request:  RequestMapping{Method: "POST", Path: CommonMatch{Exact: "/objects"}},
			Response: ResponseMapping{StatusCode: 201, Headers: map[string]string{"Location": "/objects/123"}},
			MaxScore: 1,
			Cost:     0,
			FilePath: "scenario2_1",
		}, {
			Scenario: &ScenarioMapping{Name: "Second Scenario", State: "Object Created"},
			Request:  RequestMapping{Method: "GET", Path: CommonMatch{Exact: "/objects/123"}},
			Response: ResponseMapping{StatusCode: 200, Body: "{\"id\": 123}"},
			MaxScore: 1,
			Cost:     0,
			FilePath: "scenario2_2",
		},
	},
}

func TestScenarioMappingsShareState(t *testing.T) {
	for _, path := range []string{"/one", "/two"} {
		t.Run(path, func(t *testing.T) {
			matcher := NewMatcher(NewRegexCache(), NewJSONPathCache())
			handler := NewScenarioHandler(matcher)
			for _, mapping := range []Mapping{
				{Scenario: &ScenarioMapping{Name: "shared", State: "a", StartingState: true, NewState: "b"}, Request: RequestMapping{Method: "GET", Path: CommonMatch{Exact: "/one"}}, Response: ResponseMapping{StatusCode: 201, Body: "one"}},
				{Scenario: &ScenarioMapping{Name: "shared", State: "a", StartingState: true}, Request: RequestMapping{Method: "GET", Path: CommonMatch{Exact: "/two"}}, Response: ResponseMapping{StatusCode: 202, Body: "two"}},
				{Scenario: &ScenarioMapping{Name: "shared", State: "b", NewState: "a"}, Request: RequestMapping{Method: "GET", Path: CommonMatch{Exact: "/reset"}}, Response: ResponseMapping{StatusCode: 204}},
			} {
				mapping.CalcMaxScoreAndCost()
				require.NoError(t, handler.AddScenario(mapping))
			}
			require.NoError(t, handler.ValidateScenarioStates())
			service := NewService(make(Mappings), matcher, handler, &mockDelayer{})
			result := service.MatchRequest(Request{Method: "GET", Path: path})
			require.True(t, result.Matched)
			if path == "/one" {
				assert.Equal(t, 201, result.StatusCode)
				assert.Equal(t, "one", result.Body)
				assert.Equal(t, "b", handler.scenarios["shared"].CurrentState)
			} else {
				assert.Equal(t, 202, result.StatusCode)
				assert.Equal(t, "two", result.Body)
				assert.Equal(t, "a", handler.scenarios["shared"].CurrentState)
			}
		})
	}
}

func TestLoaderRejectsInvalidScenarioMapping(t *testing.T) {
	directory := t.TempDir()
	path := filepath.Join(directory, "invalid-scenario.json")
	require.NoError(t, os.WriteFile(path, []byte(`[
 {"scenario":{"name":"invalid","state":"a","startingState":true,"newState":"b"},"request":{"path":{"exact":"/invalid"}}},
 {"scenario":{"name":"invalid","state":"b","newState":"a"},"request":{"method":"GET","path":{"exact":"/valid"}}}
 ]`), 0600))
	handler := NewScenarioHandler(NewMatcher(NewRegexCache(), NewJSONPathCache()))
	loader := NewLoader(LoaderConfig{MappingPath: directory, ResponsePath: directory}, NewRegexCache(), NewJSONPathCache(), handler)
	err := loader.loadMappings(directory, directory, make(Mappings))
	require.Error(t, err)
	assert.Contains(t, err.Error(), path)
	assert.Contains(t, err.Error(), "Method is required")
	assert.Empty(t, handler.scenarios)
	assert.Empty(t, handler.scenarioMappings)
}

func TestValidateScenarioSharedStates(t *testing.T) {
	tests := []struct {
		name   string
		states []ScenarioMapping
		want   ScenarioValidationErrors
	}{
		{name: "several mappings in the same starting state", states: []ScenarioMapping{
			{State: "a", StartingState: true, NewState: "b"},
			{State: "a", StartingState: true},
			{State: "b", NewState: "a"},
		}},
		{name: "several mappings still count as one distinct state", states: []ScenarioMapping{
			{State: "a", StartingState: true}, {State: "a", StartingState: true},
		}, want: ScenarioValidationErrors{{"shared", ScenarioSingleStateMessage}}},
		{name: "validate every mapping transition", states: []ScenarioMapping{
			{State: "a", StartingState: true, NewState: "missing"},
			{State: "a", StartingState: true, NewState: "b"},
			{State: "b", NewState: "a"},
		}, want: ScenarioValidationErrors{{"shared", "the scenario has a state pointing to a new state that is not defined in the scenario: [a -> missing]"}}},
		{name: "conflicting starting flags true then false", states: []ScenarioMapping{
			{State: "a", StartingState: true}, {State: "a"}, {State: "b"},
		}, want: ScenarioValidationErrors{{"shared", "the scenario has mappings with inconsistent startingState values for state: [a]"}}},
		{name: "conflicting starting flags false then true", states: []ScenarioMapping{
			{State: "a"}, {State: "a", StartingState: true}, {State: "b"},
		}, want: ScenarioValidationErrors{{"shared", "the scenario has mappings with inconsistent startingState values for state: [a]"}}},
		{name: "multiple distinct starting states with shared mappings", states: []ScenarioMapping{
			{State: "a", StartingState: true}, {State: "a", StartingState: true}, {State: "b", StartingState: true},
		}, want: ScenarioValidationErrors{{"shared", ScenarioMultipleStartingStateMessage}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewScenarioHandler(nil)
			for _, scenario := range tt.states {
				scenario.Name = "shared"
				require.NoError(t, handler.AddScenario(Mapping{Scenario: &scenario, Request: RequestMapping{Method: "GET", Path: CommonMatch{Exact: "/shared"}}}))
			}
			err := handler.ValidateScenarioStates()
			if tt.want == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				var errors ScenarioValidationErrors
				require.ErrorAs(t, err, &errors)
				assert.ElementsMatch(t, tt.want, errors)
			}
		})
	}
}

func TestAddScenarioRejectsInvalidMappingWithoutMutation(t *testing.T) {
	handler := NewScenarioHandler(nil)
	valid := Mapping{
		Scenario: &ScenarioMapping{Name: "unchanged", State: "a", StartingState: true},
		Request:  RequestMapping{Method: "GET", Path: CommonMatch{Exact: "/valid"}},
	}
	require.NoError(t, handler.AddScenario(valid))
	invalid := Mapping{
		Scenario: &ScenarioMapping{Name: "unchanged", State: "b", StartingState: true},
		Request:  RequestMapping{Path: CommonMatch{Exact: "/invalid"}},
	}
	require.Error(t, handler.AddScenario(invalid))
	assert.Equal(t, map[string]ScenarioState{
		"unchanged": {CurrentState: "a", States: map[string][]Mapping{"a": {valid}}},
	}, handler.scenarios)
	assert.Equal(t, Mappings{"GET": {valid}}, handler.scenarioMappings)
}
