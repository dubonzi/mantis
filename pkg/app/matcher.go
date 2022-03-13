package app

// TODO: Cache compiled regex's

type Matcher interface {
	Match(Request) (response MappingResponse, matched bool)
}

type BasicMatcher struct {
	mappings Mappings
}

func NewMatcher(m Mappings) *BasicMatcher {
	return &BasicMatcher{mappings: m}
}

func (b *BasicMatcher) Match(r Request) (*MappingResponse, bool) {
	methodMappings, ok := b.mappings[r.Method]
	if !ok {
		return &MappingResponse{}, false
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
			return &mapping.Response, true
		}

		if score > bestCandidate[1] {
			bestCandidate[0] = i
			bestCandidate[1] = score
		}
	}

	if bestCandidate[0] >= 0 {
		return &methodMappings[bestCandidate[0]].Response, false
	}

	return &MappingResponse{}, false
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
			rVal, ok := r.Headers[mKey]
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
