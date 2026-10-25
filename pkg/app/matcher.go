package app

import (
	"strings"
)

type MatchKind int

const (
	NoMatch MatchKind = iota
	PartialMatch
	FullMatch
)

type MatchOutcome struct {
	Mapping Mapping
	Kind    MatchKind
}

type Matcher struct {
	regexCache    *RegexCache
	jsonPathCache *JSONPathCache
}

func NewMatcher(r *RegexCache, j *JSONPathCache) *Matcher {
	return &Matcher{
		regexCache:    r,
		jsonPathCache: j,
	}
}

func (matcher *Matcher) Match(r Request, mappings Mappings, scenarioStates map[string]ScenarioState) MatchOutcome {
	var body jsonBody
	return matcher.match(r, mappings, scenarioStates, &body)
}

func (matcher *Matcher) match(r Request, mappings Mappings, scenarioStates map[string]ScenarioState, body *jsonBody) MatchOutcome {
	method := mappings[r.Method]
	if method == nil {
		return MatchOutcome{}
	}

	exact, nonExact := method.exact[r.Path], method.nonExact
	for len(exact) > 0 || len(nonExact) > 0 {
		var index int
		if len(exact) > 0 && (len(nonExact) == 0 || exact[0] < nonExact[0]) {
			index, exact = exact[0], exact[1:]
		} else {
			index, nonExact = nonExact[0], nonExact[1:]
		}
		mapping := &method.entries[index]
		if mapping.Scenario != nil && scenarioStates[mapping.Scenario.Name].CurrentState != mapping.Scenario.State {
			continue
		}
		if (mapping.pathScore == 0 || matcher.matchPath(r, mapping)) &&
			(mapping.headerScore == 0 || matcher.matchHeaders(r, mapping)) &&
			(mapping.bodyScore == 0 || matcher.matchBody(r, mapping, body)) {
			return MatchOutcome{Mapping: *mapping, Kind: FullMatch}
		}
	}

	bestIndex, bestScore := -1, 0
	for i := range method.entries {
		mapping := &method.entries[i]
		score := 0
		if matcher.matchPath(r, mapping) {
			score += mapping.pathScore
		}
		if mapping.headerScore > 0 && matcher.matchHeaders(r, mapping) {
			score += mapping.headerScore
		}
		if mapping.bodyScore > 0 && matcher.matchBody(r, mapping, body) {
			score += mapping.bodyScore
		}
		if score == mapping.MaxScore {
			continue
		}
		if score > bestScore {
			bestIndex, bestScore = i, score
		}
	}
	if bestIndex >= 0 {
		return MatchOutcome{Mapping: method.entries[bestIndex], Kind: PartialMatch}
	}
	return MatchOutcome{}
}

func (matcher *Matcher) matchPath(r Request, m *Mapping) bool {
	if m.Request.Path.Exact != "" {
		return r.Path == m.Request.Path.Exact
	}

	for _, c := range m.Request.Path.Contains {
		if !strings.Contains(r.Path, c) {
			return false
		}
	}

	for _, p := range m.Request.Path.Patterns {
		if !matcher.regexCache.Match(p, r.Path) {
			return false
		}
	}

	return true
}

func (matcher *Matcher) matchHeaders(r Request, m *Mapping) bool {
	for i := range m.headers {
		header := &m.headers[i]
		mVal := &header.match
		rVal, ok := r.Headers[header.key]
		if !ok {
			return false
		}

		if mVal.Exact != "" {
			if rVal != mVal.Exact {
				return false
			}
		}

		for _, c := range mVal.Contains {
			if !strings.Contains(rVal, c) {
				return false
			}
		}

		for _, p := range mVal.Patterns {
			if !matcher.regexCache.Match(p, rVal) {
				return false
			}
		}

	}

	return true
}

func (matcher *Matcher) matchBody(r Request, m *Mapping, body *jsonBody) bool {
	if m.Request.Body.Exact != "" {
		return r.Body == m.Request.Body.Exact
	}

	for _, c := range m.Request.Body.Contains {
		if !strings.Contains(r.Body, c) {
			return false
		}
	}

	for _, p := range m.Request.Body.Patterns {
		if !matcher.regexCache.Match(p, r.Body) {
			return false
		}
	}

	if len(m.Request.Body.JsonPath) > 0 {
		if !matcher.jsonPathCache.match(m.Request.Body.JsonPath, r.Body, body) {
			return false
		}
	}

	return true
}
