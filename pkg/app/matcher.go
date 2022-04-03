package app

import (
	"strings"
)

var _ Matcher = new(BasicMatcher)

type Matcher interface {
	Match(Request) (mapping *Mapping, matched bool)
}

type BasicMatcher struct {
	mappings      Mappings
	regexCache    *RegexCache
	jsonPathCache *JSONPathCache
}

func NewMatcher(m Mappings, r *RegexCache, j *JSONPathCache) *BasicMatcher {
	return &BasicMatcher{
		mappings:      m,
		regexCache:    r,
		jsonPathCache: j,
	}
}

func (b *BasicMatcher) Match(r Request) (*Mapping, bool) {
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

		if b.matchHeaders(r, mapping) {
			score += mapping.Request.HeaderScore()
		}

		if b.matchBody(r, mapping) {
			score += mapping.Request.BodyScore()
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

	if len(m.Request.Path.Contains) > 0 {
		for _, c := range m.Request.Path.Contains {
			if !strings.Contains(r.Path, c) {
				return false
			}
		}
	}

	if len(m.Request.Path.Pattern) > 0 {
		for _, p := range m.Request.Path.Pattern {
			if !b.regexCache.Match(p, r.Path) {
				return false
			}
		}
	}

	return true
}

func (b *BasicMatcher) matchHeaders(r Request, m Mapping) bool {
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

		if mVal.Contains != "" {
			if !strings.Contains(rVal, mVal.Contains) {
				return false
			}
		}

		if mVal.Pattern != "" {
			if !b.regexCache.Match(mVal.Pattern, rVal) {
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

	if m.Request.Body.Contains != "" {
		return strings.Contains(r.Body, m.Request.Body.Contains)
	}

	if m.Request.Body.Pattern != "" {
		return b.regexCache.Match(m.Request.Body.Pattern, r.Body)
	}

	if len(m.Request.Body.JsonPath) > 0 {
		return b.jsonPathCache.Match(m.Request.Body.JsonPath, r.Body)
	}

	return true
}
