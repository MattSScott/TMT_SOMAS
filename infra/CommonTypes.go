package infra

import (
	"fmt"
	"math"
	"math/bits"
	"math/rand"

	"github.com/google/uuid"
)

type PositionVector struct {
	X int
	Y int
}

type Centroid struct {
	X, Y float64
}

func (p1 PositionVector) Dist(p2 PositionVector) float64 {
	deltaX := float64(p1.X - p2.X)
	deltaY := float64(p1.Y - p2.Y)
	return math.Sqrt((deltaX * deltaX) + (deltaY * deltaY))
}

func (p1 PositionVector) Sub(p2 PositionVector) PositionVector {
	deltaX := p1.X - p2.X
	deltaY := p1.Y - p2.Y
	return PositionVector{X: deltaX, Y: deltaY}
}

func (p1 PositionVector) Add(p2 PositionVector) PositionVector {
	deltaX := p1.X + p2.X
	deltaY := p1.Y + p2.Y
	return PositionVector{X: deltaX, Y: deltaY}
}

func (pv PositionVector) PositionVectorToCentroid() *Centroid {
	return &Centroid{
		X: float64(pv.X),
		Y: float64(pv.Y),
	}
}

func (pv PositionVector) CentroidDist(c *Centroid) float64 {
	deltaX := float64(pv.X) - c.X
	deltaY := float64(pv.Y) - c.Y
	return math.Sqrt((deltaX * deltaX) + (deltaY * deltaY))
}

type Attachment struct {
	Anxiety   float32
	Avoidance float32
	Type      AttachmentType
}

// for use in Gompertz Death function - 1 - exp(-a/b * exp(kx - 1))
type Telomere struct {
	age              int
	alpha            float64
	beta             float64
	generationLength int
}

type PTSParams struct {
	CheckProb float32 // prob of sending a wellbeing check
	ReplyProb float32 // prob of replying to a check
	Alpha     float32 // reinforcement param
	Beta      float32 // reinforcement param
}

func NewTelomere() *Telomere {
	alpha := 0.001
	beta := 0.3
	return &Telomere{1, alpha, beta, 30}
}

func (t *Telomere) GetAge() int {
	return t.age
}

func (t *Telomere) IncrementAge() {
	t.age++
}

func (t *Telomere) GetProbabilityOfDeath() float64 {
	if t.age >= t.generationLength {
		return 1.0
	}
	return min(t.alpha*math.Exp(t.beta*float64(t.age)), 1)
}

type SocialNetwork map[uuid.UUID]float32

type ClusterEliminations struct {
	TotalEliminations []int // eliminations in this cluster per turn
	ClusterSizes      []int // size of this cluster per turn
}

type Ysterofimia struct {
	SelfSacrificeCount       int     // number of times agent volunteered self-sacrifices
	SelfSacrificeEsteems     float32 // total esteems from agents who self-sacrificed
	OtherEliminationCount    int     // number of times agent was eliminated by other than self-sacrifice
	OtherEliminationsEsteems float32 // total esteems from agents who eliminated other than self-sacrifice
}

func NewYsterofimia() *Ysterofimia {
	return &Ysterofimia{
		SelfSacrificeCount:       0,
		SelfSacrificeEsteems:     0.0,
		OtherEliminationCount:    0,
		OtherEliminationsEsteems: 0.0,
	}
}
func (y *Ysterofimia) IncrementSelfSacrificeCount() {
	y.SelfSacrificeCount++
}

func (y *Ysterofimia) AddSelfSacrificeEsteems(esteem float32) {
	y.SelfSacrificeEsteems += esteem
}
func (y *Ysterofimia) IncrementOtherEliminationCount() {
	y.OtherEliminationCount++
}
func (y *Ysterofimia) AddOtherEliminationsEsteems(esteem float32) {
	y.OtherEliminationsEsteems += esteem
}

func (y *Ysterofimia) ComputeYsterofimia() float32 {
	totalEliminations := y.SelfSacrificeCount + y.OtherEliminationCount
	totalEsteem := y.SelfSacrificeEsteems + y.OtherEliminationsEsteems

	if totalEliminations == 0 || totalEsteem == 0 {
		return 0.0 // no eliminations or no esteem data
	}

	averageVoluntaryEsteem := float32(y.SelfSacrificeEsteems) / float32(y.SelfSacrificeCount)
	averageInvoluntaryEsteem := float32(y.OtherEliminationsEsteems) / float32(y.OtherEliminationCount)

	if averageVoluntaryEsteem > averageInvoluntaryEsteem {
		return float32(y.SelfSacrificeCount) / float32(totalEliminations)
	} else {
		return float32(y.OtherEliminationCount) / float32(totalEliminations)
	}
}

