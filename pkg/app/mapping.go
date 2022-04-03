package app

import (
	"fmt"

	"github.com/americanas-go/log"
	"github.com/ohler55/ojg/oj"
)

type Mapping struct {
	Request  RequestMapping  `json:"request"`
	Response ResponseMapping `json:"response"`
}

func (m Mapping) MaxScore() int {
	return m.Request.PathScore() + m.Request.HeaderScore() + m.Request.BodyScore()
}

type CommonMatch struct {
	Exact    string   `json:"exact,omitempty"`
	Contains []string `json:"contains,omitempty"`
	Patterns []string `json:"pattern,omitempty"`
}

type BodyMatch struct {
	CommonMatch
	JsonPath []string `json:"jsonPath,omitempty"`
}

type RequestMapping struct {
	Method  string                 `json:"method"`
	Path    CommonMatch            `json:"path"`
	Headers map[string]CommonMatch `json:"headers,omitempty"`
	Body    BodyMatch              `json:"body,omitempty"`
}

func (m RequestMapping) HasPath() bool {
	return m.Path.Exact != "" || len(m.Path.Contains) > 0 || len(m.Path.Patterns) > 0
}

func (m RequestMapping) HeaderScore() int {
	var score int
	for _, h := range m.Headers {
		if h.Exact != "" {
			score++
			continue
		}

		score += len(h.Contains) + len(h.Patterns)
	}
	return score
}

func (m RequestMapping) PathScore() int {
	if m.Path.Exact != "" {
		return 1
	}
	return len(m.Path.Contains) + len(m.Path.Patterns)
}

func (m RequestMapping) BodyScore() int {
	if m.Body.Exact != "" {
		return 1
	}
	return len(m.Body.JsonPath) + len(m.Body.Contains) + len(m.Body.Patterns)
}

type ResponseMapping struct {
	StatusCode    int               `json:"statusCode"`
	Headers       map[string]string `json:"headers,omitempty"`
	BodyFile      string            `json:"bodyFile,omitempty"`
	Body          string            `json:"body,omitempty"`
	ResponseDelay Delay             `json:"delay,omitempty"`
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

type Mappings map[string][]*Mapping

func (m Mappings) Put(mapping Mapping) error {
	if err := mapping.Validate(); err != nil {
		return err
	}

	log.Tracef("adding mapping: %+v", mapping)
	m[mapping.Request.Method] = append(m[mapping.Request.Method], &mapping)
	return nil
}
