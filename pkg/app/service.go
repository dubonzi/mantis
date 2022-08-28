package app

import (
	"net/http"

	"github.com/ohler55/ojg/oj"
)

const (
	NoMappingFoundMessage = "No mapping found for the request"
)

type Service struct {
	matcher  *Matcher
	delayer  Delayer
	mappings Mappings
}

type MatchResult struct {
	StatusCode int
	Headers    map[string]string
	Body       any
	Matched    bool
}

func NewMatchResult(mapping *Mapping, r Request, matched bool, partial bool) MatchResult {
	result := MatchResult{
		Matched: matched,
	}

	if partial {
		result.Body = buildNotFoundResponse(r, &mapping.Request)
		result.StatusCode = http.StatusNotFound
		return result
	}

	if !matched {
		result.StatusCode = http.StatusNotFound
		result.Body = buildNotFoundResponse(r, nil)
		return result
	}

	if mapping.Response.Body != "" {
		result.Body = mapping.Response.Body
	}
	result.StatusCode = mapping.Response.StatusCode
	result.Headers = mapping.Response.Headers

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

func NewService(mappings Mappings, matcher *Matcher, delayer Delayer) *Service {
	return &Service{matcher, delayer, mappings}
}

func (s *Service) MatchRequest(r Request) MatchResult {
	mapping, matched, partial := s.matcher.Match(r, s.mappings)

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
