package structs

// Graph is a graph struct
type Graph struct {
	edgesTo   map[string][]string
	edgesFrom map[string][]string
}

// NewGraph is a constructor of the Graph
func NewGraph() *Graph {
	return &Graph{
		edgesTo:   map[string][]string{},
		edgesFrom: map[string][]string{},
	}
}

// AddEdge adds edge to the Graph
func (g *Graph) AddEdge(from, to string) {
	g.edgesTo[from] = append(g.edgesTo[from], to)
	g.edgesFrom[to] = append(g.edgesFrom[to], from)
}

// IsCyclic returns if the Graph contains any loop
func (g *Graph) IsCyclic() bool {
	return false //TODO - need it! not to infinitely loop in GetPres
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
