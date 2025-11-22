package server

import (
	"fmt"
	"maps"
	"math"
	"math/rand"

	"github.com/MattSScott/basePlatformSOMAS/v2/pkg/server"

	"github.com/MattSScott/TMT_SOMAS/config"
	"github.com/MattSScott/TMT_SOMAS/infra"
	"github.com/google/uuid"
)

type TMTServer struct {
	*server.BaseServer[infra.IExtendedAgent]
	config           config.Config
	grid             *infra.Grid
	clusterMap       map[int][]uuid.UUID // Map of cluster IDs to agent IDs
	expectedChildren float64
	gameRecorder     *infra.GameJSONRecord
}

func CreateTMTServer(config config.Config) *TMTServer {
	return &TMTServer{
		BaseServer:       server.CreateBaseServer[infra.IExtendedAgent](config.NumIterations, config.NumTurns, 0, 0),
		config:           config,
		grid:             infra.NewGrid(config.GridWidth, config.GridHeight),
		clusterMap:       make(map[int][]uuid.UUID),
		expectedChildren: config.InitialExpectedChildren,
		gameRecorder:     infra.MakeGameRecord(config),
	}
}

func (tserv *TMTServer) Start() {
	// Initialize social network after agents are created
	for _, ag := range tserv.GetAgentMap() {
		tserv.InitialiseRandomNetworkForAgent(ag)
	}
	tserv.BaseServer.Start()
	err := infra.WriteJSONLog("JSONlogs", tserv.gameRecorder)
	if err != nil {
		fmt.Println(tserv.config)
		panic(err)
	}
}

func (tserv *TMTServer) GetAgentByID(agentID uuid.UUID) (infra.IExtendedAgent, bool) {
	agentMap := tserv.GetAgentMap()
	agent, exists := agentMap[agentID]
	return agent, exists
}

func (tserv *TMTServer) GetASPThreshold() float32 {
	return float32(tserv.config.ASPThreshold)
}

func (tserv *TMTServer) InitialiseRandomNetworkForAgent(agent infra.IExtendedAgent) {
	thisAgentID := agent.GetID()
	// add self to network
	agent.AddToSocialNetwork(thisAgentID, 0.5)

	// add others with probability p
	for otherAgentID := range tserv.GetAgentMap() {
		// avoid overwriting existing connection
		if agent.ExistsInNetwork(otherAgentID) {
			continue
		}
		probability := rand.Float64() // Generate a random number
		// Connect with probability p
		if probability > tserv.config.ConnectionProbability {
			continue
		}
		// Assign a random relationship strength (0.2 to 1.0)
		strength1 := 0.2 + rand.Float32()*0.8
		tserv.CreateNetworkConnection(thisAgentID, otherAgentID, strength1)
		strength2 := 0.2 + rand.Float32()*0.8
		tserv.CreateNetworkConnection(otherAgentID, thisAgentID, strength2)
	}
}

func (tserv *TMTServer) RunStartOfIteration(iteration int) {
	// fmt.Println(iteration, len(tserv.GetAgentMap()))
	tserv.gameRecorder.PrepareNewIteration(iteration)

	if tserv.config.Debug {
		fmt.Printf("--------Start of iteration %d---------\n", iteration)
		fmt.Printf("Total agents: %d\n", len(tserv.GetAgentMap()))
	}
}

func getStep(current, target int) int {
	if target > current {
		return 1
	} else if target < current {
		return -1
	}
	return 0
}

func (tServ *TMTServer) moveIsValid(moveX, moveY int) bool {
	grid := tServ.grid
	if moveX < 0 || moveX >= grid.Width {
		return false
	}
	if moveY < 0 || moveY >= grid.Height {
		return false
	}
	return !grid.IsOccupied(moveX, moveY)
}

func (tserv *TMTServer) RunTurn(i, j int) {
	if tserv.config.Debug {
		fmt.Printf("Iteration %d, Turn %d\n", i, j)
	}
	tserv.moveAgents()
	tserv.recordTurnJSON(j)
}

