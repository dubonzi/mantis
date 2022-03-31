package app

import (
	"net/http"

	"github.com/ohler55/ojg/oj"
)

const (
	NoMappingFoundMessage = "No mapping found for the request"
)

type Service struct {
	matcher Matcher
	delayer Delayer
}

type MatchResult struct {
	StatusCode int
	Headers    map[string]string
	Body       any
	Matched    bool
}

func NewMatchResult(mapping *Mapping, r Request, matched bool) MatchResult {
	result := MatchResult{
		Matched: matched,
	}

	if mapping == nil {
		result.StatusCode = http.StatusNotFound
		result.Body = buildNotFoundResponse(r, nil)
		return result
	}

	if !matched {
		result.Body = buildNotFoundResponse(r, &mapping.Request)
		result.StatusCode = http.StatusNotFound
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

func NewService(matcher Matcher, delayer Delayer) *Service {
	return &Service{matcher, delayer}
}

func (s *Service) MatchRequest(r Request) MatchResult {
	mapping, matched := s.matcher.Match(r)

	result := NewMatchResult(mapping, r, matched)

	if mapping != nil {
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
