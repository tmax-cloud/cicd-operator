package structs

// Graph is a graph struct
type Graph struct {
	nodes     map[string]struct{}
	edgesTo   map[string][]string
	edgesFrom map[string][]string
}

// NewGraph is a constructor of the Graph
func NewGraph() *Graph {
	return &Graph{
		edgesTo:   map[string][]string{},
		edgesFrom: map[string][]string{},
		nodes:     map[string]struct{}{},
	}
}

// AddEdge adds edge to the Graph
func (g *Graph) AddEdge(from, to string) {
	g.edgesTo[from] = append(g.edgesTo[from], to)
	g.edgesFrom[to] = append(g.edgesFrom[to], from)
	g.nodes[from] = struct{}{}
	g.nodes[to] = struct{}{}
}

// IsCyclic returns if the Graph contains any loop
func (g *Graph) IsCyclic() bool {
	visited := map[string]bool{}
	explore := map[string]bool{}

	// Init visited
	for k := range g.nodes {
		visited[k] = false
		explore[k] = false
	}

	for from, tos := range g.edgesTo {
		if visited[from] {
			continue
		}

		if g.detectCycle(from, tos, visited, explore) {
			return true
		}
	}

	return false
}

func (g *Graph) detectCycle(from string, tos []string, visited, explore map[string]bool) bool {
	if explore[from] {
		return true
	}

	explore[from] = true

	for _, to := range tos {
		if visited[to] {
			continue
		}

		if g.detectCycle(to, g.edgesTo[to], visited, explore) {
			return true
		}
	}

	visited[from] = true
	explore[from] = false

	return false
}

// GetPres get the list of parents (pre-s)
func (g *Graph) GetPres(target string) []string {
	var pres []string

	froms, ok := g.edgesFrom[target]
	if !ok {
		return pres
	}

	for _, f := range froms {
		pres = appendUnique(pres, f)
		chPres := g.GetPres(f)
		for _, p := range chPres {
			pres = appendUnique(pres, p)
		}
	}

	return pres
}

func appendUnique(arr []string, val string) []string {
	for _, a := range arr {
		if a == val {
			return arr
		}
	}
	return append(arr, val)
}