func (tserv *TMTServer) RunEndOfIteration(iter int) {
	if tserv.config.Debug {
		fmt.Printf("--------End of iteration %v---------\n", iter)
	}
	initialPop := len(tserv.GetAgentMap())

	// 2. Apply clustering (k-means)
	tserv.applyClustering()

	// 4. Check for agent elimination
	tserv.updateAgentMortality()
	// 4.1 - natural deaths (old age)
	naturalDeathReport := tserv.getNaturalEliminationReport()
	tserv.applyElimination(naturalDeathReport)

	// 4.2 - unnatural deaths (sacrifice)
	sacrificialDeathReport, numberVolunteers := tserv.getSacrificialEliminationReport()
	tserv.applyElimination(sacrificialDeathReport)

	// 4.3 - create tombstones / temples
	fullDeathReport := make(map[uuid.UUID]infra.DeathInfo, len(naturalDeathReport)+len(sacrificialDeathReport))
	maps.Copy(fullDeathReport, naturalDeathReport)
	maps.Copy(fullDeathReport, sacrificialDeathReport)
	tserv.performSacrifices(fullDeathReport)

	// 5. After eliminations for agents in each cluster:
	for _, agents := range tserv.clusterMap {
		// 5.1 - Update social network (create/ cut links)
		tserv.updateSocialNetwork(agents)
		// 5.2 - Apply PTS protocol
		tserv.applyPTS(agents)
	}

	// 6. Update agent parameters
	tserv.updateClusterEliminations(fullDeathReport)
	tserv.updateAgentYsterofimia(fullDeathReport)
	tserv.pruneNetwork(fullDeathReport)

	// 7. Spawn new agents
	tserv.updateProbabilityOfChildren(numberVolunteers, initialPop)

	// Age up all agents
	for _, agent := range tserv.GetAgentMap() {
		agent.IncrementAge()
	}

	newAgents := tserv.generateNewAgents(fullDeathReport)

	newPop := initialPop + len(newAgents)
	tserv.updateAgentWorldviews(initialPop, newPop)

	tserv.spawnNewAgents(newAgents)

	tserv.addIterationJSON(numberVolunteers, fullDeathReport)
}

func (tserv *TMTServer) spawnNewAgents(newAgents []infra.IExtendedAgent) {
	for _, ag := range newAgents {
		tserv.AddAgent(ag)
	}

	for _, ag := range newAgents {
		tserv.InitialiseRandomNetworkForAgent(ag)
	}
}

// ---------------------- Helper Functions ----------------------
func runKMeans(positionMap map[uuid.UUID]infra.PositionVector, numClusters int) map[uuid.UUID]int {
	numPositions := len(positionMap)
	if numPositions == 0 {
		return nil
	}
	// ----- Initialize centroids -----
	// sample randomly from list of agent positions
	availablePositions := make([]infra.PositionVector, 0)
	for _, pos := range positionMap {
		availablePositions = append(availablePositions, pos)
	}
	centroids := make([]*infra.Centroid, numClusters)
	for i := range numClusters {
		randPos := rand.Intn(numPositions)
		samplePoint := availablePositions[randPos]
		centroids[i] = samplePoint.PositionVectorToCentroid()
	}

	// ----- Perform K-Means -----
	clusterAssignments := make(map[uuid.UUID]int)
	changed := true

	for changed {
		changed = false

		// Assign points
		for agentID, agentPos := range positionMap {
			minDist := math.MaxFloat64
			best := -1
			for j, centroid := range centroids {
				agentDist := agentPos.CentroidDist(centroid)
				if agentDist < minDist {
					minDist = agentDist
					best = j
				}
			}

			if assignedPosition, agentIsAssigned := clusterAssignments[agentID]; !agentIsAssigned || assignedPosition != best {
				clusterAssignments[agentID] = best
				changed = true
			}

		}

		// ----- Update Clusters -----
		// record mean cluster position with total size...
		clusterSize := make([]int, numClusters)
		// ...and number of elements
		clusterSum := make([]*infra.PositionVector, numClusters)

		for i := range clusterSum {
			clusterSum[i] = &infra.PositionVector{X: 0, Y: 0}
		}

		for agentID, assignedCluster := range clusterAssignments {
			agentPos := positionMap[agentID]
			clusterSum[assignedCluster].X += agentPos.X
			clusterSum[assignedCluster].Y += agentPos.Y
			clusterSize[assignedCluster]++
		}

		for i := range numClusters {
			size := clusterSize[i]
			if size == 0 {
				continue
			}
			centroids[i].X = float64(clusterSum[i].X) / float64(size)
			centroids[i].Y = float64(clusterSum[i].Y) / float64(size)
		}
	}

	return clusterAssignments
}

