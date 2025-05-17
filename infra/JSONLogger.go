package infra

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/MattSScott/TMT_SOMAS/config"
	"github.com/google/uuid"
)

type JSONAgentRecord struct {
	ID                  string         `json:"ID"`
	Age                 int            `json:"Age"`
	AttachmentStyle     string         `json:"AttachmentStyle"`
	AttachmentAnxiety   float32        `json:"AttachmentAnxiety"`
	AttachmentAvoidance float32        `json:"AttachmentAvoidance"`
	ClusterID           int            `json:"ClusterID"`
	Position            PositionVector `json:"Position"`
	Worldview           uint32         `json:"Worldview"`
	Heroism             int            `json:"Heroism"`
}

type TurnJSONRecord struct {
	Turn               int               `json:"TurnNumber"`
	Agents             []JSONAgentRecord `json:"Agents"`
	TombstoneLocations []PositionVector  `json:"TombstoneLocations"`
	TempleLocations    []PositionVector  `json:"TempleLocations"`
}

type IterationJSONRecord struct {
	Iteration                 int                   `json:"Iteration"`
	Turns                     []*TurnJSONRecord     `json:"Turns"`
	Thresholds                map[uuid.UUID]float64 `json:"AgentThresholds"`
	NumberOfAgents            int                   `json:"NumberOfAgents"`
	EliminatedAgents          []uuid.UUID           `json:"EliminatedAgents"`
	SelfSacrificedAgents      []uuid.UUID           `json:"EliminatedBySelfSacrifice"`
	TotalVolunteers           int                   `json:"NumVolunteers"`
	TotalRequiredEliminations int                   `json:"TotalRequiredEliminations"`
}

type EndOfIterationDump struct {
	NumberOfAgents            int
	EliminatedAgents          []uuid.UUID
	SelfSacrificedAgents      []uuid.UUID
	TotalVolunteers           int
	TotalRequiredEliminations int
}

func NewIterationJSONRecord(iter int) *IterationJSONRecord {
	return &IterationJSONRecord{
		Iteration:                 iter,
		Turns:                     make([]*TurnJSONRecord, 0),
		Thresholds:                make(map[uuid.UUID]float64),
		NumberOfAgents:            0,
		EliminatedAgents:          make([]uuid.UUID, 0),
		SelfSacrificedAgents:      make([]uuid.UUID, 0),
		TotalVolunteers:           0,
		TotalRequiredEliminations: 0,
	}
}

func (ijr *IterationJSONRecord) addTurnData(tjr *TurnJSONRecord) {
	ijr.Turns = append(ijr.Turns, tjr)
}

func (ijr *IterationJSONRecord) writeDecisionThreshold(agentID uuid.UUID, score float64) {
	ijr.Thresholds[agentID] = score
}

type GameJSONRecord struct {
	Config     config.Config          `json:"Config"`
	Iterations []*IterationJSONRecord `json:"Iterations"`
}

func (gjr *GameJSONRecord) PrepareNewIteration(iter int) {
	gjr.Iterations = append(gjr.Iterations, NewIterationJSONRecord(iter))
}

func (gjr *GameJSONRecord) WriteDecisionThreshold(agentID uuid.UUID, score float64) {
	mostRecentIteration := gjr.Iterations[len(gjr.Iterations)-1]
	mostRecentIteration.writeDecisionThreshold(agentID, score)
}

func (gjr *GameJSONRecord) WriteTurnRecord(turnRecord *TurnJSONRecord) {
	mostRecentIteration := gjr.Iterations[len(gjr.Iterations)-1]
	mostRecentIteration.addTurnData(turnRecord)
}

func (gjr *GameJSONRecord) DumpIteration(dump EndOfIterationDump) {
	mostRecentIteration := gjr.Iterations[len(gjr.Iterations)-1]
	mostRecentIteration.NumberOfAgents = dump.NumberOfAgents
	mostRecentIteration.EliminatedAgents = dump.EliminatedAgents
	mostRecentIteration.SelfSacrificedAgents = dump.SelfSacrificedAgents
	mostRecentIteration.TotalVolunteers = dump.TotalVolunteers
	mostRecentIteration.TotalRequiredEliminations = dump.TotalRequiredEliminations
}

func MakeGameRecord(config config.Config) *GameJSONRecord {
	return &GameJSONRecord{
		Config:     config,
		Iterations: make([]*IterationJSONRecord, 0),
	}
}

func WriteJSONLog(outputDir string, record *GameJSONRecord) error {
	err := os.MkdirAll(outputDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	checkForNaN("GameJSON", record)

	fileName := fmt.Sprintf("%s/output.json", outputDir)
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling iteration JSON: %w", err)
	}

	return os.WriteFile(fileName, data, 0644)
}

func UUIDsToStrings(ids []uuid.UUID) []string {
	result := make([]string, len(ids))
	for i, id := range ids {
		result[i] = id.String()
	}
	return result
}
