package message

import (
	"blockEmulator/core"
	"blockEmulator/shard"
)

// TriBFT 消息类型
const (
	// HotStuff 共识消息
	CHSProposal MessageType = "hs_proposal"
	CHSVote     MessageType = "hs_vote"
	CHSNewView  MessageType = "hs_newview"
	CHSTimeout  MessageType = "hs_timeout"

	// 层间通信消息
	CCitySummary      MessageType = "city_summary"      // 城市摘要
	CReputationUpdate MessageType = "reputation_update" // 信誉更新广播
)

// QuorumCertificate (QC) 法定人数证书
// HotStuff 的核心数据结构，证明有足够多的节点对某个提案投票
type QuorumCertificate struct {
	ViewNumber   uint64   // 视图编号
	BlockHash    []byte   // 区块哈希
	BlockHeight  uint64   // 区块高度
	AggSignature []byte   // 聚合签名（BLS 或简化为多个签名的集合）
	SignerIDs    []uint64 // 签名者 ID 列表
	SignerCount  int      // 签名者数量
}

// HotStuffProposal HotStuff 提案消息
type HotStuffProposal struct {
	Block        *core.Block        // 提案的区块
	HighQC       *QuorumCertificate // 对上一个区块的 QC
	ViewNumber   uint64             // 当前视图编号
	ProposerNode *shard.Node        // 提案者节点信息
	Signature    []byte             // 提案者签名
}

// HotStuffVote HotStuff 投票消息
type HotStuffVote struct {
	BlockHash   []byte      // 投票的区块哈希
	BlockHeight uint64      // 区块高度
	ViewNumber  uint64      // 视图编号
	VoterNode   *shard.Node // 投票者节点信息
	PartialSig  []byte      // 部分签名（用于聚合）
}

// HotStuffNewView HotStuff 新视图消息
// 当视图变更时，节点发送此消息
type HotStuffNewView struct {
	ViewNumber uint64             // 新视图编号
	HighestQC  *QuorumCertificate // 节点当前持有的最高 QC
	SenderNode *shard.Node        // 发送者节点信息
	Signature  []byte             // 签名
}

// HotStuffTimeout 超时消息
type HotStuffTimeout struct {
	ViewNumber uint64             // 超时的视图编号
	SenderNode *shard.Node        // 发送者节点信息
	HighestQC  *QuorumCertificate // 当前持有的最高 QC
}

// CitySummary 城市层摘要
// 区域分片向虚拟城市层提交的凭证批次摘要
type CitySummary struct {
	ShardID           uint64             // 分片 ID
	CityID            uint64             // 城市 ID（虚拟层）
	EpochID           uint64             // 周期 ID
	CredentialCount   int                // 凭证数量
	BatchHash         []byte             // 凭证批次哈希
	ReputationUpdates map[uint64]float64 // nodeID -> 信誉变化
	BlockHeight       uint64             // 区块高度
	Signature         []byte             // 签名
}

// ReputationUpdate 全局信誉更新消息
// 虚拟全局层广播给所有节点的信誉更新
type ReputationUpdate struct {
	EpochID           uint64             // 周期 ID
	GlobalReputations map[uint64]float64 // nodeID -> 全局信誉分
	Timestamp         int64              // 时间戳
	Signature         []byte             // 签名
}
