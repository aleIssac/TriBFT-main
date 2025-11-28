package tribft

import (
	"blockEmulator/core"
	"blockEmulator/message"
	"blockEmulator/networks"
	"blockEmulator/shard"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"
)

// HotStuffConsensus HotStuff consensus engine
type HotStuffConsensus struct {
	node *TriBFTNode // Parent node reference

	// View management
	viewNumber     atomic.Uint64
	currentPhase   atomic.Uint32 // HotStuffPhase
	lastCommitTime atomic.Int64

	// Blockchain state
	highQC              *message.QuorumCertificate // Current highest QC
	lockedQC            *message.QuorumCertificate // Locked QC
	prepareQC           *message.QuorumCertificate // Prepare phase QC
	pendingBlock        *core.Block                // Pending block
	blockTree           map[string]*BlockNode      // Block tree (blockHash -> BlockNode)
	lastCommittedHeight uint64                     // Last committed block height (for three-chain rule)

	// Vote collection
	voteCollectors map[string]*VoteCollector // blockHash -> collector

	// Pipeline control
	proposalLock sync.Mutex  // Prevent concurrent proposals
	canPropose   atomic.Bool // Whether can continue proposing

	// Synchronization
	mu         sync.RWMutex
	phaseLock  sync.Mutex
	commitLock sync.Mutex
}

// NewHotStuffConsensus creates a HotStuff consensus engine
func NewHotStuffConsensus(node *TriBFTNode) *HotStuffConsensus {
	hs := &HotStuffConsensus{
		node:                node,
		blockTree:           make(map[string]*BlockNode),
		voteCollectors:      make(map[string]*VoteCollector),
		lastCommittedHeight: 0,
	}

	hs.viewNumber.Store(0)
	hs.currentPhase.Store(uint32(PhaseIdle))
	hs.lastCommitTime.Store(time.Now().UnixMilli())
	hs.canPropose.Store(true)

	return hs
}

// Start starts the HotStuff consensus
func (hs *HotStuffConsensus) Start() {
	// TODO: Core implementation hidden - will be released after project completion
	panic("not implemented - code hidden for academic review")
}

// proposeLoop proposal loop (pipeline architecture: Leader continuously proposes)
func (hs *HotStuffConsensus) proposeLoop() {
	// TODO: Core implementation hidden - will be released after project completion
	panic("not implemented - code hidden for academic review")
}

// proposeNewBlock proposes a new block (pipeline: no waiting, continuous proposals)
func (hs *HotStuffConsensus) proposeNewBlock() {
	// TODO: Core implementation hidden - will be released after project completion
	panic("not implemented - code hidden for academic review")
}

// HandleProposal handles proposal messages
func (hs *HotStuffConsensus) HandleProposal(proposal *message.HotStuffProposal) {
	// TODO: Core implementation hidden - will be released after project completion
	panic("not implemented - code hidden for academic review")
}

// HandleVote handles vote messages
func (hs *HotStuffConsensus) HandleVote(vote *message.HotStuffVote) {
	// TODO: Core implementation hidden - will be released after project completion
	panic("not implemented - code hidden for academic review")
}

// tryCommitBlock tries to commit a block (HotStuff three-chain rule)
func (hs *HotStuffConsensus) tryCommitBlock(qc *message.QuorumCertificate) {
	// TODO: Core implementation hidden - will be released after project completion
	panic("not implemented - code hidden for academic review")
}

// commitBlock commits a block to the blockchain
func (hs *HotStuffConsensus) commitBlock(block *core.Block) {
	// TODO: Core implementation hidden - will be released after project completion
	panic("not implemented - code hidden for academic review")
}

// sendEmptyBlockSignal sends an empty block signal to Supervisor (indicates no more transactions)
func (hs *HotStuffConsensus) sendEmptyBlockSignal() {
	// TODO: Core implementation hidden - will be released after project completion
	panic("not implemented - code hidden for academic review")
}

// sendBlockInfoToSupervisor sends block information to Supervisor
func (hs *HotStuffConsensus) sendBlockInfoToSupervisor(block *core.Block, commitTime time.Time) {
	// TODO: Core implementation hidden - will be released after project completion
	panic("not implemented - code hidden for academic review")
}

// extractAndUpdateReputation extracts behavior credentials and updates reputation
func (hs *HotStuffConsensus) extractAndUpdateReputation(block *core.Block) {
	// TODO: Core implementation hidden - will be released after project completion
	panic("not implemented - code hidden for academic review")
}

