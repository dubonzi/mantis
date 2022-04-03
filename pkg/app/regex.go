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
	for _, p := range mapping.Request.Path.Pattern {
		err = r.compileAndPut(p)
		if err != nil {
			return errors.Wrapf(err, "failed to compile path regex with pattern:  %s ", mapping.Request.Path.Pattern)
		}
	}

	if mapping.Request.Body.Pattern != "" {
		err = r.compileAndPut(mapping.Request.Body.Pattern)
		if err != nil {
			return errors.Wrapf(err, "failed to compile body regex with pattern:  %s ", mapping.Request.Body.Pattern)
		}
	}

	for _, value := range mapping.Request.Headers {
		if value.Pattern != "" {
			err = r.compileAndPut(value.Pattern)
			if err != nil {
				return errors.Wrapf(err, "failed to compile header regex with pattern: %s ", value.Pattern)
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
