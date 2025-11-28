package tribft

import (
	"blockEmulator/core"
	"blockEmulator/message"
	"crypto/sha256"
	"encoding/json"
	"sync"
)

// HotStuffPhase HotStuff consensus phases
type HotStuffPhase uint32

const (
	PhaseIdle      HotStuffPhase = iota // Idle state
	PhasePrepare                        // Prepare phase - Leader proposes, Followers vote
	PhasePreCommit                      // PreCommit phase - Collect votes to form QC
	PhaseCommit                         // Commit phase - Enter commit state
	PhaseDecide                         // Decide phase - Final commit
)

// BlockNode node in the block tree (for three-chain rule)
type BlockNode struct {
	Block       *core.Block                // Block data
	QC          *message.QuorumCertificate // QC for this block
	ParentHash  []byte                     // Parent block hash
	Height      uint64                     // Block height
	IsCommitted bool                       // Whether committed
}

// VoteCollector vote collector, collects votes for a specific block
type VoteCollector struct {
	BlockHash   []byte                  // Block hash
	BlockHeight uint64                  // Block height
	ViewNumber  uint64                  // View number
	Votes       []*message.HotStuffVote // Collected votes
	Signatures  [][]byte                // Collected signatures
	VoterIDs    []uint64                // Voter IDs
	VoterSet    map[uint64]bool         // Prevent duplicate votes
	mu          sync.Mutex              // Lock
}

// NewVoteCollector creates a new vote collector
func NewVoteCollector(blockHash []byte, height, view uint64) *VoteCollector {
	return &VoteCollector{
		BlockHash:   blockHash,
		BlockHeight: height,
		ViewNumber:  view,
		Votes:       make([]*message.HotStuffVote, 0),
		Signatures:  make([][]byte, 0),
		VoterIDs:    make([]uint64, 0),
		VoterSet:    make(map[uint64]bool),
	}
}

// AddVote adds a vote
func (vc *VoteCollector) AddVote(vote *message.HotStuffVote) bool {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	// Check for duplicate votes
	if vc.VoterSet[vote.NodeID] {
		return false
	}

	vc.Votes = append(vc.Votes, vote)
	vc.Signatures = append(vc.Signatures, vote.Signature)
	vc.VoterIDs = append(vc.VoterIDs, vote.NodeID)
	vc.VoterSet[vote.NodeID] = true
	return true
}

// VoteCount returns the number of votes
func (vc *VoteCollector) VoteCount() int {
	vc.mu.Lock()
	defer vc.mu.Unlock()
	return len(vc.Votes)
}

// HasQuorum checks if quorum is reached
func (vc *VoteCollector) HasQuorum(threshold int) bool {
	return vc.VoteCount() >= threshold
}

// CreateQC creates a QC (Quorum Certificate)
func (vc *VoteCollector) CreateQC() *message.QuorumCertificate {
	vc.mu.Lock()
	defer vc.mu.Unlock()

	return &message.QuorumCertificate{
		BlockHash:   vc.BlockHash,
		BlockHeight: vc.BlockHeight,
		ViewNumber:  vc.ViewNumber,
		Signatures:  vc.Signatures,
		NodeIDs:     vc.VoterIDs,
	}
}

// GetQuorumThreshold calculates the quorum threshold
func GetQuorumThreshold(totalNodes int) int {
	// BFT requirement: 2f+1, where f = (n-1)/3
	// Simplified as: need more than 2/3 of nodes to agree
	return (2*totalNodes + 2) / 3
}

// ComputeBlockHash computes the hash of a block
func ComputeBlockHash(block *core.Block) []byte {
	data, _ := json.Marshal(block)
	hash := sha256.Sum256(data)
	return hash[:]
}

// PhaseString converts phase to readable string
func (p HotStuffPhase) String() string {
	switch p {
	case PhaseIdle:
		return "Idle"
	case PhasePrepare:
		return "Prepare"
	case PhasePreCommit:
		return "PreCommit"
	case PhaseCommit:
		return "Commit"
	case PhaseDecide:
		return "Decide"
	default:
		return "Unknown"
	}
}
