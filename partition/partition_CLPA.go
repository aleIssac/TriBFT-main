package partition

import (
	"blockEmulator/utils"
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"
)

// CLPAState state of constraint label propagation algorithm
type CLPAState struct {
	NetGraph          Graph          // Graph to run CLPA algorithm on
	PartitionMap      map[Vertex]int // Partition info map, which shard a vertex belongs to
	Edges2Shard       []int          // Number of edges adjacent to shard, corresponds to "total weight of edges associated with label k" in paper
	VertexsNumInShard []int          // Number of vertices in each shard
	WeightPenalty     float64        // Weight penalty, corresponds to beta in paper
	MinEdges2Shard    int            // Minimum number of shard adjacent edges
	MaxIterations     int            // Maximum iterations, constraint, corresponds to tau in paper
	CrossShardEdgeNum int            // Total number of cross-shard edges
	ShardNum          int            // Number of shards
	GraphHash         []byte
}

func (graph *CLPAState) Hash() []byte {
	hash := sha256.Sum256(graph.Encode())
	return hash[:]
}

func (graph *CLPAState) Encode() []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(graph)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

// AddVertex adds a vertex, needs to assign it to a default shard
func (cs *CLPAState) AddVertex(v Vertex) {
	cs.NetGraph.AddVertex(v)
	if val, ok := cs.PartitionMap[v]; !ok {
		cs.PartitionMap[v] = utils.Addr2Shard(v.Addr)
	} else {
		cs.PartitionMap[v] = val
	}
	cs.VertexsNumInShard[cs.PartitionMap[v]] += 1 // Can batch update VertexsNumInShard after batch processing
	// Can also skip this since CLPA will update parameters before running
}

// AddEdge adds an edge, needs to assign endpoints (if not exist) to default shards
func (cs *CLPAState) AddEdge(u, v Vertex) {
	// If vertex doesn't exist, add edge with weight fixed at 1
	if _, ok := cs.NetGraph.VertexSet[u]; !ok {
		cs.AddVertex(u)
	}
	if _, ok := cs.NetGraph.VertexSet[v]; !ok {
		cs.AddVertex(v)
	}
	cs.NetGraph.AddEdge(u, v)
	// Can batch update Edges2Shard after batch processing
	// Can also skip since CLPA will update parameters before running
}

// CopyCLPA copies CLPA state
func (dst *CLPAState) CopyCLPA(src CLPAState) {
	dst.NetGraph.CopyGraph(src.NetGraph)
	dst.PartitionMap = make(map[Vertex]int)
	for v := range src.PartitionMap {
		dst.PartitionMap[v] = src.PartitionMap[v]
	}
	dst.Edges2Shard = make([]int, src.ShardNum)
	copy(dst.Edges2Shard, src.Edges2Shard)
	dst.VertexsNumInShard = src.VertexsNumInShard
	dst.WeightPenalty = src.WeightPenalty
	dst.MinEdges2Shard = src.MinEdges2Shard
	dst.MaxIterations = src.MaxIterations
	dst.ShardNum = src.ShardNum
}

// PrintCLPA outputs CLPA state
func (cs *CLPAState) PrintCLPA() {
	cs.NetGraph.PrintGraph()
	println(cs.MinEdges2Shard)
	for v, item := range cs.PartitionMap {
		print(v.Addr, " ", item, "\t")
	}
	for _, item := range cs.Edges2Shard {
		print(item, " ")
	}
	println()
}

// ComputeEdges2Shard computes Wk (Edges2Shard) based on current partition
func (cs *CLPAState) ComputeEdges2Shard() {
	cs.Edges2Shard = make([]int, cs.ShardNum)
	interEdge := make([]int, cs.ShardNum)
	cs.MinEdges2Shard = math.MaxInt

	for idx := 0; idx < cs.ShardNum; idx++ {
		cs.Edges2Shard[idx] = 0
		interEdge[idx] = 0
	}

	for v, lst := range cs.NetGraph.EdgeSet {
		// Get the shard that vertex v belongs to
		vShard := cs.PartitionMap[v]
		for _, u := range lst {
			// Same, get the shard that vertex u belongs to
			uShard := cs.PartitionMap[u]
			if vShard != uShard {
				// If v and u are in different shards, increment corresponding Edges2Shard
				// Only count in-degree to avoid double counting
				cs.Edges2Shard[uShard] += 1
			} else {
				interEdge[uShard]++
			}
		}
	}

	cs.CrossShardEdgeNum = 0
	for _, val := range cs.Edges2Shard {
		cs.CrossShardEdgeNum += val
	}
	cs.CrossShardEdgeNum /= 2

	for idx := 0; idx < cs.ShardNum; idx++ {
		cs.Edges2Shard[idx] += interEdge[idx] / 2
	}
	// Update MinEdges2Shard, CrossShardEdgeNum
	for _, val := range cs.Edges2Shard {
		if cs.MinEdges2Shard > val {
			cs.MinEdges2Shard = val
		}
	}
}

