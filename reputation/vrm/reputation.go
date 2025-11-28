package vrm

import (
	"blockEmulator/params"
	"math"
	"sync"
	"time"
)

// ReputationLevel reputation level
type ReputationLevel uint8

const (
	LevelQuarantined ReputationLevel = iota // Quarantined level R=0
	LevelCandidate                          // Candidate level 0<R<0.2
	LevelStandard                           // Standard level 0.2≤R<0.8
	LevelTrusted                            // Trusted level 0.8≤R≤1.0
)

func (rl ReputationLevel) String() string {
	switch rl {
	case LevelQuarantined:
		return "Quarantined"
	case LevelCandidate:
		return "Candidate"
	case LevelStandard:
		return "Standard"
	case LevelTrusted:
		return "Trusted"
	default:
		return "Unknown"
	}
}

// BehaviorType behavior type
type BehaviorType uint8

const (
	// Malicious behaviors (Category 0)
	BehaviorMaliciousConsensus BehaviorType = iota // Severe consensus attack δ=0.8
	BehaviorFalseData                              // False information δ=0.5
	BehaviorUnexpectedOffline                      // Unexpected offline δ=0.1

	// Positive behaviors (Category 1)
	BehaviorConsensusSuccess // Successfully participated in consensus δ=0.05
	BehaviorDataVerified     // Data verified δ=0.01
	BehaviorVerifyOthers     // Verified others' data δ=0.005
)

func (bt BehaviorType) String() string {
	switch bt {
	case BehaviorMaliciousConsensus:
		return "MaliciousConsensus"
	case BehaviorFalseData:
		return "FalseData"
	case BehaviorUnexpectedOffline:
		return "UnexpectedOffline"
	case BehaviorConsensusSuccess:
		return "ConsensusSuccess"
	case BehaviorDataVerified:
		return "DataVerified"
	case BehaviorVerifyOthers:
		return "VerifyOthers"
	default:
		return "Unknown"
	}
}

// ReputationScore reputation score
type ReputationScore struct {
	Score             float64         // Reputation score [0, 1]
	Level             ReputationLevel // Reputation level
	LocalInteractions uint64          // Local interaction count (only valid for local reputation)
	LastActiveTime    int64           // Last active time (Unix timestamp)
	LastUpdateTime    int64           // Last update time
}

// ReputationManager VRM reputation manager
type ReputationManager struct {
	Mu sync.RWMutex

	// Global reputation storage nodeID -> GlobalReputation
	GlobalReps map[uint64]*ReputationScore

	// Local reputation storage shardID -> nodeID -> LocalReputation
	LocalReps map[uint64]map[uint64]*ReputationScore

	// Configuration parameters
	Config *VRMConfig
}

// VRMConfig VRM configuration parameters
type VRMConfig struct {
	InitialScore    float64                  // Initial reputation score
	DecayLambda     float64                  // Decay coefficient λ
	WeightDecayRate float64                  // Dynamic weight decay rate α
	BehaviorWeights map[BehaviorType]float64 // Behavior weight mapping
}

// NewReputationManager creates a reputation manager
func NewReputationManager(config *VRMConfig) *ReputationManager {
	if config == nil {
		config = DefaultVRMConfig()
	}

	return &ReputationManager{
		GlobalReps: make(map[uint64]*ReputationScore),
		LocalReps:  make(map[uint64]map[uint64]*ReputationScore),
		Config:     config,
	}
}

// GetFinalScore computes the final reputation score
// R_final = w * R_global + (1-w) * R_local
// w = e^(-α * N_local)
func (rm *ReputationManager) GetFinalScore(nodeID, shardID uint64) float64 {
	rm.Mu.RLock()
	defer rm.Mu.RUnlock()

	globalRep := rm.GlobalReps[nodeID]
	if globalRep == nil {
		return rm.Config.InitialScore
	}

	localRep := rm.getLocalRep(shardID, nodeID)
	if localRep == nil {
		// New node joining shard, use only global reputation
		return globalRep.Score
	}

	// Compute dynamic weight: w = e^(-α * N_local)
	w := math.Exp(-rm.Config.WeightDecayRate * float64(localRep.LocalInteractions))

	// Weighted average
	finalScore := w*globalRep.Score + (1-w)*localRep.Score

	return finalScore
}

// GetReputationLevel gets the reputation level
func (rm *ReputationManager) GetReputationLevel(nodeID, shardID uint64) ReputationLevel {
	score := rm.GetFinalScore(nodeID, shardID)
	return rm.scoreToLevel(score)
}

// IsTrusted checks if a node is trusted
func (rm *ReputationManager) IsTrusted(nodeID, shardID uint64) bool {
	return rm.GetReputationLevel(nodeID, shardID) == LevelTrusted
}

// UpdateReputation updates reputation (core algorithm)
// Updates both global and local reputation
func (rm *ReputationManager) UpdateReputation(
	nodeID, shardID uint64,
	behavior BehaviorType,
) {
	rm.Mu.Lock()
	defer rm.Mu.Unlock()

	// Get or create reputation record
	globalRep := rm.getOrCreateGlobalRep(nodeID)
	localRep := rm.getOrCreateLocalRep(shardID, nodeID)

	// Get behavior weight
	baseWeight := rm.Config.BehaviorWeights[behavior]

	// Choose update strategy based on behavior category
	if rm.IsPositiveBehavior(behavior) {
		// Positive behavior: diminishing marginal returns
		// R_new = R + δ * (1 - R)
		globalRep.Score = rm.updatePositive(globalRep.Score, baseWeight)
		localRep.Score = rm.updatePositive(localRep.Score, baseWeight)
	} else {
		// Negative behavior: fixed penalty
		// R_new = (1 - δ) * R
		globalRep.Score = rm.updateNegative(globalRep.Score, baseWeight)
		localRep.Score = rm.updateNegative(localRep.Score, baseWeight)
	}

	// Update metadata
	now := time.Now().Unix()
	localRep.LocalInteractions++
	globalRep.LastUpdateTime = now
	localRep.LastUpdateTime = now
	globalRep.LastActiveTime = now
	localRep.LastActiveTime = now

	// Update level
	globalRep.Level = rm.scoreToLevel(globalRep.Score)
	localRep.Level = rm.scoreToLevel(localRep.Score)
}

