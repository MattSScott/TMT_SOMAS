package infra

import (
	"github.com/MattSScott/TMT_SOMAS/gameRecorder"
	"github.com/MattSScott/basePlatformSOMAS/v2/pkg/agent"

	"github.com/google/uuid"
)

type IExtendedAgent interface {
	agent.IAgent[IExtendedAgent]

	//Getters
	GetName() uuid.UUID
	GetAttachment() Attachment
	GetNetwork() map[uuid.UUID]float32
	GetAge() int
	GetPosition() PositionVector
	SetPosition(PositionVector)
	GetWorldview() *Worldview
	UpdateWorldview(float64, int)
	GetYsterofimia() *Ysterofimia
	GetTelomere() float64
	IsAlive() bool

	GetTargetPosition() (PositionVector, bool)
	GetClusterID() int
	GetASPDecision(grid *Grid) ASPDecison
	GetPTSParams() PTSParams
	IncrementClusterEliminations(n int)
	IncrementNetworkEliminations(n int)
	IncrementHeroism()
	GetHeroism() int

	IncrementAge()
	SetClusterID(id int)
	MarkAsDead()

	// Social network functions
	AddToSocialNetwork(uuid.UUID, float32)
	ExistsInNetwork(uuid.UUID) bool
	UpdateSocialNetwork(id uuid.UUID, isCheck bool)
	RemoveFromSocialNetwork(agentID uuid.UUID)
	// Social network handlers
	PerformCreatedConnection(uuid.UUID)
	ReceiveCreatedConnection(uuid.UUID)
	PerformSeveredConnected(uuid.UUID)
	ReceiveSeveredConnected(uuid.UUID)

	//Message functions
	CreateWellbeingCheckMessage() *WellbeingCheckMessage
	CreateReplyMessage() *ReplyMessage
	HandleWellbeingCheckMessage(msg *WellbeingCheckMessage)
	HandleReplyMessage(msg *ReplyMessage)

	//Info
	AgentInitialised()

	// Data Recording
	//RecordAgentStatus(instance IExtendedAgent) gameRecorder.AgentRecord
	RecordAgentJSON(instance IExtendedAgent) gameRecorder.JSONAgentRecord
}
