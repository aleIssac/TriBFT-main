package vrm

import (
	"crypto/sha256"
	"encoding/binary"
	"sync"
)

// BehaviorCredential behavior credential
// Records node behavior events for reputation updates
type BehaviorCredential struct {
	NodeID       uint64       // Node ID
	ShardID      uint64       // Shard ID
	BehaviorType BehaviorType // Behavior type
	Timestamp    int64        // Timestamp
	BlockHeight  uint64       // Block height where behavior occurred
	Evidence     []byte       // Behavior evidence (e.g., transaction hash, signature)
	Signature    []byte       // Credential signature
}

// CredentialBatch credential batch
// Aggregation of multiple credentials for submission to upper layer
type CredentialBatch struct {
	ShardID     uint64                // Shard ID
	EpochID     uint64                // Epoch ID
	Credentials []*BehaviorCredential // Credential list
	BatchHash   []byte                // Batch hash
	BlockHeight uint64                // Block height at aggregation
}

// CredentialManager credential manager
type CredentialManager struct {
	mu sync.RWMutex

	// Pending credentials pool (organized by shard)
	pendingCredentials map[uint64][]*BehaviorCredential

	// Aggregated batches
	aggregatedBatches map[uint64][]*CredentialBatch
}

// NewCredentialManager creates a credential manager
func NewCredentialManager() *CredentialManager {
	return &CredentialManager{
		pendingCredentials: make(map[uint64][]*BehaviorCredential),
		aggregatedBatches:  make(map[uint64][]*CredentialBatch),
	}
}

// AddCredential adds a behavior credential
func (cm *CredentialManager) AddCredential(cred *BehaviorCredential) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.pendingCredentials[cred.ShardID] = append(
		cm.pendingCredentials[cred.ShardID],
		cred,
	)
}

// GetPendingCredentials gets pending credentials for aggregation
func (cm *CredentialManager) GetPendingCredentials(shardID uint64) []*BehaviorCredential {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.pendingCredentials[shardID]
}

// AggregatePendingCredentials aggregates pending credentials
func (cm *CredentialManager) AggregatePendingCredentials(
	shardID, blockHeight uint64,
) *CredentialBatch {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	creds := cm.pendingCredentials[shardID]
	if len(creds) == 0 {
		return nil
	}

	// Create batch
	batch := &CredentialBatch{
		ShardID:     shardID,
		EpochID:     blockHeight / 10, // One epoch per 10 blocks
		Credentials: creds,
		BlockHeight: blockHeight,
	}

	// Compute batch hash
	batch.BatchHash = cm.computeBatchHash(batch)

	// Store aggregated batch
	cm.aggregatedBatches[shardID] = append(
		cm.aggregatedBatches[shardID],
		batch,
	)

	// Clear pending pool
	cm.pendingCredentials[shardID] = nil

	return batch
}

// GetAggregatedBatches gets aggregated batches
func (cm *CredentialManager) GetAggregatedBatches(shardID uint64) []*CredentialBatch {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.aggregatedBatches[shardID]
}

// ClearAggregatedBatches clears aggregated batches
func (cm *CredentialManager) ClearAggregatedBatches(shardID uint64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.aggregatedBatches[shardID] = nil
}

// computeBatchHash computes batch hash
func (cm *CredentialManager) computeBatchHash(batch *CredentialBatch) []byte {
	hasher := sha256.New()

	// ShardID
	binary.Write(hasher, binary.BigEndian, batch.ShardID)
	// EpochID
	binary.Write(hasher, binary.BigEndian, batch.EpochID)
	// BlockHeight
	binary.Write(hasher, binary.BigEndian, batch.BlockHeight)

	// Key fields from each credential
	for _, cred := range batch.Credentials {
		binary.Write(hasher, binary.BigEndian, cred.NodeID)
		binary.Write(hasher, binary.BigEndian, uint8(cred.BehaviorType))
		binary.Write(hasher, binary.BigEndian, cred.Timestamp)
		hasher.Write(cred.Evidence)
	}

	return hasher.Sum(nil)
}

// VerifyBatchHash verifies batch hash
func (cm *CredentialManager) VerifyBatchHash(batch *CredentialBatch) bool {
	expectedHash := cm.computeBatchHash(batch)

	if len(expectedHash) != len(batch.BatchHash) {
		return false
	}

	for i := range expectedHash {
		if expectedHash[i] != batch.BatchHash[i] {
			return false
		}
	}

	return true
}

// GetCredentialCount gets the count of pending credentials
func (cm *CredentialManager) GetCredentialCount(shardID uint64) int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return len(cm.pendingCredentials[shardID])
}
