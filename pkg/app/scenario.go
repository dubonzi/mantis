package app

type ScenarioState struct {
	CurrentState string
	States       map[string]Mapping
}

type ScenarioHandler struct {
	scenarios map[string]ScenarioState
}

func NewScenarioHandler() *ScenarioHandler {
	return &ScenarioHandler{
		scenarios: map[string]ScenarioState{},
	}
}

func (handler *ScenarioHandler) AddScenario(mapping Mapping) {

	if mapping.Scenario == nil {
		return
	}

	scMapping := mapping.Scenario

	if sc, scenarioOK := handler.scenarios[scMapping.Name]; scenarioOK {
		if len(sc.States) == 0 {
			sc.States = make(map[string]Mapping)
		}

		if scMapping.StartingState {
			sc.CurrentState = scMapping.Name
		}

		sc.States[scMapping.State] = mapping
	}
}

// Validates the following:
//
//   - Each scenario has exactly one starting state
//   - Each scenario has at least 2 states
//   - State names are valid inside each scenario
func (handler *ScenarioHandler) ValidateScenarioStates() error {
	return nil //TODO: state validation
}
