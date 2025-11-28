package params

import "math/big"

// TriBFT 共识方法 ID
const ConsensusTriBFT = 4

// TriBFT 分层架构配置
var (
	// 分片配置
	RegionalShardNum = 16 // 区域分片数量
	CityClusterNum   = 4  // 城市集群数量（虚拟层）

	// 动态分片阈值
	ShardSplitThreshold = 100 // 节点数超过此值触发拆分
	ShardMergeThreshold = 20  // 节点数低于此值触发合并

	// VRM 信誉配置
	InitialReputation     = 0.05  // 初始信誉分
	ReputationDecayLambda = 0.995 // 衰减系数 λ
	TrustedThreshold      = 0.8   // 可信级阈值
	CandidateThreshold    = 0.2   // 候选级阈值
	WeightDecayRate       = 0.01  // 动态权重衰减率 α

	// HotStuff 配置
	HotStuffViewTimeout = 6000 // 视图超时 (ms)
	HotStuffDelta       = 150  // 网络延迟上界 δ (ms)
	QuorumRatio         = 0.67 // 法定人数比例 (2f+1)/N

	// 冗余节点配置
	BackupNodeRatio    = 0.3  // 冗余节点比例
	BackupSyncInterval = 1000 // 冗余节点同步间隔 (ms)

	// 凭证聚合周期
	CredentialAggregationPeriod = 10 // 每 N 个区块聚合一次凭证

	// 分层通信延迟配置（毫秒）
	IntraShardBaseDelay      = 30  // 分片内基础延迟
	IntraShardJitterRange    = 20  // 分片内抖动范围
	ShardToCityBaseDelay     = 50  // 分片到城市基础延迟
	ShardToCityJitterRange   = 40  // 分片到城市抖动范围
	CityToGlobalBaseDelay    = 80  // 城市到全局基础延迟
	CityToGlobalJitterRange  = 60  // 城市到全局抖动范围
	GlobalToShardBaseDelay   = 100 // 全局到分片基础延迟
	GlobalToShardJitterRange = 80  // 全局到分片抖动范围

	// 是否启用延迟日志
	EnableLayerDelayLogging = false
)

// ShardLevel 分片层级
type ShardLevel uint8

const (
	ShardLevelRegional ShardLevel = iota // 区域分片
	ShardLevelCity                       // 城市集群
	ShardLevelGlobal                     // 全局分片
)

func (sl ShardLevel) String() string {
	switch sl {
	case ShardLevelRegional:
		return "Regional"
	case ShardLevelCity:
		return "City"
	case ShardLevelGlobal:
		return "Global"
	default:
		return "Unknown"
	}
}

// LayerCommunicationType 层间通信类型
type LayerCommunicationType uint8

const (
	CommIntraShard    LayerCommunicationType = iota // 分片内通信
	CommShardToCity                                 // 分片到城市
	CommCityToGlobal                                // 城市到全局
	CommGlobalToShard                               // 全局到分片
)

func (lct LayerCommunicationType) String() string {
	switch lct {
	case CommIntraShard:
		return "IntraShard"
	case CommShardToCity:
		return "ShardToCity"
	case CommCityToGlobal:
		return "CityToGlobal"
	case CommGlobalToShard:
		return "GlobalToShard"
	default:
		return "Unknown"
	}
}

// TriBFTConfig TriBFT 链配置
type TriBFTConfig struct {
	ChainID    uint64
	NodeID     uint64
	ShardID    uint64
	ShardLevel ShardLevel // Regional/City/Global

	Nodes_perShard uint64
	ShardNums      uint64

	BlockSize     uint64
	BlockInterval uint64
	InjectSpeed   uint64

	// TriBFT 特有
	IsRSU         bool   // 是否为 RSU 节点（高稳定性节点）
	IsAnchor      bool   // 是否为锚点 RSU
	ParentShardID uint64 // 父分片 ID（城市或全局）

	// 初始余额
	Init_Balance *big.Int
}

// NewTriBFTConfig 创建 TriBFT 配置
func NewTriBFTConfig(nodeID, nodesPerShard, shardID, shardNums uint64) *TriBFTConfig {
	return &TriBFTConfig{
		ChainID:        shardID,
		NodeID:         nodeID,
		ShardID:        shardID,
		ShardLevel:     ShardLevelRegional,
		Nodes_perShard: nodesPerShard,
		ShardNums:      shardNums,
		BlockSize:      uint64(MaxBlockSize_global),
		BlockInterval:  uint64(Block_Interval),
		InjectSpeed:    uint64(InjectSpeed),
		IsRSU:          false, // 默认为普通节点
		IsAnchor:       false,
		ParentShardID:  0,
		Init_Balance:   Init_Balance,
	}
}
