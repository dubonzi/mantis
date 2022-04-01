package app

import (
	"testing"

	"github.com/go-playground/assert/v2"
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
					Path: PathMapping{Pattern: `/[A-z0-9]+/`},
				},
			},
			wantLen: 1,
			wantErr: false,
		},
		{
			mapping: Mapping{
				Request: RequestMapping{
					Path:    PathMapping{Pattern: `[A-z0-9]+`},
					Headers: map[string]HeaderMapping{"accept": {Pattern: ".*"}, "x-id": {Pattern: `\d*`}, "x-debug": {Pattern: ".*"}},
					Body:    BodyMapping{Pattern: `\d{3}\.\d{3}\.\d{3}-\d{2}`},
				},
			},
			wantLen: 4,
			wantErr: false,
		},
		{
			mapping: Mapping{
				Request: RequestMapping{
					Path: PathMapping{Pattern: `([A-z0-9]+`},
				},
			},
			wantLen: 0,
			wantErr: true,
		},
		{
			mapping: Mapping{
				Request: RequestMapping{
					Headers: map[string]HeaderMapping{"accept": {Pattern: "((.*json)"}},
				},
			},
			wantLen: 0,
			wantErr: true,
		},
		{
			mapping: Mapping{
				Request: RequestMapping{
					Body: BodyMapping{Pattern: `\d{)}*`},
				},
			},
			wantLen: 0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		rc := NewRegexCache()
		err := rc.AddFromMapping(tt.mapping)
		if !assert.IsEqual(err != nil, tt.wantErr) {
			t.Log("error parsing regex pattern: ", err)
			t.Fail()
		}

		assert.Equal(t, len(rc.cache), tt.wantLen)

		if !tt.wantErr {

			if tt.mapping.Request.Path.Pattern != "" {
				_, ok := rc.cache[tt.mapping.Request.Path.Pattern]
				assert.Equal(t, true, ok)
			}
			if tt.mapping.Request.Body.Pattern != "" {
				_, ok := rc.cache[tt.mapping.Request.Body.Pattern]
				assert.Equal(t, true, ok)
			}
			for _, v := range tt.mapping.Request.Headers {
				if v.Pattern != "" {
					_, ok := rc.cache[v.Pattern]
					assert.Equal(t, true, ok)
				}
			}
		}
	}

}
