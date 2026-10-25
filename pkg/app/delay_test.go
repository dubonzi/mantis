package app

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDurationUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Duration
		wantErr bool
	}{
		{name: "Should decode milliseconds", input: `"250ms"`, want: Duration(250 * time.Millisecond)},
		{name: "Should decode compound durations", input: `"1m30s"`, want: Duration(90 * time.Second)},
		{name: "Should reject invalid durations", input: `"soon"`, wantErr: true},
		{name: "Should reject numbers", input: `250`, wantErr: true},
		{name: "Should reject booleans", input: `true`, wantErr: true},
		{name: "Should reject null", input: `null`, wantErr: true},
		{name: "Should reject objects", input: `{}`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var duration Duration
			err := json.Unmarshal([]byte(tt.input), &duration)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, duration)
		})
	}
}
