package app

import (
	"fmt"
	"strings"

	"github.com/ohler55/ojg/oj"
)

const (
	ScenarioMultipleStartingStateMessage = "the scenario has multiple starting states defined: [%s]"
	ScenarioNoStartingStateMessage       = "the scenario has no starting state defined"
	ScenarioInvalidStateNameMessage      = "the scenario has a state pointing to a new state that is not defined in the scenario: [%s -> %s]"
	ScenarioSingleStateMessage           = "the scenario must have at least 2 defined states"
)

type ScenarioState struct {
	CurrentState string
	States       map[string]Mapping
}

type ScenarioHandler struct {
	scenarios map[string]ScenarioState
}

type ScenarioValidationError struct {
	ScenarioName string `json:"scenario"`
	Message      string `json:"message"`
}

type ScenarioValidationErrors []ScenarioValidationError

func (v ScenarioValidationErrors) Error() string {
	return fmt.Sprintf("scenario definition is invalid: %s", oj.JSON(v))
}

func NewScenarioHandler() *ScenarioHandler {
	return &ScenarioHandler{
		scenarios: map[string]ScenarioState{},
	}
}

func (hand *ScenarioHandler) AddScenario(mapping Mapping) {

	if mapping.Scenario == nil {
		return
	}

	mapping = mapping.CalcMaxScoreAndCost()
	scMapping := mapping.Scenario

	sc, scenarioOk := hand.scenarios[scMapping.Name]
	if !scenarioOk {
		sc = ScenarioState{CurrentState: scMapping.State}
	}

	if len(sc.States) == 0 {
		sc.States = make(map[string]Mapping)
	}

	if scMapping.StartingState {
		sc.CurrentState = scMapping.State
	}

	sc.States[scMapping.State] = mapping

	hand.scenarios[scMapping.Name] = sc
}

// Validates the following:
//
//   - Each scenario has exactly one starting state
//   - Each scenario has at least 2 states
//   - State names are valid inside each scenario
func (hand *ScenarioHandler) ValidateScenarioStates() error {
	errors := make(ScenarioValidationErrors, 0)

	for k, v := range hand.scenarios {
		startingStates := []string{}
		for _, s := range v.States {
			if s.Scenario.StartingState {
				startingStates = append(startingStates, s.Scenario.State)
			}

			if s.Scenario.NewState != "" {
				if _, ok := v.States[s.Scenario.NewState]; !ok {
					errors = append(errors, ScenarioValidationError{k, fmt.Sprintf(ScenarioInvalidStateNameMessage, s.Scenario.State, s.Scenario.NewState)})
				}
			}
		}

		if len(startingStates) == 0 {
			errors = append(errors, ScenarioValidationError{k, ScenarioNoStartingStateMessage})
		}

		if len(startingStates) > 1 {
			errors = append(errors, ScenarioValidationError{k, fmt.Sprintf(ScenarioMultipleStartingStateMessage, strings.Join(startingStates, ", "))})
		}

		if len(v.States) <= 1 {
			errors = append(errors, ScenarioValidationError{k, ScenarioSingleStateMessage})
		}
	}

	if len(errors) > 0 {
		return errors
	}

	return nil //TODO: state validation
}
