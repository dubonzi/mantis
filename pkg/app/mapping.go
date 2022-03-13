package app

import (
	"encoding/json"
	"fmt"
)

type Mapping struct {
	Request  MappingRequest  `json:"request"`
	Response MappingResponse `json:"response"`
}

type PathMapping struct {
	Exact    string `json:"exact"`
	Contains string `json:"contains"`
	Pattern  string `json:"pattern"`
	JsonPath string `json:"jsonPath"`
}

func (p PathMapping) IsEmpty() bool {
	return p.Exact == "" && p.Contains == "" && p.Pattern == "" && p.JsonPath == ""
}

type BodyMapping struct {
	Exact    string `json:"exact"`
	Contains string `json:"contains"`
	Pattern  string `json:"pattern"`
	JsonPath string `json:"jsonPath"`
}

type HeaderMapping struct {
	Exact    string `json:"exact"`
	Contains string `json:"contains"`
	Pattern  string `json:"pattern"`
}

type MappingRequest struct {
	Method  string                   `json:"method"`
	Path    PathMapping              `json:"path"`
	Headers map[string]HeaderMapping `json:"headers"`
	Body    BodyMapping              `json:"body"`
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
	if m.Request.Path.IsEmpty() {
		errs = append(errs, ValidationError{"Request.Path", "Path mapping is required"})
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
