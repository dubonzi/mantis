package app

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			wantLen:     0,
		},
		{
			expressions: []string{`$.person[?(@.age > 21 && name == 'John')]`},
			wantErr:     true,
			wantLen:     0,
		},
	}

	for index, tt := range tests {
		t.Run(fmt.Sprint(index), func(t *testing.T) {

			jc := NewJSONPathCache()
			err := jc.AddExpressions(tt.expressions)
			require.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.wantLen, len(jc.cache))

		})
	}

}
