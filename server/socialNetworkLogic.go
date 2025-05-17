package server

import (
	"github.com/google/uuid"
)

func (serv *TMTServer) CreateNetworkConnection(fromAgentID, toAgentID uuid.UUID, strength float32) {
	agentMap := serv.GetAgentMap()
	fromAgent, fromExists := agentMap[fromAgentID]
	toAgent, toExists := agentMap[toAgentID]

	if fromExists {
		fromAgent.AddToSocialNetwork(toAgentID, strength)
		fromAgent.PerformCreatedConnection(toAgentID)
	}

	if toExists {
		toAgent.ReceiveCreatedConnection(fromAgentID)
	}
}

func (serv *TMTServer) SeverNetworkConnection(fromAgentID, toAgentID uuid.UUID) {
	agentMap := serv.GetAgentMap()
	fromAgent, fromExists := agentMap[fromAgentID]
	toAgent, toExists := agentMap[toAgentID]

	if fromExists {
		fromAgent.RemoveFromSocialNetwork(toAgentID)
	}

	if toExists {
		toAgent.PerformSeveredConnected(fromAgentID)
		toAgent.ReceiveSeveredConnected(fromAgentID)
	}
}

func (serv *TMTServer) SubmitDecisionThreshold(agentID uuid.UUID, score float64) {
	serv.gameRecorder.WriteDecisionThreshold(agentID, score)
}
