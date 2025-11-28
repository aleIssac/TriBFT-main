package params

import "math/big"

// TriBFT consensus method ID
const ConsensusTriBFT = 4

// TriBFT hierarchical architecture configuration
var (
	// Shard configuration
	RegionalShardNum = 16 // Number of regional shards
	CityClusterNum   = 4  // Number of city clusters (virtual layer)

	// Dynamic sharding thresholds
	ShardSplitThreshold = 100 // Trigger split when node count exceeds this value
	ShardMergeThreshold = 20  // Trigger merge when node count falls below this value

	// VRM reputation configuration
	InitialReputation     = 0.05  // Initial reputation score
	ReputationDecayLambda = 0.995 // Decay coefficient lambda
	TrustedThreshold      = 0.8   // Trusted level threshold
	CandidateThreshold    = 0.2   // Candidate level threshold
	WeightDecayRate       = 0.01  // Dynamic weight decay rate alpha

	// HotStuff configuration
	HotStuffViewTimeout = 6000 // View timeout (ms)
	HotStuffDelta       = 150  // Network delay upper bound delta (ms)
	QuorumRatio         = 0.67 // Quorum ratio (2f+1)/N

	// Backup node configuration
	BackupNodeRatio    = 0.3  // Backup node ratio
	BackupSyncInterval = 1000 // Backup node sync interval (ms)

	// Credential aggregation period
	CredentialAggregationPeriod = 10 // Aggregate credentials every N blocks

	// Inter-layer communication delay configuration (milliseconds)
	IntraShardBaseDelay      = 30  // Intra-shard base delay
	IntraShardJitterRange    = 20  // Intra-shard jitter range
	ShardToCityBaseDelay     = 50  // Shard to city base delay
	ShardToCityJitterRange   = 40  // Shard to city jitter range
	CityToGlobalBaseDelay    = 80  // City to global base delay
	CityToGlobalJitterRange  = 60  // City to global jitter range
	GlobalToShardBaseDelay   = 100 // Global to shard base delay
	GlobalToShardJitterRange = 80  // Global to shard jitter range

	// Enable layer delay logging
	EnableLayerDelayLogging = false
)

// ShardLevel Shard level
type ShardLevel uint8

const (
	ShardLevelRegional ShardLevel = iota // Regional shard
	ShardLevelCity                       // City cluster
	ShardLevelGlobal                     // Global shard
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

// LayerCommunicationType Inter-layer communication type
type LayerCommunicationType uint8

const (
	CommIntraShard    LayerCommunicationType = iota // Intra-shard communication
	CommShardToCity                                 // Shard to city
	CommCityToGlobal                                // City to global
	CommGlobalToShard                               // Global to shard
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

// TriBFTConfig TriBFT chain configuration
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

	// TriBFT specific
	IsRSU         bool   // Whether this is an RSU node (high stability node)
	IsAnchor      bool   // Whether this is an anchor RSU
	ParentShardID uint64 // Parent shard ID (city or global)

	// Initial balance
	Init_Balance *big.Int
}

// NewTriBFTConfig creates a TriBFT configuration
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
		IsRSU:          false, // Default is normal node
		IsAnchor:       false,
		ParentShardID:  0,
		Init_Balance:   Init_Balance,
	}
}
