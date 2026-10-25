package app

import (
	"maps"
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
	StatusCode int
	Headers    map[string]string
	Body       any
	Matched    bool
}

func NewMatchResult(match MatchOutcome, r Request) MatchResult {
	mapping := match.Mapping
	result := MatchResult{
		Matched: match.Kind == FullMatch,
		Headers: make(map[string]string),
	}

	if match.Kind == PartialMatch {
		result.Body = buildNotFoundResponse(r, &mapping.Request)
		result.StatusCode = http.StatusNotFound
		result.Headers["Content-type"] = "application/json"
		result.Headers["X-Mapping-File"] = mapping.FilePath
		return result
	}

	if match.Kind != FullMatch {
		result.StatusCode = http.StatusNotFound
		result.Body = buildNotFoundResponse(r, nil)
		result.Headers["Content-type"] = "application/json"
		return result
	}

	if mapping.Response.Body != "" {
		result.Body = mapping.Response.Body
	}
	result.StatusCode = mapping.Response.StatusCode
	result.Headers = maps.Clone(mapping.Response.Headers)
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
	var body jsonBody
	match := s.scenarioHandler.matchScenario(r, &body)
	if match.Kind != FullMatch {
		match = s.matcher.match(r, s.mappings, nil, &body)
	}

	result := NewMatchResult(match, r)

	if match.Kind == FullMatch {
		s.delayer.Apply(&match.Mapping.Response.ResponseDelay)
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
