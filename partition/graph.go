// Graph operations
package partition

// Vertex in the graph, representing an account participating in blockchain transactions
type Vertex struct {
	Addr string // Account address
	// Other attributes to be added
}

// Graph describing the current blockchain transaction set
type Graph struct {
	VertexSet map[Vertex]bool     // Vertex set (actually a set)
	EdgeSet   map[Vertex][]Vertex // Records edges between vertices, adjacency list
	// lock      sync.RWMutex       // Lock, but each storage node has its own copy, not needed
}

// ConstructVertex creates a vertex
func (v *Vertex) ConstructVertex(s string) {
	v.Addr = s
}

// AddVertex adds a vertex to the graph
func (g *Graph) AddVertex(v Vertex) {
	if g.VertexSet == nil {
		g.VertexSet = make(map[Vertex]bool)
	}
	g.VertexSet[v] = true
}

// AddEdge adds an edge to the graph
func (g *Graph) AddEdge(u, v Vertex) {
	// If vertex doesn't exist, add edge with weight fixed at 1
	if _, ok := g.VertexSet[u]; !ok {
		g.AddVertex(u)
	}
	if _, ok := g.VertexSet[v]; !ok {
		g.AddVertex(v)
	}
	if g.EdgeSet == nil {
		g.EdgeSet = make(map[Vertex][]Vertex)
	}
	// Undirected graph, use bidirectional edges
	g.EdgeSet[u] = append(g.EdgeSet[u], v)
	g.EdgeSet[v] = append(g.EdgeSet[v], u)
}

// CopyGraph copies a graph
func (dst *Graph) CopyGraph(src Graph) {
	dst.VertexSet = make(map[Vertex]bool)
	for v := range src.VertexSet {
		dst.VertexSet[v] = true
	}
	if src.EdgeSet != nil {
		dst.EdgeSet = make(map[Vertex][]Vertex)
		for v := range src.VertexSet {
			dst.EdgeSet[v] = make([]Vertex, len(src.EdgeSet[v]))
			copy(dst.EdgeSet[v], src.EdgeSet[v])
		}
	}
}

// PrintGraph outputs the graph
func (g Graph) PrintGraph() {
	for v := range g.VertexSet {
		print(v.Addr, " ")
		print("edge:")
		for _, u := range g.EdgeSet[v] {
			print(" ", u.Addr, "\t")
		}
		println()
	}
	println()
}
