package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONStringOutput(t *testing.T) {
	tests := []struct {
		name  string
		value interface{ String() string }
		want  string
	}{
		{
			name:  "Should use match result tags without HTML escaping",
			value: MatchResult{StatusCode: 200, Headers: map[string]string{"X-Test": "<tag>&"}, Body: "<tag>&", Matched: true},
			want:  `{"statusCode":200,"headers":{"X-Test":"<tag>&"},"body":"<tag>&","matched":true}`,
		},
		{
			name:  "Should omit absent match result headers and body",
			value: MatchResult{StatusCode: 404},
			want:  `{"statusCode":404,"matched":false}`,
		},
		{
			name:  "Should describe unsupported match result bodies",
			value: MatchResult{Body: make(chan int)},
			want:  "error encoding JSON: json: unsupported type: chan int",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.value.String())
		})
	}
}

func TestValidationErrorJSON(t *testing.T) {
	tests := []struct {
		name  string
		value error
		want  string
	}{
		{
			name:  "Should use mapping validation tags without HTML escaping",
			value: ValidationErrors{{Field: "Request.Method", Message: "<method>& is required"}},
			want:  `mapping definition is invalid: [{"field":"Request.Method","message":"<method>& is required"}]`,
		},
		{
			name:  "Should use scenario validation tags without HTML escaping",
			value: ScenarioValidationErrors{{ScenarioName: "<scenario>&", Message: "state a -> b is invalid"}},
			want:  `scenario definition is invalid: [{"scenario":"<scenario>&","message":"state a -> b is invalid"}]`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.value.Error())
		})
	}
}
