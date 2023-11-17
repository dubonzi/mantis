package app

import (
	"net/http"

	"github.com/ohler55/ojg/oj"
)

const (
	NoMappingFoundMessage = "No mapping found for the request"
)

type Service struct {
	matcher         *Matcher
	scenarioHandler *ScenarioHandler
	delayer         Delayer
	mappings        Mappings
}

type MatchResult struct {
	StatusCode  int
	Headers     map[string]string
	Body        any
	Matched     bool
	MappingFile string
}

func NewMatchResult(mapping *Mapping, r Request, matched bool, partial bool) MatchResult {
	result := MatchResult{
		Matched: matched,
		Headers: make(map[string]string),
	}

	if partial {
		result.Body = buildNotFoundResponse(r, &mapping.Request)
		result.StatusCode = http.StatusNotFound
		result.Headers["Content-type"] = "application/json"
		result.Headers["X-Mapping-File"] = mapping.FilePath
		return result
	}

	if !matched {
		result.StatusCode = http.StatusNotFound
		result.Body = buildNotFoundResponse(r, nil)
		result.Headers["Content-type"] = "application/json"
		return result
	}

	if mapping.Response.Body != "" {
		result.Body = mapping.Response.Body
	}
	result.StatusCode = mapping.Response.StatusCode
	result.Headers = mapping.Response.Headers
	if result.Headers == nil {
		result.Headers = make(map[string]string)
	}
	result.Headers["X-Mapping-File"] = mapping.FilePath

	return result
}

func (m MatchResult) String() string {
	return oj.JSON(m)
}

type NotFoundResponse struct {
	Message        string          `json:"message"`
	Request        Request         `json:"request"`
	ClosestMapping *RequestMapping `json:"closestMapping,omitempty"`
}

func NewService(mappings Mappings, matcher *Matcher, scenarioHandler *ScenarioHandler, delayer Delayer) *Service {
	return &Service{
		matcher:         matcher,
		scenarioHandler: scenarioHandler,
		delayer:         delayer,
		mappings:        mappings,
	}
}

func (s *Service) MatchRequest(r Request) MatchResult {
	var mapping Mapping
	var matched, partial bool

	mapping, matched, partial = s.scenarioHandler.MatchScenario(r)
	if !matched {
		mapping, matched, partial = s.matcher.Match(r, s.mappings, nil)
	}

	result := NewMatchResult(&mapping, r, matched, partial)

	if matched {
		s.delayer.Apply(&mapping.Response.ResponseDelay)
	}

	return result
}

func buildNotFoundResponse(r Request, mapping *RequestMapping) NotFoundResponse {
	return NotFoundResponse{
		Message:        NoMappingFoundMessage,
		Request:        r,
		ClosestMapping: mapping,
	}
}
