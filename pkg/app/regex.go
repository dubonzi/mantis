package app

import (
	"regexp"

	"github.com/pkg/errors"
)

type RegexCache struct {
	cache map[string]*regexp.Regexp
}

func NewRegexCache() *RegexCache {
	return &RegexCache{
		cache: make(map[string]*regexp.Regexp),
	}
}

func (r *RegexCache) AddFromMapping(mapping Mapping) error {
	var err error
	for _, p := range mapping.Request.Path.Patterns {
		err = r.compileAndPut(p)
		if err != nil {
			return errors.Wrapf(err, "failed to compile path regex with pattern:  %s ", p)
		}
	}

	for _, p := range mapping.Request.Body.Patterns {
		err = r.compileAndPut(p)
		if err != nil {
			return errors.Wrapf(err, "failed to compile body regex with pattern:  %s ", p)
		}
	}

	for _, value := range mapping.Request.Headers {
		for _, p := range value.Patterns {
			err = r.compileAndPut(p)
			if err != nil {
				return errors.Wrapf(err, "failed to compile header regex with pattern: %s ", value.Patterns)
			}
		}
	}

	return nil
}

func (r *RegexCache) compileAndPut(pattern string) error {
	if _, ok := r.cache[pattern]; ok {
		return nil
	}

	rgxp, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}

	r.cache[pattern] = rgxp
	return nil
}

func (r *RegexCache) Match(pattern, value string) bool {
	return r.cache[pattern].Match([]byte(value))
}
