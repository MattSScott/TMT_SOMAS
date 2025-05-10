package agents

import (
	"fmt"
	"math"

	"github.com/MattSScott/TMT_SOMAS/infra"
)

type SecureAgent struct {
	*ExtendedAgent
}

func CreateSecureAgent(server infra.IServer) *SecureAgent {
	worldview := infra.NewWorldview(byte(0b11))
	extendedAgent := CreateExtendedAgent(server, worldview)

	// Set Secure-style attachment: low anxiety, low avoidance
	extendedAgent.attachment = infra.Attachment{
		Anxiety:   randInRange(0.0, 0.5),
		Avoidance: randInRange(0.0, 0.5),
		Type:      infra.SECURE,
	}
	// these ranges to be tweaked
	extendedAgent.PTW = infra.PTSParams{
		CheckProb: randInRange(0.0, 0.5),
		ReplyProb: randInRange(0.5, 1.0),
		Alpha:     randInRange(0.5, 1.0),
		Beta:      randInRange(0.0, 0.5),
	}

	return &SecureAgent{
		ExtendedAgent: extendedAgent,
	}
}

func (sa *SecureAgent) AgentInitialised() {
	atch := sa.GetAttachment()
	fmt.Printf("Secure Agent %v added with with Age: %d, Attachment: [%.2f, %.2f]\n", sa.GetID(), sa.GetAge(), atch.Anxiety, atch.Avoidance)
}

// Secure agent movement policy
// TODO: moves towards closest in social network
func (da *SecureAgent) GetTargetPosition() (infra.PositionVector, bool) {
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

	// closest->self + self == self - closest + self
	return closestInNetwork.GetPosition(), true
}
