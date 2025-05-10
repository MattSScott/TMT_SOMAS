package main

import (
	"github.com/MattSScott/TMT_SOMAS/agents"
	"github.com/MattSScott/TMT_SOMAS/config"
	"github.com/MattSScott/TMT_SOMAS/infra"
	"github.com/MattSScott/TMT_SOMAS/server"
)

func main() {
	config := config.NewConfig()
	serv := server.CreateTMTServer(config)
	serv.SetGameRunner(serv)

	total := config.NumAgents
	sum := config.SecureRatio + config.DismissiveRatio + config.PreoccupiedRatio + config.FearfulRatio
	if sum != 1.0 {
		panic("The sum of the agent ratios must equal 1.0")
	}
	numSecure := int(float64(total) * config.SecureRatio)
	numDismissive := int(float64(total) * config.DismissiveRatio)
	numPreoccupied := int(float64(total) * config.PreoccupiedRatio)
	numFearful := int(float64(total) * config.FearfulRatio)

	agentPopulation := make([]infra.IExtendedAgent, 0)

	for i := 0; i < numSecure; i++ {
		agentPopulation = append(agentPopulation, agents.CreateSecureAgent(serv))
	}
	for i := 0; i < numDismissive; i++ {
		agentPopulation = append(agentPopulation, agents.CreateDismissiveAgent(serv))
	}
	for i := 0; i < numPreoccupied; i++ {
		agentPopulation = append(agentPopulation, agents.CreatePreoccupiedAgent(serv))
	}
	for i := 0; i < numFearful; i++ {
		agentPopulation = append(agentPopulation, agents.CreateFearfulAgent(serv))
	}

	// for i := 0; i < config.NumAgents; i += 4 {
	// 	agentPopulation = append(agentPopulation, agents.CreateSecureAgent(serv))
	// 	agentPopulation = append(agentPopulation, agents.CreateDismissiveAgent(serv))
	// 	agentPopulation = append(agentPopulation, agents.CreatePreoccupiedAgent(serv))
	// 	agentPopulation = append(agentPopulation, agents.CreateFearfulAgent(serv))
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