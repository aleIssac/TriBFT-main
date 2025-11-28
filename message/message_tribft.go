package message

import (
	"blockEmulator/core"
	"blockEmulator/shard"
)

// TriBFT Message Types
const (
	// HotStuff consensus messages
	CHSProposal MessageType = "hs_proposal"
	CHSVote     MessageType = "hs_vote"
	CHSNewView  MessageType = "hs_newview"
	CHSTimeout  MessageType = "hs_timeout"

	// Inter-layer communication messages
	CCitySummary      MessageType = "city_summary"      // City summary
	CReputationUpdate MessageType = "reputation_update" // Reputation update broadcast
)

// QuorumCertificate (QC) Quorum Certificate
// Core data structure of HotStuff, proving that enough nodes voted for a proposal
type QuorumCertificate struct {
	ViewNumber   uint64   // View number
	BlockHash    []byte   // Block hash
	BlockHeight  uint64   // Block height
	AggSignature []byte   // Aggregated signature (BLS or simplified as a collection of multiple signatures)
	SignerIDs    []uint64 // Signer ID list
	SignerCount  int      // Number of signers
	Signatures   [][]byte // Individual signatures (for aggregation)
	NodeIDs      []uint64 // Node IDs that voted
}

// HotStuffProposal HotStuff proposal message
type HotStuffProposal struct {
	Block        *core.Block        // Proposed block
	HighQC       *QuorumCertificate // QC for the previous block
	ViewNumber   uint64             // Current view number
	ProposerNode *shard.Node        // Proposer node information
	Signature    []byte             // Proposer signature
}

// HotStuffVote HotStuff vote message
type HotStuffVote struct {
	BlockHash   []byte      // Voted block hash
	BlockHeight uint64      // Block height
	ViewNumber  uint64      // View number
	VoterNode   *shard.Node // Voter node information
	PartialSig  []byte      // Partial signature (for aggregation)
	NodeID      uint64      // Voter node ID
	Signature   []byte      // Vote signature
}

// HotStuffNewView HotStuff new view message
// Sent when view changes
type HotStuffNewView struct {
	ViewNumber uint64             // New view number
	HighestQC  *QuorumCertificate // Highest QC held by the node
	SenderNode *shard.Node        // Sender node information
	Signature  []byte             // Signature
}

// HotStuffTimeout Timeout message
type HotStuffTimeout struct {
	ViewNumber uint64             // Timed out view number
	SenderNode *shard.Node        // Sender node information
	HighestQC  *QuorumCertificate // Current highest QC held
}

// CitySummary City layer summary
// Credential batch summary submitted from regional shard to virtual city layer
type CitySummary struct {
	ShardID           uint64             // Shard ID
	CityID            uint64             // City ID (virtual layer)
	EpochID           uint64             // Epoch ID
	CredentialCount   int                // Number of credentials
	BatchHash         []byte             // Credential batch hash
	ReputationUpdates map[uint64]float64 // nodeID -> reputation change
	BlockHeight       uint64             // Block height
	Signature         []byte             // Signature
}

// ReputationUpdate Global reputation update message
// Reputation update broadcast from virtual global layer to all nodes
type ReputationUpdate struct {
	EpochID           uint64             // Epoch ID
	GlobalReputations map[uint64]float64 // nodeID -> global reputation score
	Timestamp         int64              // Timestamp
	Signature         []byte             // Signature
}

// CityAggregatedSummary City layer aggregated summary
type CityAggregatedSummary struct {
	ShardID           uint64             // Shard ID
	CityID            uint64             // City ID
	EpochID           uint64             // Epoch ID
	TxCount           int                // Transaction count
	CrossShardTxCount int                // Cross-shard transaction count
	CredentialBatch   []byte             // Aggregated credential batch
	ReputationDeltas  map[uint64]float64 // nodeID -> reputation delta
	BlockHeight       uint64             // Block height
	Signature         []byte             // Signature
}

// DistrictReputationReport District layer reputation report
type DistrictReputationReport struct {
	ShardID          uint64             // Shard ID
	DistrictID       uint64             // District ID
	EpochID          uint64             // Epoch ID
	ReputationScores map[uint64]float64 // nodeID -> reputation score
	AggregatedCount  int                // Number of aggregated city reports
	BlockHeight      uint64             // Block height
	Signature        []byte             // Signature
}
