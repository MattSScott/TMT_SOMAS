package agents

import (
	"fmt"

	"github.com/MattSScott/TMT_SOMAS/infra"
)

type DismissiveAgent struct {
	*ExtendedAgent
}

func CreateDismissiveAgent(server infra.IServer) *DismissiveAgent {
	worldview := infra.NewWorldview(byte(0b01))

	extendedAgent := CreateExtendedAgent(server, worldview)

	// Set Dismissive-style attachment: low anxiety, high avoidance
	extendedAgent.attachment = infra.Attachment{
		Anxiety:   randInRange(0.0, 0.5),
		Avoidance: randInRange(0.5, 1.0),
		Type:      infra.DISMISSIVE,
	}
	// these ranges to be tweaked
	extendedAgent.PTW = infra.PTSParams{
		CheckProb: randInRange(0.0, 0.5),
		ReplyProb: randInRange(0.0, 0.5),
		Alpha:     randInRange(0.0, 0.5),
		Beta:      randInRange(0.0, 0.5),
	}

	return &DismissiveAgent{
		ExtendedAgent: extendedAgent,
	}
}

func (da *DismissiveAgent) AgentInitialised() {
	atch := da.GetAttachment()
	fmt.Printf("Dismissive Agent %v added with with Age: %d, Attachment: [%.2f, %.2f]\n", da.GetID(), da.GetAge(), atch.Anxiety, atch.Avoidance)
}

// dismissive agent movement policy
// Objective: moves away from mean position of social network
func (da *DismissiveAgent) GetTargetPosition() (infra.PositionVector, bool) {
	meanPos := da.GetNetworkWeightedMeanPosition()
	targetPos := da.NormalizeToUnit(da.GetPosition().Sub(meanPos))
	return targetPos, true
}
