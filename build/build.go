package build

import (
	"blockEmulator/consensus_shard/pbft_all"
	"blockEmulator/consensus_shard/tribft"
	"blockEmulator/networks"
	"blockEmulator/params"
	"blockEmulator/supervisor"
	"log"
	"time"
)

func initConfig(nid, nnm, sid, snm uint64) *params.ChainConfig {
	// Read the contents of ipTable.json
	ipMap := readIpTable("./ipTable.json")
	params.IPmap_nodeTable = ipMap
	params.SupervisorAddr = params.IPmap_nodeTable[params.SupervisorShard][0]

	// check the correctness of params
	if len(ipMap)-1 < int(snm) {
		log.Panicf("Input ShardNumber = %d, but only %d shards in ipTable.json.\n", snm, len(ipMap)-1)
	}
	for shardID := 0; shardID < len(ipMap)-1; shardID++ {
		if len(ipMap[uint64(shardID)]) < int(nnm) {
			log.Panicf("Input NodeNumber = %d, but only %d nodes in Shard %d.\n", nnm, len(ipMap[uint64(shardID)]), shardID)
		}
	}

	params.NodesInShard = int(nnm)
	params.ShardNum = int(snm)

	// init the network layer
	networks.InitNetworkTools()

	pcc := &params.ChainConfig{
		ChainID:        sid,
		NodeID:         nid,
		ShardID:        sid,
		Nodes_perShard: uint64(params.NodesInShard),
		ShardNums:      snm,
		BlockSize:      uint64(params.MaxBlockSize_global),
		BlockInterval:  uint64(params.Block_Interval),
		InjectSpeed:    uint64(params.InjectSpeed),
	}
	return pcc
}

func BuildSupervisor(nnm, snm uint64) {
	methodID := params.ConsensusMethod
	var measureMod []string
	// TriBFT (methodID=4) uses BrokerMod like CLPA_Broker and Broker
	if methodID == 0 || methodID == 2 || methodID == 4 {
		measureMod = params.MeasureBrokerMod
	} else {
		measureMod = params.MeasureRelayMod
	}
	measureMod = append(measureMod, "Tx_Details")

	lsn := new(supervisor.Supervisor)
	lsn.NewSupervisor(params.SupervisorAddr, initConfig(123, nnm, 123, snm), params.CommitteeMethod[methodID], measureMod...)
	go lsn.TcpListen()
	time.Sleep(5000 * time.Millisecond)
	lsn.SupervisorTxHandling()
}

func BuildNewPbftNode(nid, nnm, sid, snm uint64) {
	methodID := params.ConsensusMethod

	// 如果是 TriBFT，使用专门的构建函数
	if methodID == params.ConsensusTriBFT {
		BuildTriBFTNode(nid, nnm, sid, snm)
		return
	}

	// 否则使用原有的 PBFT
	worker := pbft_all.NewPbftNode(sid, nid, initConfig(nid, nnm, sid, snm), params.CommitteeMethod[methodID])
	go worker.TcpListen()
	worker.Propose()
}

// BuildTriBFTNode 构建 TriBFT 节点
func BuildTriBFTNode(nid, nnm, sid, snm uint64) {
	// 读取 IP 表
	ipMap := readIpTable("./ipTable.json")
	params.IPmap_nodeTable = ipMap
	params.SupervisorAddr = params.IPmap_nodeTable[params.SupervisorShard][0]

	// 检查参数正确性
	if len(ipMap)-1 < int(snm) {
		log.Panicf("Input ShardNumber = %d, but only %d shards in ipTable.json.\n", snm, len(ipMap)-1)
	}
	for shardID := 0; shardID < len(ipMap)-1; shardID++ {
		if len(ipMap[uint64(shardID)]) < int(nnm) {
			log.Panicf("Input NodeNumber = %d, but only %d nodes in Shard %d.\n", nnm, len(ipMap[uint64(shardID)]), shardID)
		}
	}

	params.NodesInShard = int(nnm)
	params.ShardNum = int(snm)

	// 初始化网络层
	networks.InitNetworkTools()

	// 创建 TriBFT 配置
	config := params.NewTriBFTConfig(nid, nnm, sid, snm)

	// 创建 TriBFT 节点
	node := tribft.NewTriBFTNode(config)

	// 启动 TCP 监听
	go node.TcpListen()

	// 等待其他节点启动
	time.Sleep(5 * time.Second)

	// 启动共识
	node.StartConsensus()
}
