package app

import (
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
				handler.AddScenario(m)
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
					Message:      "the scenario has multiple starting states defined: [First, Second]",
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
				handler.AddScenario(m)
			}

			err := handler.ValidateScenarioStates()
			require.Error(t, err)
			assert.Equal(t, tt.want, err)
		})
	}

}

// func TestScenarioMatching(t *testing.T) {
// 	type scenarioCase struct {
// 		request  Request
// 		expected MatchResult
// 	}

// 	tests := []struct {
// 		name     string
// 		mappings []Mapping
// 		cases    []scenarioCase
// 	}{
// 		{
// 			name:     "should follow state progression",
// 			mappings: validScenarios["firstScenario"],
// 			cases: []scenarioCase{
// 				{
// 					request:  Request{Method: "DELETE", Path: "/scenario/123"},
// 					expected: MatchResult{StatusCode: 204, Matched: true},
// 				},
// 				{
// 					request:  Request{Method: "DELETE", Path: "/scenario/123"},
// 					expected: MatchResult{StatusCode: 404, Matched: true},
// 				},
// 				{
// 					request:  Request{Method: "GET", Path: "/scenario/123"},
// 					expected: MatchResult{StatusCode: 404, Matched: true},
// 				},
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		handler := NewScenarioHandler()
// 		mappings := make(Mappings)
// 		for _, m := range tt.mappings {
// 			handler.AddScenario(m)
// 			mappings.Put(m)
// 		}

// 		t.Run(tt.name, func(t *testing.T) {
// 			err := handler.ValidateScenarioStates()
// 			require.NoError(t, err)

// 			matcher := NewMatcher(NewRegexCache(), NewJSONPathCache(), handler)
// 			for _, c := range tt.cases {
// 				matched, ok, _ := matcher.Match(c.request, mappings)
// 				got := NewMatchResult(&matched, c.request, ok, false)
// 				assert.True(t, ok)
// 				assert.Equal(t, c.expected, got)
// 			}
// 		})
// 	}
// }

func getMappingsMap(mappings []Mapping) map[string]Mapping {
	res := make(map[string]Mapping)
	for _, m := range mappings {
		res[m.Scenario.State] = m
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
		}, {
			Scenario: &ScenarioMapping{Name: "First Scenario", State: "Object Deleted", NewState: "Get Deleted Object"},
			Request:  RequestMapping{Method: "DELETE", Path: CommonMatch{Exact: "/scenario/123"}},
			Response: ResponseMapping{StatusCode: 404},
			MaxScore: 1,
		}, {
			Scenario: &ScenarioMapping{Name: "First Scenario", State: "Get Deleted Object"},
			Request:  RequestMapping{Method: "GET", Path: CommonMatch{Exact: "/scenario/123"}},
			Response: ResponseMapping{StatusCode: 404},
			MaxScore: 1,
		},
	},
	"secondScenario": {
		{
			Scenario: &ScenarioMapping{Name: "Second Scenario", StartingState: true, State: "Create Object", NewState: "Object Created"},
			Request:  RequestMapping{Method: "POST", Path: CommonMatch{Exact: "/objects"}},
			Response: ResponseMapping{StatusCode: 201, Headers: map[string]string{"Location": "123"}},
			MaxScore: 1,
		}, {
			Scenario: &ScenarioMapping{Name: "Second Scenario", State: "Object Created"},
			Request:  RequestMapping{Method: "GET", Path: CommonMatch{Exact: "/objects/123"}},
			Response: ResponseMapping{StatusCode: 200, Body: "{\"id\": 123}"},
			MaxScore: 1,
		},
	},
}
