package tribft

import (
	"blockEmulator/chain"
	"blockEmulator/params"
	"blockEmulator/reputation/vrm"
	"blockEmulator/shard"
	"bufio"
	"io"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/ethdb"
)

// Type aliases for VRM types (exported for use in other packages)
type (
	ReputationScore = vrm.ReputationScore
	BehaviorType    = vrm.BehaviorType
)

const (
	BehaviorMaliciousConsensus = vrm.BehaviorMaliciousConsensus
	BehaviorFalseData          = vrm.BehaviorFalseData
	BehaviorUnexpectedOffline  = vrm.BehaviorUnexpectedOffline
	BehaviorConsensusSuccess   = vrm.BehaviorConsensusSuccess
	BehaviorDataVerified       = vrm.BehaviorDataVerified
	BehaviorVerifyOthers       = vrm.BehaviorVerifyOthers
)

// TriBFTNode is the TriBFT consensus node
// Integrates three-layer architecture, VRM reputation mechanism and HotStuff consensus
type TriBFTNode struct {
	// Basic configuration
	nodeID      uint64
	shardID     uint64
	runningNode *shard.Node
	config      *params.TriBFTConfig

	// === Regional Shard Layer (actual execution) ===
	hotstuffEngine *HotStuffConsensus // HotStuff consensus engine
	blockchain     *chain.BlockChain  // Blockchain
	db             ethdb.Database     // State database

	// VRM Reputation Management
	vrmManager    *vrm.ReputationManager // Reputation manager
	credentialMgr *vrm.CredentialManager // Credential manager

	// === City Cluster Layer (embedded simulation) ===
	cityAggregator *CityAggregator // City aggregator

	// === Global Shard Layer (embedded simulation) ===
	globalRepStore *GlobalStore // Global reputation store

	// Network
	ipNodeTable map[uint64]map[uint64]string
	tcpln       net.Listener
	stopSignal  atomic.Bool
	tcpPoolLock sync.Mutex

	// Delay simulation
	delaySimulator *LayerDelaySimulator

	// Logger
	logger *log.Logger
}

// NewTriBFTNode creates a new TriBFT node
func NewTriBFTNode(config *params.TriBFTConfig) *TriBFTNode {
	node := &TriBFTNode{
		nodeID:      config.NodeID,
		shardID:     config.ShardID,
		config:      config,
		ipNodeTable: params.IPmap_nodeTable,
	}

	// Set node information
	node.runningNode = &shard.Node{
		NodeID:  config.NodeID,
		ShardID: config.ShardID,
		IPaddr:  params.IPmap_nodeTable[config.ShardID][config.NodeID],
	}

	// Initialize database (create directory first)
	dbDir := params.DatabaseWrite_path + "mptDB/ldb/s" + strconv.FormatUint(config.ShardID, 10) +
		"/n" + strconv.FormatUint(config.NodeID, 10)
	err := os.MkdirAll(dbDir, os.ModePerm)
	if err != nil {
		log.Panic("cannot create database directory:", err)
	}

	node.db, err = rawdb.NewLevelDBDatabase(dbDir, 0, 1, "accountState", false)
	if err != nil {
		log.Panic(err)
	}

	// Initialize blockchain
	chainConfig := &params.ChainConfig{
		ChainID:        config.ChainID,
		NodeID:         config.NodeID,
		ShardID:        config.ShardID,
		Nodes_perShard: config.Nodes_perShard,
		ShardNums:      config.ShardNums,
		BlockSize:      config.BlockSize,
		BlockInterval:  config.BlockInterval,
		InjectSpeed:    config.InjectSpeed,
	}
	node.blockchain, err = chain.NewBlockChain(chainConfig, node.db)
	if err != nil {
		log.Panic("cannot create blockchain:", err)
	}

	// Initialize VRM reputation management
	node.vrmManager = vrm.NewReputationManager(nil)
	node.credentialMgr = vrm.NewCredentialManager()

	// Initialize delay simulator
	node.delaySimulator = NewLayerDelaySimulator()

	// Initialize virtual upper layer components
	node.cityAggregator = NewCityAggregator(node)
	node.globalRepStore = NewGlobalStore(node)

	// Initialize HotStuff consensus engine
	node.hotstuffEngine = NewHotStuffConsensus(node)

	// Initialize logger (create directory first)
	logDir := params.LogWrite_path
	err = os.MkdirAll(logDir, os.ModePerm)
	if err != nil {
		log.Panic("cannot create log directory:", err)
	}

	logPath := logDir + "/s" + strconv.FormatUint(config.ShardID, 10) +
		"n" + strconv.FormatUint(config.NodeID, 10) + ".log"
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Panic("cannot create log file:", err)
	}
	node.logger = log.New(logFile, "", log.Ldate|log.Ltime|log.Lshortfile)

	node.stopSignal.Store(false)

	return node
}

