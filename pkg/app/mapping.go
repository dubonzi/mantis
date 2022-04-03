package app

import (
	"fmt"

	"github.com/americanas-go/log"
	"github.com/ohler55/ojg/oj"
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

	score += m.Request.HeaderScore()
	score += m.Request.BodyScore()

	return score
}

type PathMapping struct {
	Exact    string   `json:"exact,omitempty"`
	Contains []string `json:"contains,omitempty"`
	Pattern  []string `json:"pattern,omitempty"`
}

type BodyMapping struct {
	Exact    string   `json:"exact,omitempty"`
	Contains string   `json:"contains,omitempty"`
	Pattern  string   `json:"pattern,omitempty"`
	JsonPath []string `json:"jsonPath,omitempty"`
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
	return m.Path.Exact != "" || len(m.Path.Contains) > 0 || len(m.Path.Pattern) > 0
}

func (m RequestMapping) HeaderScore() int {
	return len(m.Headers)
}

func (m RequestMapping) BodyScore() int {
	if m.Body.Exact != "" || m.Body.Contains != "" || m.Body.Pattern != "" {
		return 1
	}
	return len(m.Body.JsonPath)
}

type ResponseMapping struct {
	StatusCode    int               `json:"statusCode"`
	Headers       map[string]string `json:"headers"`
	BodyFile      string            `json:"bodyFile"`
	Body          string            `json:"body"`
	ResponseDelay Delay             `json:"delay"`
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	return fmt.Sprintf("mapping definition is invalid: %s", oj.JSON(v))
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
