package app

type Matcher struct {
	loader Loader

	mappings map[string][]Mapping
}

func NewMatcher(loader Loader) *Matcher {
	loader.GetMappings()
	return &Matcher{loader: loader}
}