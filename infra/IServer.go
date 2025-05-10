package infra

import (
	"github.com/MattSScott/basePlatformSOMAS/v2/pkg/agent"
	"github.com/google/uuid"
)

type IServer interface {
	agent.IExposedServerFunctions[IExtendedAgent]

	GetAgentByID(agentID uuid.UUID) (IExtendedAgent, bool)
	GetAgentMap() map[uuid.UUID]IExtendedAgent
	SubmitDecisionThreshold(uuid.UUID, float64)
	GetASPThreshold() float32
	GetInitNumberAgents() int
	GetGridDims() (int, int)
}
