package types

import (
	"encoding/json"
	"fmt"
)

type RolloutPhase string

const (
	PhaseRunning   RolloutPhase = "Running"
	PhaseSucceeded RolloutPhase = "Succeeded"
	PhaseFailed    RolloutPhase = "Failed"
	PhaseErrir     RolloutPhase = "Error"
)

type RpcRolloutResult struct {
	Phase               RolloutPhase
	Message             string
	RequeueAfterSeconds int
	Status              json.RawMessage
}

type RpcRolloutContext struct {
	PluginName string
	Config     json.RawMessage
	Status     json.RawMessage
}

func (p RolloutPhase) Validate() error {
	switch p {
	case PhaseRunning, PhaseSucceeded, PhaseFailed, PhaseErrir:
		return nil
	default:
		return fmt.Errorf("invalid phase: %s", p)
	}

}