func (tserv *TMTServer) moveAgents() {
	for _, agent := range tserv.GetAgentMap() {
		agentPos := agent.GetPosition()
		moveX, moveY := tserv.grid.GetValidMove(agentPos.X, agentPos.Y)
		targetPos, posExists := agent.GetTargetPosition()

		if posExists {
			attemptX := agentPos.X + getStep(agentPos.X, targetPos.X)
			attemptY := agentPos.Y + getStep(agentPos.Y, targetPos.Y)
			if tserv.moveIsValid(attemptX, attemptY) {
				moveX, moveY = attemptX, attemptY
			}
		}

		newPos := infra.PositionVector{X: moveX, Y: moveY}
		tserv.grid.UpdateAgentPosition(agent, newPos)
		agent.SetPosition(newPos)
	}
}

func (tserv *TMTServer) applyClustering() {
	agentMap := tserv.GetAgentMap()
	if len(agentMap) == 0 {
		return // Nothing to cluster
	}

	agentPositionMap := make(map[uuid.UUID]infra.PositionVector)

	for agentID, agent := range agentMap {
		pos := agent.GetPosition()
		agentPositionMap[agentID] = pos
	}

	clusterAssignments := runKMeans(agentPositionMap, tserv.config.NumClusters)

	for agentID, assigment := range clusterAssignments {
		if agent, ok := tserv.GetAgentByID(agentID); ok {
			agent.SetClusterID(assigment)
		}
	}

	tserv.clusterMap = make(map[int][]uuid.UUID)
	for _, agent := range tserv.GetAgentMap() {
		tserv.clusterMap[agent.GetClusterID()] = append(tserv.clusterMap[agent.GetClusterID()], agent.GetID())
	}
}

func (tserv *TMTServer) updateSocialNetwork(cluster []uuid.UUID) {
	for _, agentID := range cluster {
		agent, ok := tserv.GetAgentByID(agentID)
		if !ok || !agent.IsAlive() {
			continue
		}

		agentSocialNetwork := agent.GetNetwork()

		for _, otherID := range cluster {
			if otherID == agentID {
				continue
			}

			_, connected := agentSocialNetwork[otherID]

			// add link based on anxiety
			if !connected && rand.Float32() < agent.GetAttachment().Anxiety {
				// Add symmetric relationship with random strength
				strength := 0.2 + rand.Float32()*0.8
				tserv.CreateNetworkConnection(agentID, otherID, strength)
				// tserv.CreateBidirectionalConnection(agentID, otherID, strength)
			}

			// remove link based on attachment
			if connected && rand.Float32() < agent.GetAttachment().Avoidance {
				tserv.SeverNetworkConnection(agentID, otherID)
				// agent.RemoveRelationship(otherID)
				// tserv.RemoveRelationship(agentID, otherID)
			}
		}
	}
}

