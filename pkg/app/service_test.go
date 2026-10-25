package app

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockDelayer struct {
	FixedCalled bool
}

func (m *mockDelayer) Apply(delay *Delay) {
	if delay == nil {
		return
	}

	if delay.Fixed.Duration != 0 {
		m.FixedCalled = true
	}
}

func TestService(t *testing.T) {

	mappings := Mappings{
		"GET": []Mapping{
			{
				Request:  RequestMapping{Method: "GET", Path: CommonMatch{Exact: "/fixed/delay"}},
				Response: ResponseMapping{StatusCode: 204, ResponseDelay: Delay{FixedDelay{Duration: Duration(time.Millisecond * 10000)}}},
				MaxScore: 1,
				Cost:     0,
				FilePath: "file_1",
			},
			{
				Request:  RequestMapping{Method: "GET", Path: CommonMatch{Exact: "/no/delay"}},
				Response: ResponseMapping{StatusCode: 204},
				MaxScore: 1,
				Cost:     0,
				FilePath: "file_2",
			},
		},
	}

	tests := []struct {
		name       string
		request    Request
		wantResult MatchResult
		wantDelay  bool
	}{
		{
			name:       "Should match request with no delay",
			request:    Request{Method: "GET", Path: "/no/delay"},
			wantResult: MatchResult{StatusCode: 204, Matched: true, Headers: map[string]string{"X-Mapping-File": "file_2"}},
			wantDelay:  false,
		},
		{
			name:       "Should match request with fixed delay",
			request:    Request{Method: "GET", Path: "/fixed/delay"},
			wantResult: MatchResult{StatusCode: 204, Matched: true, Headers: map[string]string{"X-Mapping-File": "file_1"}},
			wantDelay:  true,
		},
	}

	matcher := NewMatcher(NewRegexCache(), NewJSONPathCache())

	for _, tt := range tests {
		delayer := mockDelayer{}
		service := NewService(mappings, matcher, NewScenarioHandler(matcher), &delayer)

		t.Run(tt.name, func(t *testing.T) {
			res := service.MatchRequest(tt.request)
			assert.Equal(t, tt.wantResult, res)
			assert.Equal(t, tt.wantDelay, delayer.FixedCalled)
		})
	}

}

func TestServiceConcurrentResponseHeaders(t *testing.T) {
	mapping := Mapping{
		Request:  RequestMapping{Method: "GET", Path: CommonMatch{Exact: "/headers"}},
		Response: ResponseMapping{StatusCode: 200, Headers: map[string]string{"Content-Type": "text/plain"}},
		FilePath: "headers.json",
	}
	mapping.CalcMaxScoreAndCost()
	mappings := make(Mappings)
	require.NoError(t, mappings.Put(mapping))
	matcher := NewMatcher(NewRegexCache(), NewJSONPathCache())
	service := NewService(mappings, matcher, NewScenarioHandler(matcher), &mockDelayer{})
	results := concurrentServiceRequests(service, Request{Method: "GET", Path: "/headers"}, 32)
	for _, result := range results {
		require.True(t, result.Matched)
		assert.Equal(t, "headers.json", result.Headers["X-Mapping-File"])
		assert.Equal(t, "text/plain", result.Headers["Content-Type"])
	}
	assert.Equal(t, map[string]string{"Content-Type": "text/plain"}, mapping.Response.Headers)
	results[0].Headers["Content-Type"] = "changed"
	assert.Equal(t, "text/plain", results[1].Headers["Content-Type"])
	assert.Equal(t, "text/plain", mapping.Response.Headers["Content-Type"])
}

func TestServiceConcurrentScenarioTransitions(t *testing.T) {
	matcher := NewMatcher(NewRegexCache(), NewJSONPathCache())
	handler := NewScenarioHandler(matcher)
	for _, scenario := range []ScenarioMapping{
		{Name: "cycle", State: "a", StartingState: true, NewState: "b"},
		{Name: "cycle", State: "b", NewState: "a"},
	} {
		mapping := Mapping{
			Scenario: &scenario,
			Request:  RequestMapping{Method: "GET", Path: CommonMatch{Exact: "/cycle"}},
			Response: ResponseMapping{StatusCode: 200, Body: scenario.State},
		}
		mapping.CalcMaxScoreAndCost()
		require.NoError(t, handler.AddScenario(mapping))
	}
	require.NoError(t, handler.ValidateScenarioStates())
	service := NewService(make(Mappings), matcher, handler, &mockDelayer{})
	const requests = 101
	results := concurrentServiceRequests(service, Request{Method: "GET", Path: "/cycle"}, requests)
	counts := map[any]int{}
	for _, result := range results {
		require.True(t, result.Matched)
		counts[result.Body]++
	}
	assert.Equal(t, map[any]int{"a": 51, "b": 50}, counts)
	assert.Equal(t, "b", handler.scenarios["cycle"].CurrentState)
	assert.Equal(t, "b", service.MatchRequest(Request{Method: "GET", Path: "/cycle"}).Body)
	assert.Equal(t, "a", handler.scenarios["cycle"].CurrentState)
}

func concurrentServiceRequests(service *Service, request Request, count int) []MatchResult {
	results := make([]MatchResult, count)
	start := make(chan struct{})
	var workers sync.WaitGroup
	for i := range results {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			results[i] = service.MatchRequest(request)
		}()
	}
	close(start)
	workers.Wait()
	return results
}
