package agents

import (
	"fmt"
	"math"

	"github.com/MattSScott/TMT_SOMAS/infra"
)

type PreoccupiedAgent struct {
	*ExtendedAgent
}

func CreatePreoccupiedAgent(server infra.IServer) *PreoccupiedAgent {
	worldview := infra.NewWorldview(byte(0b10))
	extendedAgent := CreateExtendedAgent(server, worldview)

	// Set Preoccupied-style attachment: high anxiety, low avoidance
	extendedAgent.attachment = infra.Attachment{
		Anxiety:   randInRange(0.5, 1.0),
		Avoidance: randInRange(0.0, 0.5),
		Type:      infra.PREOCCUPIED,
	}
	// these ranges to be tweaked
	extendedAgent.PTW = infra.PTSParams{
		CheckProb: randInRange(0.5, 1.0),
		ReplyProb: randInRange(0.5, 1.0),
		Alpha:     randInRange(0.5, 1.0),
		Beta:      randInRange(0.5, 1.0),
	}

	return &PreoccupiedAgent{
		ExtendedAgent: extendedAgent,
	}
}

func (pa *PreoccupiedAgent) AgentInitialised() {
	atch := pa.GetAttachment()
	fmt.Printf("Preoccupied Agent %v added with with Age: %d, Attachment: [%.2f, %.2f]\n", pa.GetID(), pa.GetAge(), atch.Anxiety, atch.Avoidance)
}

// preoccupied agent movement policy
// TODO: moves towards closest in cluster
func (pa *PreoccupiedAgent) GetTargetPosition() (infra.PositionVector, bool) {
	// occupied := grid.GetAllOccupiedAgentPositions()

	var closestInCluster infra.IExtendedAgent = nil
	minDist := math.Inf(1)

	for otherID, otherAgent := range pa.GetAgentMap() {
		// Ignore agents outside of cluster
		if otherAgent.GetClusterID() != pa.clusterID {
			continue
		}

		// Ignore self
		if otherID == pa.GetID() {
			continue
		}

		dist := pa.position.Dist(otherAgent.GetPosition())
		if dist < minDist {
			minDist = dist
			closestInCluster = otherAgent
		}
	}

	if closestInCluster == nil {
		return infra.PositionVector{}, false
	}

	return closestInCluster.GetPosition(), true
}
