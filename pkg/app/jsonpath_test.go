package app

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestJSONPathCache(t *testing.T) {
	tests := []struct {
		expression string
		wantErr    bool
	}{
		{
			expression: `$[?(@.product.id == '12345')]`,
			wantErr:    false,
		},
		{
			expression: `$.person[?(@.age > 21 || @.name == 'John')]`,
			wantErr:    false,
		},
		{
			expression: `$.person.[a]`,
			wantErr:    true,
		},
		{
			expression: `$.person[?(@.age > 21 && name == 'John')]`,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		jc := NewJSONPathCache()
		err := jc.AddExpressions([]string{tt.expression})
		if !assert.IsEqual(err != nil, tt.wantErr) {
			t.Logf("error parsing jsonpath expression: %s , %s", tt.expression, err)
			t.Fail()
		}

		if !tt.wantErr {
			assert.Equal(t, len(jc.cache), 1)
		}
	}

}
