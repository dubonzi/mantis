package app

type Mappings map[string][]Mapping

type Mapping struct {
	Request struct {
		URL     string            `json:"url"`
		Headers map[string]string `json:"headers"`
	}

	Response struct {
		StatusCode int               `json:"statusCode"`
		Headers    map[string]string `json:"headers"`
		BodyFile   string            `json:"bodyFile"`
		Body       string            `json:"body"`
	}
}
