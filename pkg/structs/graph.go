/*
 Copyright 2021 The CI/CD Operator Authors

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package structs

// Graph is an graph interface
type Graph interface {
	AddEdge(from, to string)
	IsCyclic() bool
	GetPres(target string) []string
}

// graph is a graph struct
type graph struct {
	nodes     map[string]struct{}
	edgesTo   map[string][]string
	edgesFrom map[string][]string
}

// NewGraph is a constructor of the Graph
func NewGraph() *graph {
	return &graph{
		edgesTo:   map[string][]string{},
		edgesFrom: map[string][]string{},
		nodes:     map[string]struct{}{},
	}
}

// AddEdge adds edge to the Graph
func (g *graph) AddEdge(from, to string) {
	g.edgesTo[from] = append(g.edgesTo[from], to)
	g.edgesFrom[to] = append(g.edgesFrom[to], from)
	g.nodes[from] = struct{}{}
	g.nodes[to] = struct{}{}
}

// IsCyclic returns if the Graph contains any loop
func (g *graph) IsCyclic() bool {
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

func (g *graph) detectCycle(from string, tos []string, visited, explore map[string]bool) bool {
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
func (g *graph) GetPres(target string) []string {
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
