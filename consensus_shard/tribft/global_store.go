package tribft

import (
	"blockEmulator/message"
	"blockEmulator/networks"
	"blockEmulator/reputation/vrm"
	"encoding/json"
	"sync"
)

// GlobalStore Global layer reputation storage
// Responsible for:
// 1. Maintaining global reputation state
// 2. Receiving reputation reports from all District layers
// 3. Providing reputation query services
type GlobalStore struct {
	node *TriBFTNode

	// Reputation storage
	globalReputation map[string]*vrm.ReputationScore // address -> reputation score
	reputationLock   sync.RWMutex

	// District reports aggregation
	districtReports map[uint64]*message.DistrictReputationReport // shardID -> latest report
	reportLock      sync.RWMutex
}

// NewGlobalStore creates a Global layer storage
func NewGlobalStore(node *TriBFTNode) *GlobalStore {
	return &GlobalStore{
		node:             node,
		globalReputation: make(map[string]*vrm.ReputationScore),
		districtReports:  make(map[uint64]*message.DistrictReputationReport),
	}
}

// HandleDistrictReport handles reputation reports from District layer
func (gs *GlobalStore) HandleDistrictReport(report *message.DistrictReputationReport) {
	// TODO: Core implementation hidden - will be released after project completion
	panic("not implemented - code hidden for academic review")
}

// AggregateReputationUpdates aggregates reputation updates from all shards
func (gs *GlobalStore) AggregateReputationUpdates() map[string]*vrm.ReputationScore {
	// TODO: Core implementation hidden - will be released after project completion
	panic("not implemented - code hidden for academic review")
}

// BroadcastGlobalReputation broadcasts global reputation state
func (gs *GlobalStore) BroadcastGlobalReputation() {
	// TODO: Core implementation hidden - will be released after project completion
	panic("not implemented - code hidden for academic review")
}

// GetReputation gets the reputation score of a specific address
func (gs *GlobalStore) GetReputation(address string) *vrm.ReputationScore {
	gs.reputationLock.RLock()
	defer gs.reputationLock.RUnlock()

	if score, exists := gs.globalReputation[address]; exists {
		return score
	}
	return vrm.DefaultReputationScore()
}

// UpdateReputation updates the reputation of a specific address
func (gs *GlobalStore) UpdateReputation(address string, score *vrm.ReputationScore) {
	gs.reputationLock.Lock()
	defer gs.reputationLock.Unlock()

	gs.globalReputation[address] = score
}

// BatchUpdateReputation batch updates reputations
func (gs *GlobalStore) BatchUpdateReputation(updates map[string]*vrm.ReputationScore) {
	gs.reputationLock.Lock()
	defer gs.reputationLock.Unlock()

	for addr, score := range updates {
		gs.globalReputation[addr] = score
	}
}

// sendToGlobalLeader sends aggregated data to Global layer leader
func (gs *GlobalStore) sendToGlobalLeader(data interface{}, msgType message.MessageType) {
	// TODO: Core implementation hidden - will be released after project completion
	panic("not implemented - code hidden for academic review")
}

// IsGlobalLeader checks if current node is the Global layer leader
func (gs *GlobalStore) IsGlobalLeader() bool {
	// Default: Node 0 of Shard 0 is the Global leader
	return gs.node.shardID == 0 && gs.node.nodeID == 0
}

// GetAllReputations gets all reputation data (for debugging/query)
func (gs *GlobalStore) GetAllReputations() map[string]*vrm.ReputationScore {
	gs.reputationLock.RLock()
	defer gs.reputationLock.RUnlock()

	result := make(map[string]*vrm.ReputationScore)
	for k, v := range gs.globalReputation {
		result[k] = v
	}
	return result
}

// Unused imports to prevent compile errors
var (
	_ = networks.TcpDial
	_ = json.Marshal
)
