package app

type Matcher struct {
	loader Loader

	mappings Mappings
}

func NewMatcher(loader Loader) *Matcher {
	return &Matcher{loader: loader, mappings: loader.GetMappings()}
}
