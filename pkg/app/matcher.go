package app

import "github.com/americanas-go/log"

type Matcher struct {
	loader Loader

	mappings Mappings
}

func NewMatcher(loader Loader) *Matcher {
	m, err := loader.GetMappings()
	if err != nil {
		log.Fatal(err)
	}
	return &Matcher{loader: loader, mappings: m}
}