// changeShardRecompute recomputes parameters when account shard changes, faster
func (cs *CLPAState) changeShardRecompute(v Vertex, old int) {
	new := cs.PartitionMap[v]
	for _, u := range cs.NetGraph.EdgeSet[v] {
		neighborShard := cs.PartitionMap[u]
		if neighborShard != new && neighborShard != old {
			cs.Edges2Shard[new]++
			cs.Edges2Shard[old]--
		} else if neighborShard == new {
			cs.Edges2Shard[old]--
			cs.CrossShardEdgeNum--
		} else {
			cs.Edges2Shard[new]++
			cs.CrossShardEdgeNum++
		}
	}
	cs.MinEdges2Shard = math.MaxInt
	// Update MinEdges2Shard, CrossShardEdgeNum
	for _, val := range cs.Edges2Shard {
		if cs.MinEdges2Shard > val {
			cs.MinEdges2Shard = val
		}
	}
}

// Init_CLPAState sets parameters
func (cs *CLPAState) Init_CLPAState(wp float64, mIter, sn int) {
	cs.WeightPenalty = wp
	cs.MaxIterations = mIter
	cs.ShardNum = sn
	cs.VertexsNumInShard = make([]int, cs.ShardNum)
	cs.PartitionMap = make(map[Vertex]int)
}

// Init_Partition initializes partition using address suffix, should ensure no empty shards at init
func (cs *CLPAState) Init_Partition() {
	// Set default partition parameters
	cs.VertexsNumInShard = make([]int, cs.ShardNum)
	cs.PartitionMap = make(map[Vertex]int)
	for v := range cs.NetGraph.VertexSet {
		var va = v.Addr[len(v.Addr)-8:]
		num, err := strconv.ParseInt(va, 16, 64)
		if err != nil {
			log.Panic()
		}
		cs.PartitionMap[v] = int(num) % cs.ShardNum
		cs.VertexsNumInShard[cs.PartitionMap[v]] += 1
	}
	cs.ComputeEdges2Shard() // Removing this would be faster, but convenient for output (only runs once at Init anyway)
}

// Stable_Init_Partition initializes partition without empty shards
func (cs *CLPAState) Stable_Init_Partition() error {
	// Set default partition parameters
	if cs.ShardNum > len(cs.NetGraph.VertexSet) {
		return errors.New("too many shards, number of shards should be less than nodes. ")
	}
	cs.VertexsNumInShard = make([]int, cs.ShardNum)
	cs.PartitionMap = make(map[Vertex]int)
	cnt := 0
	for v := range cs.NetGraph.VertexSet {
		cs.PartitionMap[v] = int(cnt) % cs.ShardNum
		cs.VertexsNumInShard[cs.PartitionMap[v]] += 1
		cnt++
	}
	cs.ComputeEdges2Shard() // Removing this would be faster, but convenient for output (only runs once at Init anyway)
	return nil
}

// getShard_score computes the score of placing vertex v into uShard
func (cs *CLPAState) getShard_score(v Vertex, uShard int) float64 {
	var score float64
	// Out-degree of vertex v
	v_outdegree := len(cs.NetGraph.EdgeSet[v])
	// Number of edges from vertex v to uShard
	Edgesto_uShard := 0
	for _, item := range cs.NetGraph.EdgeSet[v] {
		if cs.PartitionMap[item] == uShard {
			Edgesto_uShard += 1
		}
	}
	score = float64(Edgesto_uShard) / float64(v_outdegree) * (1 - cs.WeightPenalty*float64(cs.Edges2Shard[uShard])/float64(cs.MinEdges2Shard))
	return score
}

// CLPA_Partition CLPA partition algorithm
func (cs *CLPAState) CLPA_Partition() (map[string]uint64, int) {
	cs.ComputeEdges2Shard()
	fmt.Println("Before running CLPA, cross-shard edge number:", cs.CrossShardEdgeNum)
	res := make(map[string]uint64)
	updateTreshold := make(map[string]int)
	for iter := 0; iter < cs.MaxIterations; iter += 1 { // First loop controls algorithm iterations, constraint
		for v := range cs.NetGraph.VertexSet {
			if updateTreshold[v.Addr] >= 50 {
				continue
			}
			neighborShardScore := make(map[int]float64)
			max_score := -9999.0
			vNowShard, max_scoreShard := cs.PartitionMap[v], cs.PartitionMap[v]
			for _, u := range cs.NetGraph.EdgeSet[v] {
				uShard := cs.PartitionMap[u]
				// For neighbors in uShard, only compute once
				if _, computed := neighborShardScore[uShard]; !computed {
					neighborShardScore[uShard] = cs.getShard_score(v, uShard)
					if max_score < neighborShardScore[uShard] {
						max_score = neighborShardScore[uShard]
						max_scoreShard = uShard
					}
				}
			}
			if vNowShard != max_scoreShard && cs.VertexsNumInShard[vNowShard] > 1 {
				cs.PartitionMap[v] = max_scoreShard
				res[v.Addr] = uint64(max_scoreShard)
				updateTreshold[v.Addr]++
				// Recompute VertexsNumInShard
				cs.VertexsNumInShard[vNowShard] -= 1
				cs.VertexsNumInShard[max_scoreShard] += 1
				// Recompute Wk
				cs.changeShardRecompute(v, vNowShard)
			}
		}
	}
	for sid, n := range cs.VertexsNumInShard {
		fmt.Printf("%d has vertexs: %d\n", sid, n)
	}

	cs.ComputeEdges2Shard()
	fmt.Println("After running CLPA, cross-shard edge number:", cs.CrossShardEdgeNum)
	return res, cs.CrossShardEdgeNum
}

func (cs *CLPAState) EraseEdges() {
	cs.NetGraph.EdgeSet = make(map[Vertex][]Vertex)
}
