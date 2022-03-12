package app

type Matcher interface {
	Match(Request) MappingResponse
}

type BasicMatcher struct {
	mappings Mappings
}

func NewMatcher(m Mappings) *BasicMatcher {
	return &BasicMatcher{mappings: m}
}

func (m *BasicMatcher) Match(r Request) MappingResponse {
	return MappingResponse{}
}
