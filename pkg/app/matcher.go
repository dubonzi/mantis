package app

import (
	"strings"
)

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

func (matcher *Matcher) Match(r Request, mappings Mappings, scenarioStates map[string]ScenarioState) (Mapping, bool, bool) {
	methodMappings, ok := mappings[r.Method]
	if !ok {
		return Mapping{}, false, false
	}

	bestIndex, bestScore := -1, 0

	for i, mapping := range methodMappings {
		var score int

		if matcher.matchPath(r, mapping) {
			score += mapping.Request.PathScore()
		}

		if matcher.matchHeaders(r, mapping) {
			score += mapping.Request.HeaderScore()
		}

		if matcher.matchBody(r, mapping) {
			score += mapping.Request.BodyScore()
		}

		if score == mapping.MaxScore {
			if mapping.Scenario != nil {
				sc := scenarioStates[mapping.Scenario.Name]
				if sc.CurrentState != mapping.Scenario.State {
					continue
				}
			}
			return mapping, true, false
		}

		if score > bestScore {
			bestIndex = i
			bestScore = score
		}
	}

	if bestIndex >= 0 {
		return methodMappings[bestIndex], false, true
	}

	return Mapping{}, false, false
}

func (matcher *Matcher) matchPath(r Request, m Mapping) bool {
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

func (matcher *Matcher) matchHeaders(r Request, m Mapping) bool {
	for mKey, mVal := range m.Request.Headers {
		rVal, ok := r.Headers[strings.ToLower(mKey)]
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

func (matcher *Matcher) matchBody(r Request, m Mapping) bool {
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
		if !matcher.jsonPathCache.Match(m.Request.Body.JsonPath, r.Body) {
			return false
		}
	}

	return true
}
