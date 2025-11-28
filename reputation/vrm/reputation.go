package vrm

import (
	"blockEmulator/params"
	"math"
	"sync"
	"time"
)

// ReputationLevel 信誉等级
type ReputationLevel uint8

const (
	LevelQuarantined ReputationLevel = iota // 隔离级 R=0
	LevelCandidate                          // 候选级 0<R<0.2
	LevelStandard                           // 标准级 0.2≤R<0.8
	LevelTrusted                            // 可信级 0.8≤R≤1.0
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

// BehaviorType 行为类型
type BehaviorType uint8

const (
	// 恶意行为 (Category 0)
	BehaviorMaliciousConsensus BehaviorType = iota // 严重共识作恶 δ=0.8
	BehaviorFalseData                              // 虚假信息 δ=0.5
	BehaviorUnexpectedOffline                      // 意外离线 δ=0.1

	// 积极行为 (Category 1)
	BehaviorConsensusSuccess // 成功参与共识 δ=0.05
	BehaviorDataVerified     // 数据被证实 δ=0.01
	BehaviorVerifyOthers     // 验证他人数据 δ=0.005
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

// ReputationScore 信誉分数
type ReputationScore struct {
	Score             float64         // 信誉分 [0, 1]
	Level             ReputationLevel // 信誉等级
	LocalInteractions uint64          // 本地交互次数（仅本地信誉有效）
	LastActiveTime    int64           // 最后活跃时间（Unix 时间戳）
	LastUpdateTime    int64           // 最后更新时间
}

// ReputationManager VRM 信誉管理器
type ReputationManager struct {
	Mu sync.RWMutex

	// 全局信誉存储 nodeID -> GlobalReputation
	GlobalReps map[uint64]*ReputationScore

	// 本地信誉存储 shardID -> nodeID -> LocalReputation
	LocalReps map[uint64]map[uint64]*ReputationScore

	// 配置参数
	Config *VRMConfig
}

// VRMConfig VRM 配置参数
type VRMConfig struct {
	InitialScore    float64                  // 初始信誉分
	DecayLambda     float64                  // 衰减系数 λ
	WeightDecayRate float64                  // 动态权重衰减率 α
	BehaviorWeights map[BehaviorType]float64 // 行为权重映射
}

// NewReputationManager 创建信誉管理器
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

// GetFinalScore 计算最终信誉分
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
		// 新节点加入分片，仅使用全局信誉
		return globalRep.Score
	}

	// 计算动态权重：w = e^(-α * N_local)
	w := math.Exp(-rm.Config.WeightDecayRate * float64(localRep.LocalInteractions))

	// 加权平均
	finalScore := w*globalRep.Score + (1-w)*localRep.Score

	return finalScore
}

// GetReputationLevel 获取信誉等级
func (rm *ReputationManager) GetReputationLevel(nodeID, shardID uint64) ReputationLevel {
	score := rm.GetFinalScore(nodeID, shardID)
	return rm.scoreToLevel(score)
}

// IsTrusted 判断是否为可信节点
func (rm *ReputationManager) IsTrusted(nodeID, shardID uint64) bool {
	return rm.GetReputationLevel(nodeID, shardID) == LevelTrusted
}

// UpdateReputation 更新信誉（核心算法）
// 同时更新全局信誉和本地信誉
func (rm *ReputationManager) UpdateReputation(
	nodeID, shardID uint64,
	behavior BehaviorType,
) {
	rm.Mu.Lock()
	defer rm.Mu.Unlock()

	// 获取或创建信誉记录
	globalRep := rm.getOrCreateGlobalRep(nodeID)
	localRep := rm.getOrCreateLocalRep(shardID, nodeID)

	// 获取行为权重
	baseWeight := rm.Config.BehaviorWeights[behavior]

	// 根据行为类别选择更新策略
	if rm.IsPositiveBehavior(behavior) {
		// 积极行为：边际效益递减
		// R_new = R + δ * (1 - R)
		globalRep.Score = rm.updatePositive(globalRep.Score, baseWeight)
		localRep.Score = rm.updatePositive(localRep.Score, baseWeight)
	} else {
		// 消极行为：固定惩罚
		// R_new = (1 - δ) * R
		globalRep.Score = rm.updateNegative(globalRep.Score, baseWeight)
		localRep.Score = rm.updateNegative(localRep.Score, baseWeight)
	}

	// 更新元数据
	now := time.Now().Unix()
	localRep.LocalInteractions++
	globalRep.LastUpdateTime = now
	localRep.LastUpdateTime = now
	globalRep.LastActiveTime = now
	localRep.LastActiveTime = now

	// 更新等级
	globalRep.Level = rm.scoreToLevel(globalRep.Score)
	localRep.Level = rm.scoreToLevel(localRep.Score)
}

// ApplyDecay 应用时间衰减
// R_new = R * λ^Δt
func (rm *ReputationManager) ApplyDecay(nodeID uint64, deltaTime int64) {
	rm.Mu.Lock()
	defer rm.Mu.Unlock()

	rep := rm.GlobalReps[nodeID]
	if rep == nil {
		return
	}

	// 计算衰减因子
	decayFactor := math.Pow(rm.Config.DecayLambda, float64(deltaTime))
	rep.Score *= decayFactor

	// 更新等级
	rep.Level = rm.scoreToLevel(rep.Score)
}

// GetGlobalReputation 获取全局信誉分
func (rm *ReputationManager) GetGlobalReputation(nodeID uint64) float64 {
	rm.Mu.RLock()
	defer rm.Mu.RUnlock()

	if rep, exists := rm.GlobalReps[nodeID]; exists {
		return rep.Score
	}
	return rm.Config.InitialScore
}

// GetLocalReputation 获取本地信誉分
func (rm *ReputationManager) GetLocalReputation(nodeID, shardID uint64) float64 {
	rm.Mu.RLock()
	defer rm.Mu.RUnlock()

	localRep := rm.getLocalRep(shardID, nodeID)
	if localRep == nil {
		return rm.Config.InitialScore
	}
	return localRep.Score
}

// --- 内部方法 ---

// updatePositive 积极行为更新（边际效益递减）
func (rm *ReputationManager) updatePositive(score, weight float64) float64 {
	newScore := score + weight*(1-score)
	return math.Min(newScore, 1.0)
}

// updateNegative 消极行为更新（固定惩罚）
func (rm *ReputationManager) updateNegative(score, weight float64) float64 {
	newScore := (1 - weight) * score
	return math.Max(newScore, 0.0)
}

// IsPositiveBehavior 判断是否为积极行为
func (rm *ReputationManager) IsPositiveBehavior(bt BehaviorType) bool {
	return bt >= BehaviorConsensusSuccess
}

// scoreToLevel 信誉分转换为等级
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

// getOrCreateGlobalRep 获取或创建全局信誉记录
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

// getOrCreateLocalRep 获取或创建本地信誉记录
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

// getLocalRep 获取本地信誉记录（不创建）
func (rm *ReputationManager) getLocalRep(shardID, nodeID uint64) *ReputationScore {
	if shardMap, exists := rm.LocalReps[shardID]; exists {
		return shardMap[nodeID]
	}
	return nil
}

// DefaultVRMConfig 返回默认配置
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