func (tserv *TMTServer) applyPTS(cluster []uuid.UUID) {
	receivedCheck := make(map[uuid.UUID]bool) // Track who got a check
	// get ExtendedAgent from agentID
	for _, senderID := range cluster {
		sender, ok := tserv.GetAgentByID(senderID)
		if !ok {
			continue
		}
		if rand.Float32() < sender.GetPTSParams().CheckProb {
			for _, receiverID := range cluster {
				if receiverID == senderID {
					continue // don't send to self
				}
				receiver, ok := tserv.GetAgentByID(receiverID)
				if !ok {
					continue
				}
				// send wellbeing check message
				msg := sender.CreateWellbeingCheckMessage()
				sender.SendSynchronousMessage(msg, receiver.GetID())
				// mark agent as getting checked on
				receivedCheck[receiver.GetID()] = true
			}
		}
	}

	//update beta for agents that didn't get checked on
	for _, agentID := range cluster {
		if receivedCheck[agentID] {
			continue // skip agents who received a check
		}
		agent, ok := tserv.GetAgentByID(agentID)
		if !ok {
			continue
		}
		//update beta
		for neighbourID := range agent.GetNetwork() {
			agent.UpdateSocialNetwork(neighbourID, false)
		}
	}
}

func (tserv *TMTServer) updateAgentWorldviews(initialPop, newPop int) {
	trendChange := float64(newPop) / float64(tserv.config.NumAgents)
	seasonalChange := newPop - initialPop
	for _, agent := range tserv.GetAgentMap() {
		agent.UpdateWorldview(trendChange, seasonalChange)
	}
}

func (tserv *TMTServer) updateProbabilityOfChildren(numVolunteers, initPop int) {
	proportionOfVolunteers := float64(numVolunteers) / float64(initPop)

	if proportionOfVolunteers >= tserv.config.PopulationRho {
		tserv.expectedChildren = math.Min(tserv.expectedChildren*1.05, tserv.config.MaxExpectedChildren)
	} else {
		tserv.expectedChildren = math.Max(tserv.expectedChildren*0.95, tserv.config.MinExpectedChildren)
	}
}

func (tserv *TMTServer) GetInitNumberAgents() int {
	return tserv.config.NumAgents
}

func (tserv *TMTServer) GetGridDims() (int, int) {
	return tserv.config.GridWidth, tserv.config.GridHeight
}

// ---------------------- Recording Turn Data ----------------------

func (tserv *TMTServer) recordTurnJSON(turn int) {
	allAgentRecords := make([]infra.JSONAgentRecord, 0)

	for _, agent := range tserv.GetAgentMap() {
		record := agent.RecordAgentJSON(agent)
		allAgentRecords = append(allAgentRecords, record)
	}

	tombstonePositions := make([]infra.PositionVector, len(tserv.grid.Tombstones))
	copy(tombstonePositions, tserv.grid.Tombstones)

	templePositions := make([]infra.PositionVector, len(tserv.grid.Temples))
	copy(templePositions, tserv.grid.Temples)

	turnRecord := &infra.TurnJSONRecord{
		Turn:               turn,
		Agents:             allAgentRecords,
		TombstoneLocations: tombstonePositions,
		TempleLocations:    templePositions,
	}

	tserv.gameRecorder.WriteTurnRecord(turnRecord)
}

func (tserv *TMTServer) addIterationJSON(numVolunteers int, deathReport map[uuid.UUID]infra.DeathInfo) {
	totalAgents := len(tserv.GetAgentMap())
	reqElims := int(tserv.config.PopulationRho * float64(totalAgents))

	allElims := make([]uuid.UUID, 0)
	selfSacs := make([]uuid.UUID, 0)

	for id, report := range deathReport {
		allElims = append(allElims, id)
		if report.WasVoluntary {
			selfSacs = append(selfSacs, id)
		}
	}

	iterationDump := infra.EndOfIterationDump{
		NumberOfAgents:            totalAgents,
		EliminatedAgents:          allElims,
		SelfSacrificedAgents:      selfSacs,
		TotalVolunteers:           numVolunteers,
		TotalRequiredEliminations: reqElims,
	}

	tserv.gameRecorder.DumpIteration(iterationDump)
}

// func agentsToStrings(agents []infra.IExtendedAgent) []string {
// 	result := make([]string, len(agents))
// 	for i, agent := range agents {
// 		result[i] = agent.GetID().String()
// 	}
// 	return result
// }
