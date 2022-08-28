package app

import (
	"fmt"

	"github.com/americanas-go/log"
	"github.com/ohler55/ojg/oj"
)

const (
	ContainsCost = 2
	JsonPathCost = 4
	RegexCost    = 5
)

type Mapping struct {
	Scenario *ScenarioMapping `json:"scenario"`
	Request  RequestMapping   `json:"request"`
	Response ResponseMapping  `json:"response"`

	MaxScore int
	Cost     int
}

func (m Mapping) CalcMaxScoreAndCost() Mapping {
	m.MaxScore = m.Request.PathScore() + m.Request.HeaderScore() + m.Request.BodyScore()

	var cost int

	cost += m.Request.Path.Cost() + m.Request.Body.Cost()

	for _, v := range m.Request.Headers {
		cost += v.Cost()
	}

	m.Cost = cost

	return m
}

func (m Mapping) Validate() error {
	errs := make(ValidationErrors, 0)
	if m.Request.Method == "" {
		errs = append(errs, ValidationError{"Request.Method", "Method is required"})
	}

	if !m.Request.HasPath() {
		errs = append(errs, ValidationError{"Request.Path", "Path mapping is required"})
	}

	if m.Scenario != nil {
		errs = append(errs, m.Scenario.Validate()...)
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

	mapping = mapping.CalcMaxScoreAndCost()

	m[mapping.Request.Method] = append(m[mapping.Request.Method], mapping)
	return nil
}

func (m Mappings) PutAll(mappings []Mapping) error {
	for _, mapping := range mappings {
		err := m.Put(mapping)
		if err != nil {
			return err
		}
	}
	return nil
}

type CommonMatch struct {
	Exact    string   `json:"exact,omitempty"`
	Contains []string `json:"contains,omitempty"`
	Patterns []string `json:"pattern,omitempty"`
}

func (c CommonMatch) Cost() int {
	return (len(c.Contains) * ContainsCost) + (len(c.Patterns) * RegexCost)
}

type BodyMatch struct {
	CommonMatch
	JsonPath []string `json:"jsonPath,omitempty"`
}

func (b BodyMatch) Cost() int {
	return (len(b.Contains) * ContainsCost) + (len(b.Patterns) * RegexCost) + (len(b.JsonPath) * JsonPathCost)
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

type ScenarioMapping struct {
	Name          string `json:"name"`
	StartingState bool   `json:"startingState"`
	State         string `json:"state"`
	NewState      string `json:"newState"`
}

func (s *ScenarioMapping) Validate() ValidationErrors {
	errs := make(ValidationErrors, 0)
	if s.Name == "" {
		errs = append(errs, ValidationError{"Scenario.Name", "Scenario name is required"})
	}
	if s.State == "" {
		errs = append(errs, ValidationError{"Scenario.State", "Scenario state is required"})
	}

	return errs
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
