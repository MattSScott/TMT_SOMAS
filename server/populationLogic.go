package server

import (
	"math/rand"
	"time"

	"github.com/MattSScott/TMT_SOMAS/agents"
	"github.com/MattSScott/TMT_SOMAS/infra"
	"github.com/google/uuid"
	"gonum.org/v1/gonum/stat/distuv"
)

func (tserv *TMTServer) updateAgentMortality() {
	for _, agent := range tserv.GetAgentMap() {
		probDeath := agent.GetTelomere()
		randVal := rand.Float64()
		if randVal < probDeath {
			// fmt.Printf("Agent age: %d, Death prob: %f\n", agent.GetAge(), probDeath)
			agent.MarkAsDead()
		}
	}
}

func (tserv *TMTServer) voluntarilySacrificeAgent(agent infra.IExtendedAgent) {
	pos := agent.GetPosition()
	tserv.grid.PlaceTemple(pos.X, pos.Y)
	tserv.lastEliminatedAgents = append(tserv.lastEliminatedAgents, agent)
	tserv.lastSelfSacrificedAgents = append(tserv.lastSelfSacrificedAgents, agent)
	// fmt.Printf("Agent %v has been eliminated (voluntary)\n", agent.GetID())
}

func (tserv *TMTServer) involuntarilySacrificeAgent(agent infra.IExtendedAgent) {
	pos := agent.GetPosition()
	tserv.grid.PlaceTombstone(pos.X, pos.Y)
	tserv.lastEliminatedAgents = append(tserv.lastEliminatedAgents, agent)
	// fmt.Printf("Agent %v has been eliminated (non-voluntary)\n", agent.GetID())
}

func (tserv *TMTServer) getNaturalEliminationReport() map[uuid.UUID]infra.DeathInfo {
	naturalReport := make(map[uuid.UUID]infra.DeathInfo)
	for agentID, agent := range tserv.GetAgentMap() {
		if !agent.IsAlive() {
			naturalReport[agentID] = infra.DeathInfo{Agent: agent, WasVoluntary: false}
		}
	}
	return naturalReport
}

func (tserv *TMTServer) stratifyVolunteers() ([]infra.IExtendedAgent, []infra.IExtendedAgent) {
	volunteers := make([]infra.IExtendedAgent, 0)
	nonVolunteers := make([]infra.IExtendedAgent, 0)

	// Separate volunteers and non-volunteers
	for _, agent := range tserv.GetAgentMap() {
		// don't allow naturally-dead agents to sacrifice
		if !agent.IsAlive() {
			continue
		}
		// check if agent volunteered
		if agent.GetASPDecision(tserv.grid) == infra.SELF_SACRIFICE {
			volunteers = append(volunteers, agent)
		} else {
			nonVolunteers = append(nonVolunteers, agent)
		}
	}
	return volunteers, nonVolunteers
}

func (tserv *TMTServer) getSacrificialEliminationReport() map[uuid.UUID]infra.DeathInfo {
	volunteers, nonVolunteers := tserv.stratifyVolunteers()
	sacrificialReport := make(map[uuid.UUID]infra.DeathInfo)

	totalAgents := float64(len(tserv.GetAgentMap()))
	neededVolunteers := int(tserv.config.PopulationRho * totalAgents)
	actualVolunteers := len(volunteers)
	// record number of volunteers
	tserv.numVolunteeredAgents = actualVolunteers

	// fmt.Println(totalAgents, neededVolunteers, actualVolunteers, tserv.expectedChildren)

	if actualVolunteers >= neededVolunteers {
		//randomly select n volunteers to eliminate
		rand.Shuffle(actualVolunteers, func(i, j int) {
			volunteers[i], volunteers[j] = volunteers[j], volunteers[i]
		})
		for i := range neededVolunteers {
			agent := volunteers[i]
			agentID := agent.GetID()
			sacrificialReport[agentID] = infra.DeathInfo{Agent: agent, WasVoluntary: true}
		}
	} else {
		//eliminate all volunteers...
		for _, agent := range volunteers {
			agentID := agent.GetID()
			sacrificialReport[agentID] = infra.DeathInfo{Agent: agent, WasVoluntary: true}
		}
		// ...plus 2*(n-v) random non-volunteers
		numNonVol := len(nonVolunteers)
		rand.Shuffle(numNonVol, func(i, j int) {
			nonVolunteers[i], nonVolunteers[j] = nonVolunteers[j], nonVolunteers[i]
		})

		for i := range min(numNonVol, 2*(neededVolunteers-actualVolunteers)) {
			agent := nonVolunteers[i]
			agentID := agent.GetID()
			sacrificialReport[agentID] = infra.DeathInfo{Agent: agent, WasVoluntary: false}
		}
	}

	return sacrificialReport
}

func (tserv *TMTServer) updateAgentYsterofimia(deathReport map[uuid.UUID]infra.DeathInfo) {
	for _, agent := range tserv.GetAgentMap() {
		networkEliminationCount := 0
		for friendID, esteem := range agent.GetNetwork() {
			// friend was eliminated (found in death report)
			if deathInfo, dead := deathReport[friendID]; dead {
				networkEliminationCount++
				ysterofimia := agent.GetYsterofimia()
				if deathInfo.WasVoluntary {
					ysterofimia.IncrementSelfSacrificeCount()
					ysterofimia.AddSelfSacrificeEsteems(esteem)
				} else {
					ysterofimia.IncrementOtherEliminationCount()
					ysterofimia.AddOtherEliminationsEsteems(esteem)
				}
			}
		}
		agent.IncrementNetworkEliminations(networkEliminationCount)
	}
}

