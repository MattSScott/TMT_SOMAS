package agents

import (
	"fmt"
	"math"

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
// TODO: moves away from closest in cluster
func (fa *FearfulAgent) GetTargetPosition() (infra.PositionVector, bool) {
	// occupied := grid.GetAllOccupiedAgentPositions()

	var closestInCluster infra.IExtendedAgent = nil
	minDist := math.Inf(1)

	for otherID, otherAgent := range fa.GetAgentMap() {
		// Ignore agents outside of cluster
		if otherAgent.GetClusterID() != fa.clusterID {
			continue
		}

		// Ignore self
		if otherID == fa.GetID() {
			continue
		}

		dist := fa.position.Dist(otherAgent.GetPosition())
		if dist < minDist {
			minDist = dist
			closestInCluster = otherAgent
		}
	}

	if closestInCluster == nil {
		return infra.PositionVector{}, false
	}

	closestPos := closestInCluster.GetPosition()
	selfPos := fa.GetPosition()

	// closest->self + self == self - closest + self

	return selfPos.Sub(closestPos).Add(selfPos), true
}
