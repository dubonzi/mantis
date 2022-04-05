package app

import (
	"fmt"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestJSONPathCache(t *testing.T) {
	tests := []struct {
		expressions []string
		wantErr     bool
		wantLen     int
	}{
		{
			expressions: []string{`$[?(@.product.id == '12345')]`},
			wantErr:     false,
			wantLen:     1,
		},
		{
			expressions: []string{`$.person[?(@.age > 21 || @.name == 'John')]`, `$.name`},
			wantErr:     false,
			wantLen:     2,
		},
		{
			expressions: []string{`$[?(@.product.id == '12345')]`, `$[?(@.product.id == '12345')]`},
			wantErr:     false,
			wantLen:     1,
		},
		{
			expressions: []string{`$.person.[a]`},
			wantErr:     true,
			wantLen:     1,
		},
		{
			expressions: []string{`$.person[?(@.age > 21 && name == 'John')]`},
			wantErr:     true,
			wantLen:     1,
		},
	}

	for index, tt := range tests {
		t.Run(fmt.Sprint(index), func(t *testing.T) {

			jc := NewJSONPathCache()
			err := jc.AddExpressions(tt.expressions)
			if !assert.IsEqual(err != nil, tt.wantErr) {
				t.Logf("error parsing jsonpath expression: %s , %s", tt.expressions, err)
				t.FailNow()
			}

			if !tt.wantErr {
				assert.Equal(t, len(jc.cache), tt.wantLen)
			}
		})
	}

}
