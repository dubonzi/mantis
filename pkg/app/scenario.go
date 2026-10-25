package app

import (
	"fmt"
	"sync"
)

const (
	ScenarioMultipleStartingStateMessage     = "the scenario has multiple starting states defined"
	ScenarioNoStartingStateMessage           = "the scenario has no starting state defined"
	ScenarioInvalidStateNameMessage          = "the scenario has a state pointing to a new state that is not defined in the scenario: [%s -> %s]"
	ScenarioInconsistentStartingStateMessage = "the scenario has mappings with inconsistent startingState values for state: [%s]"
	ScenarioSingleStateMessage               = "the scenario must have at least 2 defined states"
)

type ScenarioState struct {
	CurrentState string
	States       map[string][]Mapping
}

type ScenarioHandler struct {
	mu               sync.Mutex
	matcher          *Matcher
	scenarioMappings Mappings
	scenarios        map[string]ScenarioState
}

type ScenarioValidationError struct {
	ScenarioName string `json:"scenario"`
	Message      string `json:"message"`
}

type ScenarioValidationErrors []ScenarioValidationError

func (v ScenarioValidationErrors) Error() string {
	return fmt.Sprintf("scenario definition is invalid: %s", jsonString(v))
}

func NewScenarioHandler(matcher *Matcher) *ScenarioHandler {
	return &ScenarioHandler{
		scenarios:        map[string]ScenarioState{},
		scenarioMappings: make(Mappings),
		matcher:          matcher,
	}
}

func (hand *ScenarioHandler) AddScenario(mapping Mapping) error {
	if mapping.Scenario == nil {
		return nil
	}

	hand.mu.Lock()
	defer hand.mu.Unlock()

	if err := hand.scenarioMappings.Put(mapping); err != nil {
		return err
	}

	scMapping := mapping.Scenario
	sc, scenarioOk := hand.scenarios[scMapping.Name]
	if !scenarioOk {
		sc = ScenarioState{CurrentState: scMapping.State, States: make(map[string][]Mapping)}
	}
	if scMapping.StartingState {
		sc.CurrentState = scMapping.State
	}
	sc.States[scMapping.State] = append(sc.States[scMapping.State], mapping)
	hand.scenarios[scMapping.Name] = sc
	return nil
}

func (hand *ScenarioHandler) MatchScenario(request Request) MatchOutcome {
	var body jsonBody
	return hand.matchScenario(request, &body)
}

func (hand *ScenarioHandler) matchScenario(request Request, body *jsonBody) MatchOutcome {
	hand.mu.Lock()
	defer hand.mu.Unlock()

	match := hand.matcher.match(request, hand.scenarioMappings, hand.scenarios, body)
	if match.Kind != FullMatch || match.Mapping.Scenario == nil {
		return MatchOutcome{}
	}

	scenario := match.Mapping.Scenario
	if scenario.NewState != "" {
		state := hand.scenarios[scenario.Name]
		state.CurrentState = scenario.NewState
		hand.scenarios[scenario.Name] = state
	}
	return match
}

// Validates the following:
//
//   - Each scenario has exactly one distinct starting state
//   - Mappings in the same state agree on startingState
//   - Each scenario has at least 2 states
//   - State names are valid inside each scenario
func (hand *ScenarioHandler) ValidateScenarioStates() error {
	hand.mu.Lock()
	defer hand.mu.Unlock()
	errors := make(ScenarioValidationErrors, 0)

	for k, v := range hand.scenarios {
		startingStates := 0
		for stateName, mappings := range v.States {
			starting, inconsistent := false, false
			for _, mapping := range mappings {
				s := mapping.Scenario
				starting = starting || s.StartingState
				if s.StartingState != mappings[0].Scenario.StartingState {
					inconsistent = true
				}
				if s.NewState != "" {
					if _, ok := v.States[s.NewState]; !ok {
						errors = append(errors, ScenarioValidationError{k, fmt.Sprintf(ScenarioInvalidStateNameMessage, s.State, s.NewState)})
					}
				}
			}
			if starting {
				startingStates++
			}
			if inconsistent {
				errors = append(errors, ScenarioValidationError{k, fmt.Sprintf(ScenarioInconsistentStartingStateMessage, stateName)})
			}
		}

		if startingStates == 0 {
			errors = append(errors, ScenarioValidationError{k, ScenarioNoStartingStateMessage})
		}
		if startingStates > 1 {
			errors = append(errors, ScenarioValidationError{k, ScenarioMultipleStartingStateMessage})
		}

		if len(v.States) <= 1 {
			errors = append(errors, ScenarioValidationError{k, ScenarioSingleStateMessage})
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil
}