// ApplyDecay applies time decay
// R_new = R * λ^Δt
func (rm *ReputationManager) ApplyDecay(nodeID uint64, deltaTime int64) {
	rm.Mu.Lock()
	defer rm.Mu.Unlock()

	rep := rm.GlobalReps[nodeID]
	if rep == nil {
		return
	}

	// Compute decay factor
	decayFactor := math.Pow(rm.Config.DecayLambda, float64(deltaTime))
	rep.Score *= decayFactor

	// Update level
	rep.Level = rm.scoreToLevel(rep.Score)
}

// GetGlobalReputation gets global reputation score
func (rm *ReputationManager) GetGlobalReputation(nodeID uint64) float64 {
	rm.Mu.RLock()
	defer rm.Mu.RUnlock()

	if rep, exists := rm.GlobalReps[nodeID]; exists {
		return rep.Score
	}
	return rm.Config.InitialScore
}

// GetLocalReputation gets local reputation score
func (rm *ReputationManager) GetLocalReputation(nodeID, shardID uint64) float64 {
	rm.Mu.RLock()
	defer rm.Mu.RUnlock()

	localRep := rm.getLocalRep(shardID, nodeID)
	if localRep == nil {
		return rm.Config.InitialScore
	}
	return localRep.Score
}

// --- Internal methods ---

// updatePositive positive behavior update (diminishing marginal returns)
func (rm *ReputationManager) updatePositive(score, weight float64) float64 {
	newScore := score + weight*(1-score)
	return math.Min(newScore, 1.0)
}

// updateNegative negative behavior update (fixed penalty)
func (rm *ReputationManager) updateNegative(score, weight float64) float64 {
	newScore := (1 - weight) * score
	return math.Max(newScore, 0.0)
}

// IsPositiveBehavior checks if a behavior is positive
func (rm *ReputationManager) IsPositiveBehavior(bt BehaviorType) bool {
	return bt >= BehaviorConsensusSuccess
}

// scoreToLevel converts reputation score to level
func (rm *ReputationManager) scoreToLevel(score float64) ReputationLevel {
	switch {
	case score == 0:
		return LevelQuarantined
	case score < params.CandidateThreshold:
		return LevelCandidate
	case score < params.TrustedThreshold:
		return LevelStandard
	default:
		return LevelTrusted
	}
}

// getOrCreateGlobalRep gets or creates global reputation record
func (rm *ReputationManager) getOrCreateGlobalRep(nodeID uint64) *ReputationScore {
	if rep, exists := rm.GlobalReps[nodeID]; exists {
		return rep
	}

	rep := &ReputationScore{
		Score:          rm.Config.InitialScore,
		Level:          LevelQuarantined,
		LastActiveTime: time.Now().Unix(),
		LastUpdateTime: time.Now().Unix(),
	}
	rm.GlobalReps[nodeID] = rep
	return rep
}

// getOrCreateLocalRep gets or creates local reputation record
func (rm *ReputationManager) getOrCreateLocalRep(shardID, nodeID uint64) *ReputationScore {
	if _, exists := rm.LocalReps[shardID]; !exists {
		rm.LocalReps[shardID] = make(map[uint64]*ReputationScore)
	}

	if rep, exists := rm.LocalReps[shardID][nodeID]; exists {
		return rep
	}

	rep := &ReputationScore{
		Score:             rm.Config.InitialScore,
		Level:             LevelQuarantined,
		LocalInteractions: 0,
		LastActiveTime:    time.Now().Unix(),
		LastUpdateTime:    time.Now().Unix(),
	}
	rm.LocalReps[shardID][nodeID] = rep
	return rep
}

// getLocalRep gets local reputation record (without creating)
func (rm *ReputationManager) getLocalRep(shardID, nodeID uint64) *ReputationScore {
	if shardMap, exists := rm.LocalReps[shardID]; exists {
		return shardMap[nodeID]
	}
	return nil
}

// DefaultVRMConfig returns default configuration
func DefaultVRMConfig() *VRMConfig {
	return &VRMConfig{
		InitialScore:    params.InitialReputation,
		DecayLambda:     params.ReputationDecayLambda,
		WeightDecayRate: params.WeightDecayRate,
		BehaviorWeights: map[BehaviorType]float64{
			BehaviorMaliciousConsensus: 0.8,
			BehaviorFalseData:          0.5,
			BehaviorUnexpectedOffline:  0.1,
			BehaviorConsensusSuccess:   0.05,
			BehaviorDataVerified:       0.01,
			BehaviorVerifyOthers:       0.005,
		},
	}
}

// DefaultReputationScore returns default reputation score
func DefaultReputationScore() *ReputationScore {
	return &ReputationScore{
		Score:             params.InitialReputation,
		Level:             LevelCandidate,
		LocalInteractions: 0,
		LastActiveTime:    time.Now().Unix(),
		LastUpdateTime:    time.Now().Unix(),
	}
}
