package agents

import (
	"fmt"
	"math"

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
// TODO: moves away from closest in social network
func (da *DismissiveAgent) GetTargetPosition() (infra.PositionVector, bool) {
	var closestInNetwork infra.IExtendedAgent = nil
	minDist := math.Inf(1)

	for otherID := range da.network {
		// Ignore self
		if otherID == da.GetID() {
			continue
		}

		otherAgent, alive := da.GetAgentByID(otherID)

		// ignore dead agents
		if !alive {
			continue
		}

		dist := da.position.Dist(otherAgent.GetPosition())
		if dist < minDist {
			minDist = dist
			closestInNetwork = otherAgent
		}
	}

	if closestInNetwork == nil {
		return infra.PositionVector{}, false
	}

	closestPos := closestInNetwork.GetPosition()
	selfPos := da.GetPosition()

	// closest->self + self == self - closest + self
	return selfPos.Sub(closestPos).Add(selfPos), true
}