func (tserv *TMTServer) updateClusterEliminations(deathReport map[uuid.UUID]infra.DeathInfo) {
	counts := make(map[int]int)
	for _, deathInfo := range deathReport {
		clusterID := deathInfo.Agent.GetClusterID()
		counts[clusterID]++
	}
	for _, agent := range tserv.GetAgentMap() {
		clusterID := agent.GetClusterID()
		if count, exists := counts[clusterID]; exists {
			agent.IncrementClusterEliminations(count)
		}
	}
}

func (tserv *TMTServer) removeFromNetwork(deadAgent infra.IExtendedAgent) {
	deadID := deadAgent.GetID()
	for aliveID := range tserv.GetAgentMap() {
		tserv.SeverNetworkConnection(aliveID, deadID)
	}
}

func (tserv *TMTServer) pruneNetwork(deathReport map[uuid.UUID]infra.DeathInfo) {
	for _, deathInfo := range deathReport {
		deadAgent := deathInfo.Agent
		tserv.removeFromNetwork(deadAgent)
	}
}

func (tserv *TMTServer) applyElimination(deathReport map[uuid.UUID]infra.DeathInfo) {
	for _, deathInfo := range deathReport {
		deadAgent := deathInfo.Agent
		tserv.RemoveAgent(deadAgent)
	}
}

func (tserv *TMTServer) performSacrifices(deathReport map[uuid.UUID]infra.DeathInfo) {
	tserv.lastEliminatedAgents = nil
	tserv.lastSelfSacrificedAgents = nil

	for _, deathInfo := range deathReport {
		deadAgent := deathInfo.Agent
		if deathInfo.WasVoluntary {
			tserv.voluntarilySacrificeAgent(deadAgent)
		} else {
			tserv.involuntarilySacrificeAgent(deadAgent)
		}
	}
}

func (tserv *TMTServer) generateNewAgents() []infra.IExtendedAgent {
	newAgents := make([]infra.IExtendedAgent, 0)

	dist := distuv.Poisson{
		Lambda: tserv.expectedChildren,
		Src:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	parentPool := tserv.lastEliminatedAgents
	poolSize := len(parentPool)

	rand.Shuffle(poolSize, func(i, j int) {
		parentPool[i], parentPool[j] = parentPool[j], parentPool[i]
	})

	for i := 1; i < poolSize; i += 2 {
		parent1 := parentPool[i-1]
		parent2 := parentPool[i]
		childrenToSpawn := int(dist.Rand())
		for range childrenToSpawn {
			newAgents = append(newAgents, tserv.generateChild(parent1, parent2))
		}
	}

	if poolSize%2 == 1 && poolSize > 1 {
		clonerAgent := parentPool[poolSize-1]
		newAgents = append(newAgents, tserv.generateChild(clonerAgent, clonerAgent))
	}

	return newAgents
}

func (tserv *TMTServer) getChildProbabilities(parent1, parent2 infra.AttachmentType) map[infra.AttachmentType]float64 {
	nonMutationRate := 1.0 - tserv.config.MutationRate
	// hashset to track chosen types
	chosenTypes := make(map[infra.AttachmentType]struct{})
	probs := make(map[infra.AttachmentType]float64)

	// account for parent 1 and parent 2
	probs[parent1] += nonMutationRate / 2
	probs[parent2] += nonMutationRate / 2

	chosenTypes[parent1] = struct{}{}
	chosenTypes[parent2] = struct{}{}

	remainingTypes := len(infra.AllAttachmentTypes) - len(chosenTypes)
	mutationChance := tserv.config.MutationRate / float64(remainingTypes)

	for _, attachType := range infra.AllAttachmentTypes {
		if _, seen := chosenTypes[attachType]; seen {
			continue
		}
		probs[attachType] = mutationChance
	}

	return probs
}

func (tserv *TMTServer) mixAttachmentTypes(parent1, parent2 infra.AttachmentType) infra.AttachmentType {
	probMap := tserv.getChildProbabilities(parent1, parent2)

	randVal := rand.Float64()
	cumulative := 0.0

	for attachType, prob := range probMap {
		cumulative += prob
		if randVal < cumulative {
			return attachType
		}
	}

	panic("Failed to select attachment type from probability map")

}

func (tserv *TMTServer) generateChild(parent1, parent2 infra.IExtendedAgent) infra.IExtendedAgent {
	type1 := parent1.GetAttachment().Type
	type2 := parent2.GetAttachment().Type
	childAttachmentType := tserv.mixAttachmentTypes(type1, type2)
	// childWorldview := tserv.mixWorldviews(parent1.GetWorldviewBinary(), parent2.GetWorldviewBinary())

	var newAgent infra.IExtendedAgent
	switch {
	case childAttachmentType == infra.SECURE:
		newAgent = agents.CreateSecureAgent(tserv)
	case childAttachmentType == infra.DISMISSIVE:
		newAgent = agents.CreateDismissiveAgent(tserv)
	case childAttachmentType == infra.PREOCCUPIED:
		newAgent = agents.CreatePreoccupiedAgent(tserv)
	case childAttachmentType == infra.FEARFUL:
		newAgent = agents.CreateFearfulAgent(tserv)
	default:
		newAgent = agents.CreateFearfulAgent(tserv)
	}

	return newAgent

	//add new agent to server
	// tserv.AddAgent(newAgent)
	// initialise agent's social network
	// tserv.InitialiseRandomNetworkForAgent(newAgent)

	// fmt.Println(len(newAgent.GetNetwork()))
}
