package main

import (
	"github.com/MattSScott/TMT_SOMAS/agents"
	"github.com/MattSScott/TMT_SOMAS/config"
	"github.com/MattSScott/TMT_SOMAS/infra"
	"github.com/MattSScott/TMT_SOMAS/server"
	"github.com/google/uuid"
)

func main() {
	config := config.NewConfig()
	serv := server.CreateTMTServer(config)
	serv.SetGameRunner(serv)

	total := config.NumAgents
	numSecure := int(float64(total) * config.SecureRatio)
	numDismissive := int(float64(total) * config.DismissiveRatio)
	numPreoccupied := int(float64(total) * config.PreoccupiedRatio)
	numFearful := total - (numSecure + numDismissive + numPreoccupied)

	parent1ID, parent2ID := uuid.Nil, uuid.Nil
	agentPopulation := make([]infra.IExtendedAgent, 0)

	for i := 0; i < numSecure; i++ {
		agentPopulation = append(agentPopulation, agents.CreateSecureAgent(serv, parent1ID, parent2ID))
	}
	for i := 0; i < numDismissive; i++ {
		agentPopulation = append(agentPopulation, agents.CreateDismissiveAgent(serv, parent1ID, parent2ID))
	}
	for i := 0; i < numPreoccupied; i++ {
		agentPopulation = append(agentPopulation, agents.CreatePreoccupiedAgent(serv, parent1ID, parent2ID))
	}
	for i := 0; i < numFearful; i++ {
		agentPopulation = append(agentPopulation, agents.CreateFearfulAgent(serv, parent1ID, parent2ID))
	}

	// for i := 0; i < config.NumAgents; i += 4 {
	// 	agentPopulation = append(agentPopulation, agents.CreateSecureAgent(serv, parent1ID, parent2ID))
	// 	agentPopulation = append(agentPopulation, agents.CreateDismissiveAgent(serv, parent1ID, parent2ID))
	// 	agentPopulation = append(agentPopulation, agents.CreatePreoccupiedAgent(serv, parent1ID, parent2ID))
	// 	agentPopulation = append(agentPopulation, agents.CreateFearfulAgent(serv, parent1ID, parent2ID))
	// }

	for _, agent := range agentPopulation {
		serv.AddAgent(agent)
		if config.Debug {
			agent.AgentInitialised()
		}
	}

	// Start server
	serv.Start()
}
