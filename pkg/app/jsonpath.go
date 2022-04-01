package app

import (
	"github.com/americanas-go/log"
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
	"github.com/pkg/errors"
)

type JSONPathCache struct {
	cache map[string]jp.Expr
}

func NewJSONPathCache() *JSONPathCache {
	return &JSONPathCache{
		cache: make(map[string]jp.Expr),
	}
}

func (j *JSONPathCache) AddExpressions(expressions []string) error {
	for _, expr := range expressions {
		if _, ok := j.cache[expr]; ok {
			continue
		}

		parsed, err := jp.ParseString(expr)
		if err != nil {
			return errors.Wrapf(err, "failed to parse jsonpath expression: %s ", expr)
		}

		j.cache[expr] = parsed
	}
	return nil
}

func (j *JSONPathCache) Match(expressions []string, value string) bool {
	for _, sExpr := range expressions {
		expr := j.cache[sExpr]
		parsedValue, err := oj.ParseString(value)
		if err != nil {
			log.Errorf("error parsing body json value for jsonpath matching: %s", err)
			return false
		}
		if len(expr.Get(parsedValue)) == 0 {
			return false
		}
	}
	return true
}
