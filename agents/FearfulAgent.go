package agents

import (
	"fmt"

	"github.com/MattSScott/TMT_SOMAS/infra"
)

type FearfulAgent struct {
	*ExtendedAgent
}

func CreateFearfulAgent(server infra.IServer) *FearfulAgent {
	worldview := infra.NewWorldview(byte(0b00))
	extendedAgent := CreateExtendedAgent(server, worldview)

	// Set Fearful-style attachment: high anxiety, high avoidance
	extendedAgent.attachment = infra.Attachment{
		Anxiety:   randInRange(0.5, 1.0),
		Avoidance: randInRange(0.5, 1.0),
		Type:      infra.FEARFUL,
	}
	// these ranges to be tweaked
	extendedAgent.PTW = infra.PTSParams{
		CheckProb: randInRange(0.5, 1.0),
		ReplyProb: randInRange(0.0, 0.5),
		Alpha:     randInRange(0.0, 0.5),
		Beta:      randInRange(0.5, 1.0),
	}

	return &FearfulAgent{
		ExtendedAgent: extendedAgent,
	}
}
func (fa *FearfulAgent) AgentInitialised() {
	atch := fa.GetAttachment()
	fmt.Printf("Fearful Agent %v added with with Age: %d, Attachment: [%.2f, %.2f]\n", fa.GetID(), fa.GetAge(), atch.Anxiety, atch.Avoidance)
}

// Fearful agent movement policy
// Objective: moves away from mean cluster position
func (fa *FearfulAgent) GetTargetPosition() (infra.PositionVector, bool) {
	meanPos := fa.GetClusterWeightedMeanPosition()
	targetPos := fa.NormalizeToUnit(fa.GetPosition().Sub(meanPos))
	return targetPos, true
}
