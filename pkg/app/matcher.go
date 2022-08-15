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

func (b *Matcher) Match(r Request, mappings Mappings) (*Mapping, bool) {
	methodMappings, ok := mappings[r.Method]
	if !ok {
		return nil, false
	}

	bestIndex, bestScore := -1, 0

	for i, mapping := range methodMappings {
		var score int

		if b.matchPath(r, mapping) {
			score += mapping.Request.PathScore()
		}

		if b.matchHeaders(r, mapping) {
			score += mapping.Request.HeaderScore()
		}

		if b.matchBody(r, mapping) {
			score += mapping.Request.BodyScore()
		}

		if score == mapping.MaxScore {
			return mapping, true
		}

		if score > bestScore {
			bestIndex = i
			bestScore = score
		}
	}

	if bestIndex >= 0 {
		return methodMappings[bestIndex], false
	}

	return nil, false
}

func (b *Matcher) matchPath(r Request, m *Mapping) bool {
	if m.Request.Path.Exact != "" {
		return r.Path == m.Request.Path.Exact
	}

	for _, c := range m.Request.Path.Contains {
		if !strings.Contains(r.Path, c) {
			return false
		}
	}

	for _, p := range m.Request.Path.Patterns {
		if !b.regexCache.Match(p, r.Path) {
			return false
		}
	}

	return true
}

func (b *Matcher) matchHeaders(r Request, m *Mapping) bool {
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
			if !b.regexCache.Match(p, rVal) {
				return false
			}
		}

	}

	return true
}

func (b *Matcher) matchBody(r Request, m *Mapping) bool {
	if m.Request.Body.Exact != "" {
		return r.Body == m.Request.Body.Exact
	}

	for _, c := range m.Request.Body.Contains {
		if !strings.Contains(r.Body, c) {
			return false
		}
	}

	for _, p := range m.Request.Body.Patterns {
		if !b.regexCache.Match(p, r.Body) {
			return false
		}
	}

	if len(m.Request.Body.JsonPath) > 0 {
		if !b.jsonPathCache.Match(m.Request.Body.JsonPath, r.Body) {
			return false
		}
	}

	return true
}
