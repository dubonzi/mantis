package app

import (
	"fmt"
	"sync"
	"testing"

	"github.com/americanas-go/log"
	"github.com/stretchr/testify/require"
)

type parsingLog struct {
	log.Logger
	errors int
}

func (l *parsingLog) Errorf(format string, args ...interface{}) {
	l.errors++
}

func captureParsingLog(t *testing.T) *parsingLog {
	t.Helper()
	previous := log.GetLogger()
	logger := &parsingLog{Logger: previous}
	log.SetGlobalLogger(logger)
	t.Cleanup(func() { log.SetGlobalLogger(previous) })
	return logger
}

func TestMatcherLazyJSONBody(t *testing.T) {
	tests := []struct {
		name       string
		first      RequestMapping
		wantErrors int
	}{
		{name: "unrelated exact path", first: RequestMapping{Method: "POST", Path: CommonMatch{Exact: "/other"}, Body: BodyMatch{JsonPath: []string{"$.id"}}}},
		{name: "header rejects candidate", first: RequestMapping{Method: "POST", Path: CommonMatch{Exact: "/json"}, Headers: map[string]CommonMatch{"x-token": {Exact: "required"}}, Body: BodyMatch{JsonPath: []string{"$.id"}}}},
		{name: "body contains rejects candidate", first: RequestMapping{Method: "POST", Path: CommonMatch{Exact: "/json"}, Body: BodyMatch{CommonMatch: CommonMatch{Contains: []string{"absent"}}, JsonPath: []string{"$.id"}}}},
		{name: "repeated invalid body logs once", first: RequestMapping{Method: "POST", Path: CommonMatch{Exact: "/json"}, Body: BodyMatch{JsonPath: []string{"$.id", "$.name"}}}, wantErrors: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := captureParsingLog(t)
			matcher, mappings := prepareMatcher(t, []Mapping{
				{Request: tt.first},
				{Request: tt.first},
				{Request: RequestMapping{Method: "POST", Path: CommonMatch{Exact: "/json"}}, FilePath: "fallback"},
			})
			result := matcher.Match(Request{Method: "POST", Path: "/json", Body: "invalid json"}, mappings, nil)
			require.Equal(t, FullMatch, result.Kind)
			require.Equal(t, "fallback", result.Mapping.FilePath)
			require.Equal(t, tt.wantErrors, logger.errors)
		})
	}
}

func TestServiceSharesJSONBodyAcrossScenarioAndOrdinaryMappings(t *testing.T) {
	logger := captureParsingLog(t)
	entry := Mapping{Request: RequestMapping{Method: "POST", Path: CommonMatch{Exact: "/json"}, Body: BodyMatch{JsonPath: []string{"$.id", "$.name"}}}}
	matcher, mappings := prepareMatcher(t, []Mapping{entry, entry})
	handler := NewScenarioHandler(matcher)
	entry.Scenario = &ScenarioMapping{Name: "json", State: "ready", StartingState: true}
	require.NoError(t, handler.AddScenario(entry))
	service := NewService(mappings, matcher, handler, &mockDelayer{})
	for i := 1; i <= 2; i++ {
		result := service.MatchRequest(Request{Method: "POST", Path: "/json", Body: "invalid json"})
		require.False(t, result.Matched)
		require.Equal(t, 404, result.StatusCode)
		require.Equal(t, i, logger.errors)
	}
}

func TestMatcherConcurrentJSONBodies(t *testing.T) {
	matcher, mappings := prepareMatcher(t, []Mapping{
		{Request: RequestMapping{Method: "POST", Path: CommonMatch{Exact: "/json"}, Body: BodyMatch{JsonPath: []string{"$.a", "$.name"}}}, FilePath: "a"},
		{Request: RequestMapping{Method: "POST", Path: CommonMatch{Exact: "/json"}, Body: BodyMatch{JsonPath: []string{"$.b", "$.name"}}}, FilePath: "b"},
	})
	const count = 64
	results := make([]MatchOutcome, count)
	var workers sync.WaitGroup
	for i := range results {
		workers.Add(1)
		go func() {
			defer workers.Done()
			key := "a"
			if i%2 == 1 {
				key = "b"
			}
			results[i] = matcher.Match(Request{Method: "POST", Path: "/json", Body: fmt.Sprintf(`{"%s":true,"name":"Ada"}`, key)}, mappings, nil)
		}()
	}
	workers.Wait()
	for i, result := range results {
		want := "a"
		if i%2 == 1 {
			want = "b"
		}
		require.Equal(t, FullMatch, result.Kind)
		require.Equal(t, want, result.Mapping.FilePath)
	}
}

func TestMatcherNormalizedHeadersPreserveConstraints(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string]CommonMatch
		want    MatchKind
	}{
		{"case insensitive", map[string]CommonMatch{"X-Token": {Exact: "abc"}}, FullMatch},
		{"duplicate case variants both match", map[string]CommonMatch{"X-Token": {Contains: []string{"a"}}, "x-token": {Exact: "abc"}}, FullMatch},
		{"duplicate case variants both required", map[string]CommonMatch{"X-Token": {Exact: "different"}, "x-token": {Exact: "abc"}}, PartialMatch},
		{"zero score header does not prevent full score", map[string]CommonMatch{"missing": {}}, FullMatch},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matcher, mappings := prepareMatcher(t, []Mapping{{Request: RequestMapping{Method: "GET", Path: CommonMatch{Exact: "/headers"}, Headers: tt.headers}}})
			result := matcher.Match(Request{Method: "GET", Path: "/headers", Headers: map[string]string{"x-token": "abc"}}, mappings, nil)
			require.Equal(t, tt.want, result.Kind)
			require.Equal(t, tt.headers, result.Mapping.Request.Headers)
		})
	}
}
