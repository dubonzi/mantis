package app

import (
	"encoding/json"
	"fmt"
)

type Mapping struct {
	Request  RequestMapping  `json:"request"`
	Response ResponseMapping `json:"response"`
}

func (m Mapping) MaxScore() int {
	score := 1 // Starts at 1 since Path is required

	if m.Request.HasBody() {
		score++
	}

	if m.Request.HasHeaders() {
		score++
	}

	return score
}

type PathMapping struct {
	Exact    string `json:"exact"`
	Contains string `json:"contains"`
	Pattern  string `json:"pattern"`
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

type RequestMapping struct {
	Method  string                   `json:"method"`
	Path    PathMapping              `json:"path"`
	Headers map[string]HeaderMapping `json:"headers"`
	Body    BodyMapping              `json:"body"`
}

func (m RequestMapping) HasPath() bool {
	return m.Path.Exact != "" || m.Path.Contains != "" || m.Path.Pattern != ""
}

func (m RequestMapping) HasHeaders() bool {
	return len(m.Headers) > 0
}

func (m RequestMapping) HasBody() bool {
	return m.Body.Exact != "" || m.Body.Contains != "" || m.Body.Pattern != "" || m.Body.JsonPath != ""
}

type ResponseMapping struct {
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
	if !m.Request.HasPath() {
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
