package app

import (
	"encoding/json"
	"fmt"
)

type Mapping struct {
	Request  MappingRequest  `json:"request"`
	Response MappingResponse `json:"response"`
}

type MappingRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}

type MappingResponse struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	BodyFile   string            `json:"bodyFile"`
	Body       string            `json:"body"`
}

type ValidationError struct {
	Field, Message string
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	b, _ := json.Marshal(v)
	return fmt.Sprintf("mapping definition is invalid: %s", string(b))
}

func (m Mapping) Validate() error {
	errs := make(ValidationErrors, 0)
	if m.Request.Method == "" {
		errs = append(errs, ValidationError{"Request.Method", "Method is required"})
	}
	if m.Request.URL == "" {
		errs = append(errs, ValidationError{"Request.URL", "URL is required"})
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

type Mappings map[string][]Mapping

func (m Mappings) Put(mapping Mapping) error {
	if err := mapping.Validate(); err != nil {
		return err
	}

	m[mapping.Request.Method] = append(m[mapping.Request.Method], mapping)
	return nil
}
