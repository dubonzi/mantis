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

func (b *BasicMatcher) Match(r Request) MappingResponse {
	return MappingResponse{}
}

func (b *BasicMatcher) matchPath(r Request, m Mapping) bool {
	return false
}
