package app

import (
	"fmt"

	"github.com/americanas-go/log"
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
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
			return fmt.Errorf("failed to parse jsonpath expression: %s : %w", expr, err)
		}

		j.cache[expr] = parsed
	}
	return nil
}

type jsonBody struct {
	value  any
	parsed bool
	valid  bool
}

func (b *jsonBody) parse(value string) bool {
	if !b.parsed {
		b.parsed = true
		var err error
		b.value, err = oj.ParseString(value)
		b.valid = err == nil
		if err != nil {
			log.Errorf("error parsing body json value for jsonpath matching: %s", err)
		}
	}
	return b.valid
}

func (j *JSONPathCache) Match(expressions []string, value string) bool {
	var body jsonBody
	return j.match(expressions, value, &body)
}

func (j *JSONPathCache) match(expressions []string, value string, body *jsonBody) bool {
	if len(expressions) == 0 {
		return true
	}
	if !body.parse(value) {
		return false
	}
	for _, sExpr := range expressions {
		expr := j.cache[sExpr]
		if len(expr.Get(body.value)) == 0 {
			return false
		}
	}
	return true
}