type DeathInfo struct {
	Agent        IExtendedAgent
	WasVoluntary bool
}

type Worldview struct {
	worldviewHash    byte
	worldviewHistory []byte
	dunbarProportion float64
}

// low-frequency pop. variance - how does population chance across sim
func (wv *Worldview) getTrendWorldview(delta float64) byte {
	// 1 if within, 0 without
	if delta >= wv.dunbarProportion || delta <= 1/wv.dunbarProportion {
		return byte(0b10)
	}
	return byte(0b00)
}

// high-frequency - how does population chance from turn-to-turn
func (wv *Worldview) getSeasonalWorldview(delta int) byte {
	if delta > 0 {
		return byte(0b01)
	}
	return byte(0b00)
}

func (wv *Worldview) GetWorldviewHistory() []byte {
	return wv.worldviewHistory
}

func (wv *Worldview) UpdateWorldview(trendDelta float64, seasonalDelta int) {
	fullWorldviewData := wv.getTrendWorldview(trendDelta) | wv.getSeasonalWorldview(seasonalDelta)
	worldviewOpinion := ^(wv.worldviewHash ^ fullWorldviewData)
	wv.worldviewHistory = append(wv.worldviewHistory, worldviewOpinion)
}

func (wv1 *Worldview) CompareWorldviews(wv2 *Worldview) float64 {
	M, N := len(wv1.worldviewHistory), len(wv2.worldviewHistory)
	windowLen := min(M, N)
	if windowLen == 0 {
		return 0.0
	}
	totalBits := 2 * windowLen
	alignedBits := 0
	for i := range windowLen {
		wv1Data := wv1.worldviewHistory[M-i-1]
		wv2Data := wv2.worldviewHistory[N-i-1]
		alignment := ^(wv1Data ^ wv2Data) & 3
		alignedBits += bits.OnesCount8(alignment)
	}

	worldviewAlignment := float64(alignedBits) / float64(totalBits)

	if worldviewAlignment > 1.0 {
		misalignment := fmt.Sprintf("Invalid worldview alignment - Aligned bits: %d, Total bits: %d\n", alignedBits, totalBits)
		panic(misalignment)
	}
	return worldviewAlignment
}

func NewWorldview(hash byte) *Worldview {
	return &Worldview{
		worldviewHash:    hash,
		worldviewHistory: make([]byte, 0),
		dunbarProportion: rand.Float64() + 1,
	}
}

type PTS_Stats struct {
	createdBy int
	createdTo int
	severedBy int
	severedTo int
}

func (pts *PTS_Stats) IncrementCreatedBy() {
	pts.createdBy++
}

func (pts *PTS_Stats) IncrementCreatedTo() {
	pts.createdTo++
}

func (pts *PTS_Stats) IncrementSeveredBy() {
	pts.severedBy++
}

func (pts *PTS_Stats) IncrementSeveredTo() {
	pts.severedTo++
}

func (pts *PTS_Stats) GetEstrangement() float32 {
	propCreated := float32(pts.createdBy) / float32(pts.createdBy+pts.createdTo)
	propSevered := float32(pts.severedTo) / float32(pts.severedBy+pts.severedTo)
	return 0.5 * (propCreated + propSevered)
}

func NewPTS_Stats() *PTS_Stats {
	return &PTS_Stats{
		createdBy: 0,
		createdTo: 1,
		severedBy: 1,
		severedTo: 0,
	}
}

type EliminationHistory struct {
	observedEliminationsCluster int
	clusterEliminationTolerance int
	observedEliminationsNetwork int
	networkEliminationTolerance int
}

func (eh *EliminationHistory) IncrementClusterEliminations(n int) {
	eh.observedEliminationsCluster += n
}

func (eh *EliminationHistory) IncrementNetworkEliminations(n int) {
	eh.observedEliminationsNetwork += n
}

func (eh *EliminationHistory) GetClusterEliminationThreshold() float32 {
	propElim := float32(eh.observedEliminationsCluster) / float32(eh.clusterEliminationTolerance)
	return min(propElim, 1.0)
}

func (eh *EliminationHistory) GetNetworkEliminationThreshold() float32 {
	propElim := float32(eh.observedEliminationsNetwork) / float32(eh.networkEliminationTolerance)
	return min(propElim, 1.0)
}

func NewEliminationHistory(numAgents float64) *EliminationHistory {
	return &EliminationHistory{
		observedEliminationsCluster: 0,
		observedEliminationsNetwork: 0,
		clusterEliminationTolerance: int(numAgents * 0.5),
		networkEliminationTolerance: int(numAgents * 0.5),
	}
}