// verifyProposal verifies a proposal
func (hs *HotStuffConsensus) verifyProposal(proposal *message.HotStuffProposal) bool {
	// TODO: Core implementation hidden - will be released after project completion
	panic("not implemented - code hidden for academic review")
}

// isSafeToVote checks if it's safe to vote
func (hs *HotStuffConsensus) isSafeToVote(proposal *message.HotStuffProposal) bool {
	// TODO: Core implementation hidden - will be released after project completion
	panic("not implemented - code hidden for academic review")
}

// createVote creates a vote
func (hs *HotStuffConsensus) createVote(block *core.Block, viewNumber uint64) *message.HotStuffVote {
	// TODO: Core implementation hidden - will be released after project completion
	panic("not implemented - code hidden for academic review")
}

// sendVote sends a vote to the leader
func (hs *HotStuffConsensus) sendVote(vote *message.HotStuffVote, leader *shard.Node) {
	// TODO: Core implementation hidden - will be released after project completion
	panic("not implemented - code hidden for academic review")
}

// broadcastProposal broadcasts a proposal
func (hs *HotStuffConsensus) broadcastProposal(proposal *message.HotStuffProposal) {
	// TODO: Core implementation hidden - will be released after project completion
	panic("not implemented - code hidden for academic review")
}

// broadcastQC broadcasts QC to all nodes (for view synchronization)
func (hs *HotStuffConsensus) broadcastQC(qc *message.QuorumCertificate) {
	// TODO: Core implementation hidden - will be released after project completion
	panic("not implemented - code hidden for academic review")
}

// HandleQCSync handles QC synchronization (Follower nodes receive Leader's QC)
func (hs *HotStuffConsensus) HandleQCSync(qc *message.QuorumCertificate) {
	// TODO: Core implementation hidden - will be released after project completion
	panic("not implemented - code hidden for academic review")
}

// isLeader checks if the current node is the leader
func (hs *HotStuffConsensus) isLeader() bool {
	return hs.node.nodeID == hs.getLeaderID()
}

// getLeaderID gets the leader ID of the current view
func (hs *HotStuffConsensus) getLeaderID() uint64 {
	// Simple rotation: view % totalNodes
	currentView := hs.viewNumber.Load()
	nodesPerShard := hs.node.config.Nodes_perShard
	leaderID := currentView % nodesPerShard
	return leaderID
}

// watchViewTimeout monitors view timeout
func (hs *HotStuffConsensus) watchViewTimeout() {
	// TODO: Core implementation hidden - will be released after project completion
	panic("not implemented - code hidden for academic review")
}

// triggerViewChange triggers a view change
func (hs *HotStuffConsensus) triggerViewChange() {
	// TODO: Core implementation hidden - will be released after project completion
	panic("not implemented - code hidden for academic review")
}

// compareQC compares the heights of two QCs
func (hs *HotStuffConsensus) compareQC(qc1, qc2 *message.QuorumCertificate) int {
	if qc1 == nil && qc2 == nil {
		return 0
	}
	if qc1 == nil {
		return -1
	}
	if qc2 == nil {
		return 1
	}

	if qc1.BlockHeight > qc2.BlockHeight {
		return 1
	} else if qc1.BlockHeight < qc2.BlockHeight {
		return -1
	}
	return 0
}

// signBlock signs a block (simplified implementation)
func (hs *HotStuffConsensus) signBlock(block *core.Block) []byte {
	// In practice, should use the node's private key for signing
	// Here we return the block hash as simplification
	return ComputeBlockHash(block)
}

// getCurrentPhase gets the current phase
func (hs *HotStuffConsensus) getCurrentPhase() HotStuffPhase {
	return HotStuffPhase(hs.currentPhase.Load())
}

// setPhase sets the phase
func (hs *HotStuffConsensus) setPhase(phase HotStuffPhase) {
	hs.currentPhase.Store(uint32(phase))
}

// getBlockByHash gets a block node by hash (for three-chain rule)
func (hs *HotStuffConsensus) getBlockByHash(blockHash []byte) *BlockNode {
	blockHashStr := string(blockHash)
	blockNode, exists := hs.blockTree[blockHashStr]
	if !exists {
		return nil
	}
	return blockNode
}

// Unused imports to prevent compile errors
var (
	_ = networks.TcpDial
	_ = json.Marshal
)
