package app

import (
	"encoding/json"
	"net/http"
)

const (
	NoMappingFoundMessage = "No mapping found for the request"
)

type Service struct {
	matcher Matcher
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
	s, _ := json.Marshal(m)
	return string(s)
}

type NotFoundResponse struct {
	Message        string          `json:"message"`
	Request        Request         `json:"request"`
	ClosestMapping *RequestMapping `json:"closestMapping,omitempty"`
}

func NewService(matcher Matcher) *Service {
	return &Service{matcher}
}

func (s *Service) MatchRequest(r Request) MatchResult {
	mapping, matched := s.matcher.Match(r)

	result := NewMatchResult(mapping, r, matched)

	return result
}

func buildNotFoundResponse(r Request, mapping *RequestMapping) NotFoundResponse {
	return NotFoundResponse{
		Message:        NoMappingFoundMessage,
		Request:        r,
		ClosestMapping: mapping,
	}
}
