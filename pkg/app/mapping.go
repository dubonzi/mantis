package app

import (
	"encoding/json"
	"fmt"

	"github.com/americanas-go/log"
)

const (
	StartingScore = 1
)

type Mapping struct {
	Request  RequestMapping  `json:"request"`
	Response ResponseMapping `json:"response"`
}

func (m Mapping) MaxScore() int {
	score := StartingScore // Starts at 1 since Path is required

	if m.Request.HasBody() {
		score++
	}

	if m.Request.HasHeaders() {
		score++
	}

	return score
}

type PathMapping struct {
	Exact    string `json:"exact,omitempty"`
	Contains string `json:"contains,omitempty"`
	Pattern  string `json:"pattern,omitempty"`
}

type BodyMapping struct {
	Exact    string `json:"exact,omitempty"`
	Contains string `json:"contains,omitempty"`
	Pattern  string `json:"pattern,omitempty"`
	JsonPath string `json:"jsonPath,omitempty"`
}

type HeaderMapping struct {
	Exact    string `json:"exact,omitempty"`
	Contains string `json:"contains,omitempty"`
	Pattern  string `json:"pattern,omitempty"`
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
	StatusCode    int               `json:"statusCode"`
	Headers       map[string]string `json:"headers"`
	BodyFile      string            `json:"bodyFile"`
	Body          string            `json:"body"`
	ResponseDelay Delay             `json:"delay"`
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

	log.Tracef("adding mapping: %+v", mapping)
	m[mapping.Request.Method] = append(m[mapping.Request.Method], mapping)
	return nil
}
