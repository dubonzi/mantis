package app

import (
	"testing"
	"time"

	"github.com/go-playground/assert/v2"
)

type wantedDelay struct {
	fixed bool
}

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
		"GET": []*Mapping{
			{
				Request:  RequestMapping{Method: "GET", Path: CommonMatch{Exact: "/fixed/delay"}},
				Response: ResponseMapping{StatusCode: 204, ResponseDelay: Delay{FixedDelay{Duration: Duration(time.Millisecond * 10000)}}},
			},
			{
				Request:  RequestMapping{Method: "GET", Path: CommonMatch{Exact: "/no/delay"}},
				Response: ResponseMapping{StatusCode: 204},
			},
		},
	}

	tests := []struct {
		name       string
		request    Request
		wantResult MatchResult
		wantDelay  wantedDelay
	}{
		{
			name:       "Should match request with no delay",
			request:    Request{Method: "GET", Path: "/no/delay"},
			wantResult: MatchResult{StatusCode: 204, Matched: true},
			wantDelay:  wantedDelay{false},
		},
		{
			name:       "Should match request with fixed delay",
			request:    Request{Method: "GET", Path: "/fixed/delay"},
			wantResult: MatchResult{StatusCode: 204, Matched: true},
			wantDelay:  wantedDelay{true},
		},
	}

	matcher := NewMatcher(mappings, NewRegexCache(), NewJSONPathCache())

	for _, tt := range tests {
		delayer := mockDelayer{}
		service := NewService(matcher, &delayer)

		t.Run(tt.name, func(t *testing.T) {
			res := service.MatchRequest(tt.request)
			assert.Equal(t, res, tt.wantResult)
			assertDelay(tt.wantDelay, delayer)
		})
	}

}

func assertDelay(want wantedDelay, delayer mockDelayer) bool {
	return want.fixed == delayer.FixedCalled
}
