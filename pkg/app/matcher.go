package app

import (
	"encoding/json"
	"net/http"
	"strings"
)

var _ Matcher = new(BasicMatcher)

const (
	NoMappingFoundMessage = "No mapping found for the request"
)

type MatcherResult struct {
	StatusCode int
	Headers    map[string]string
	Body       any
	Matched    bool
}

func (m MatcherResult) String() string {
	s, _ := json.Marshal(m)
	return string(s)
}

type NotFoundResponse struct {
	Message        string          `json:"message"`
	Request        Request         `json:"request"`
	ClosestMapping *RequestMapping `json:"closestMapping,omitempty"`
}

type Matcher interface {
	Match(Request) (result MatcherResult)
}

type BasicMatcher struct {
	mappings Mappings
}

func NewMatcher(m Mappings) *BasicMatcher {
	return &BasicMatcher{mappings: m}
}

func (b *BasicMatcher) Match(r Request) MatcherResult {
	mapping, matched := b.match(r)

	result := MatcherResult{
		Matched: matched,
	}

	if mapping == nil {
		result.StatusCode = http.StatusNotFound
		result.Body = b.buildNotFoundResponse(r, nil)
		return result
	}

	if !matched {
		result.Body = b.buildNotFoundResponse(r, &mapping.Request)
		result.StatusCode = http.StatusNotFound
		return result
	}

	if mapping.Response.Body != "" {
		result.Body = mapping.Response.Body

	}
	result.StatusCode = mapping.Response.StatusCode
	result.Headers = mapping.Response.Headers

	return result
}

func (b *BasicMatcher) match(r Request) (*Mapping, bool) {
	methodMappings, ok := b.mappings[r.Method]
	if !ok {
		return nil, false
	}

	bestCandidate := [2]int{-1, 0} // index, score

	for i, mapping := range methodMappings {
		var score int

		if b.matchPath(r, mapping) {
			score++
		}

		if b.matchHeaders(r, mapping) && mapping.Request.HasHeaders() {
			score++
		}

		if b.matchBody(r, mapping) && mapping.Request.HasBody() {
			score++
		}

		if score == mapping.MaxScore() {
			return &mapping, true
		}

		if score > bestCandidate[1] {
			bestCandidate[0] = i
			bestCandidate[1] = score
		}
	}

	if bestCandidate[0] >= 0 {
		return &methodMappings[bestCandidate[0]], false
	}

	return nil, false
}

func (b *BasicMatcher) matchPath(r Request, m Mapping) bool {
	if m.Request.Path.Exact != "" {
		return r.Path == m.Request.Path.Exact
	}

	return true
}

func (b *BasicMatcher) matchHeaders(r Request, m Mapping) bool {
	for mKey, mVal := range m.Request.Headers {
		if mVal.Exact != "" {
			rVal, ok := r.Headers[strings.ToLower(mKey)]
			if !ok {
				return false
			}
			if rVal != mVal.Exact {
				return false
			}
		}
	}

	return true
}

func (b *BasicMatcher) matchBody(r Request, m Mapping) bool {
	if m.Request.Body.Exact != "" {
		return r.Body == m.Request.Body.Exact
	}
	return true
}

func (b *BasicMatcher) buildNotFoundResponse(r Request, mapping *RequestMapping) NotFoundResponse {
	return NotFoundResponse{
		Message:        NoMappingFoundMessage,
		Request:        r,
		ClosestMapping: mapping,
	}
}
