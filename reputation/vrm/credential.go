package vrm

import (
	"crypto/sha256"
	"encoding/binary"
	"sync"
)

// BehaviorCredential 行为凭证
// 记录节点的行为事件，用于信誉更新
type BehaviorCredential struct {
	NodeID       uint64       // 节点 ID
	ShardID      uint64       // 分片 ID
	BehaviorType BehaviorType // 行为类型
	Timestamp    int64        // 时间戳
	BlockHeight  uint64       // 所在区块高度
	Evidence     []byte       // 行为证据（如交易哈希、签名等）
	Signature    []byte       // 凭证签名
}

// CredentialBatch 凭证批次
// 多个凭证的聚合，用于向上层提交
type CredentialBatch struct {
	ShardID     uint64                // 分片 ID
	EpochID     uint64                // 周期 ID
	Credentials []*BehaviorCredential // 凭证列表
	BatchHash   []byte                // 批次哈希
	BlockHeight uint64                // 聚合时的区块高度
}

// CredentialManager 凭证管理器
type CredentialManager struct {
	mu sync.RWMutex

	// 待聚合的凭证池（按分片组织）
	pendingCredentials map[uint64][]*BehaviorCredential

	// 已聚合的批次
	aggregatedBatches map[uint64][]*CredentialBatch
}

// NewCredentialManager 创建凭证管理器
func NewCredentialManager() *CredentialManager {
	return &CredentialManager{
		pendingCredentials: make(map[uint64][]*BehaviorCredential),
		aggregatedBatches:  make(map[uint64][]*CredentialBatch),
	}
}

// AddCredential 添加行为凭证
func (cm *CredentialManager) AddCredential(cred *BehaviorCredential) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.pendingCredentials[cred.ShardID] = append(
		cm.pendingCredentials[cred.ShardID],
		cred,
	)
}

// GetPendingCredentials 获取待聚合的凭证
func (cm *CredentialManager) GetPendingCredentials(shardID uint64) []*BehaviorCredential {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.pendingCredentials[shardID]
}

// AggregatePendingCredentials 聚合待处理凭证
func (cm *CredentialManager) AggregatePendingCredentials(
	shardID, blockHeight uint64,
) *CredentialBatch {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	creds := cm.pendingCredentials[shardID]
	if len(creds) == 0 {
		return nil
	}

	// 创建批次
	batch := &CredentialBatch{
		ShardID:     shardID,
		EpochID:     blockHeight / 10, // 每10个区块一个epoch
		Credentials: creds,
		BlockHeight: blockHeight,
	}

	// 计算批次哈希
	batch.BatchHash = cm.computeBatchHash(batch)

	// 存储已聚合的批次
	cm.aggregatedBatches[shardID] = append(
		cm.aggregatedBatches[shardID],
		batch,
	)

	// 清空待聚合池
	cm.pendingCredentials[shardID] = nil

	return batch
}

// GetAggregatedBatches 获取已聚合的批次
func (cm *CredentialManager) GetAggregatedBatches(shardID uint64) []*CredentialBatch {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.aggregatedBatches[shardID]
}

// ClearAggregatedBatches 清空已聚合的批次
func (cm *CredentialManager) ClearAggregatedBatches(shardID uint64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.aggregatedBatches[shardID] = nil
}

// computeBatchHash 计算批次哈希
func (cm *CredentialManager) computeBatchHash(batch *CredentialBatch) []byte {
	hasher := sha256.New()

	// ShardID
	binary.Write(hasher, binary.BigEndian, batch.ShardID)
	// EpochID
	binary.Write(hasher, binary.BigEndian, batch.EpochID)
	// BlockHeight
	binary.Write(hasher, binary.BigEndian, batch.BlockHeight)

	// 每个凭证的关键字段
	for _, cred := range batch.Credentials {
		binary.Write(hasher, binary.BigEndian, cred.NodeID)
		binary.Write(hasher, binary.BigEndian, uint8(cred.BehaviorType))
		binary.Write(hasher, binary.BigEndian, cred.Timestamp)
		hasher.Write(cred.Evidence)
	}

	return hasher.Sum(nil)
}

// VerifyBatchHash 验证批次哈希
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

// GetCredentialCount 获取待聚合凭证数量
func (cm *CredentialManager) GetCredentialCount(shardID uint64) int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return len(cm.pendingCredentials[shardID])
}
