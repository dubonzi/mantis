package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

func encodeJSON(value any) (string, error) {
	var body bytes.Buffer
	encoder := json.NewEncoder(&body)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		return "", err
	}
	return strings.TrimSuffix(body.String(), "\n"), nil
}

func jsonString(value any) string {
	body, err := encodeJSON(value)
	if err != nil {
		return fmt.Sprintf("error encoding JSON: %v", err)
	}
	return body
}