// TcpListen starts TCP listener for the node
func (tn *TriBFTNode) TcpListen() {
	ln, err := net.Listen("tcp", tn.runningNode.IPaddr)
	tn.tcpln = ln
	if err != nil {
		log.Panic(err)
	}

	tn.logger.Printf("S%dN%d: TCP listening on %s\n",
		tn.shardID, tn.nodeID, tn.runningNode.IPaddr)

	for {
		conn, err := tn.tcpln.Accept()
		if err != nil {
			return
		}
		go tn.handleClientRequest(conn)
	}
}

// handleClientRequest handles client requests
func (tn *TriBFTNode) handleClientRequest(conn net.Conn) {
	defer conn.Close()
	clientReader := bufio.NewReader(conn)

	for {
		clientRequest, err := clientReader.ReadBytes('\n')
		if tn.stopSignal.Load() {
			return
		}

		switch err {
		case nil:
			tn.tcpPoolLock.Lock()
			tn.handleMessage(clientRequest)
			tn.tcpPoolLock.Unlock()
		case io.EOF:
			tn.logger.Println("Client closed connection")
			return
		default:
			tn.logger.Printf("Error: %v\n", err)
			return
		}
	}
}

// handleMessage handles incoming messages
func (tn *TriBFTNode) handleMessage(msg []byte) {
	// TODO: Core implementation hidden - will be released after project completion
	// This function handles message routing for:
	// - HotStuff consensus messages (Proposal, Vote, NewView)
	// - Inter-layer communication (CitySummary, ReputationUpdate)
	// - Transaction injection
	// - Stop signals
	panic("not implemented - code hidden for academic review")
}

// handleHotStuffProposal handles HotStuff proposals
func (tn *TriBFTNode) handleHotStuffProposal(content []byte) {
	// TODO: Core implementation hidden - will be released after project completion
	// This function unmarshals and processes HotStuff proposals
	panic("not implemented - code hidden for academic review")
}

// handleHotStuffVote handles HotStuff votes
func (tn *TriBFTNode) handleHotStuffVote(content []byte) {
	// TODO: Core implementation hidden - will be released after project completion
	// This function unmarshals and processes HotStuff votes
	panic("not implemented - code hidden for academic review")
}

// handleHotStuffNewView handles HotStuff new view messages
func (tn *TriBFTNode) handleHotStuffNewView(content []byte) {
	// TODO: Core implementation hidden - will be released after project completion
	// This function handles view synchronization with QC
	panic("not implemented - code hidden for academic review")
}

// handleCitySummary handles city summary messages
func (tn *TriBFTNode) handleCitySummary(content []byte) {
	// TODO: Core implementation hidden - will be released after project completion
	// This function processes city-level aggregation summaries
	panic("not implemented - code hidden for academic review")
}

// handleReputationUpdate handles reputation update messages
func (tn *TriBFTNode) handleReputationUpdate(content []byte) {
	// TODO: Core implementation hidden - will be released after project completion
	// This function processes global reputation updates from VRM
	panic("not implemented - code hidden for academic review")
}

// handleInjectTx handles transaction injection
func (tn *TriBFTNode) handleInjectTx(content []byte) {
	// TODO: Core implementation hidden - will be released after project completion
	// This function handles transaction injection and pipeline resumption
	panic("not implemented - code hidden for academic review")
}

// handleStop handles stop signal
func (tn *TriBFTNode) handleStop() {
	// TODO: Core implementation hidden - will be released after project completion
	// This function handles graceful shutdown
	panic("not implemented - code hidden for academic review")
}

// StartConsensus starts the consensus process
func (tn *TriBFTNode) StartConsensus() {
	// TODO: Core implementation hidden - will be released after project completion
	// This function starts the HotStuff consensus engine
	panic("not implemented - code hidden for academic review")
}

// getNeighborNodes returns the addresses of neighbor nodes
func (tn *TriBFTNode) getNeighborNodes() []string {
	neighbors := make([]string, 0)
	for nid := uint64(0); nid < tn.config.Nodes_perShard; nid++ {
		if nid != tn.nodeID {
			neighbors = append(neighbors, tn.ipNodeTable[tn.shardID][nid])
		}
	}
	return neighbors
}
