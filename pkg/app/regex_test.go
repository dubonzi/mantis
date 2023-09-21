package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegexCache(t *testing.T) {
	tests := []struct {
		mapping Mapping
		wantLen int
		wantErr bool
	}{
		{
			mapping: Mapping{
				Request: RequestMapping{
					Path: CommonMatch{Patterns: []string{`/[A-z0-9]+/`}},
				},
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			mapping: Mapping{
				Request: RequestMapping{
					Path:    CommonMatch{Patterns: []string{`[A-z0-9]+`}},
					Headers: map[string]CommonMatch{"accept": {Patterns: []string{".*"}}, "x-id": {Patterns: []string{`\d*`}}, "x-debug": {Patterns: []string{".*"}}},
					Body:    BodyMatch{CommonMatch: CommonMatch{Patterns: []string{`\d{3}\.\d{3}\.\d{3}-\d{2}`}}},
				},
			},
			wantLen: 4,
			wantErr: false,
		},
		{
			mapping: Mapping{
				Request: RequestMapping{
					Path: CommonMatch{Patterns: []string{`([A-z0-9]+`}},
				},
			},
			wantLen: 0,
			wantErr: true,
		},
		{
			mapping: Mapping{
				Request: RequestMapping{
					Headers: map[string]CommonMatch{"accept": {Patterns: []string{"((.*json)"}}},
				},
			},
			wantLen: 0,
			wantErr: true,
		},
		{
			mapping: Mapping{
				Request: RequestMapping{
					Body: BodyMatch{CommonMatch: CommonMatch{Patterns: []string{`\d{)}*`}}},
				},
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		rc := NewRegexCache()
		err := rc.AddFromMapping(tt.mapping)

		require.Equal(t, tt.wantErr, err != nil)
		assert.Equal(t, len(rc.cache), tt.wantLen)

		if !tt.wantErr {

			for _, p := range tt.mapping.Request.Path.Patterns {
				_, ok := rc.cache[p]
				assert.Equal(t, true, ok)
			}
			for _, p := range tt.mapping.Request.Body.Patterns {
				_, ok := rc.cache[p]
				assert.Equal(t, true, ok)
			}
			for _, v := range tt.mapping.Request.Headers {
				for _, p := range v.Patterns {
					_, ok := rc.cache[p]
					assert.Equal(t, true, ok)
				}
			}
		}
	}

}
