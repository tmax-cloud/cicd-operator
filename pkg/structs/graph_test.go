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

import (
	"github.com/bmizerany/assert"
	"testing"
)

func TestGraph_AddEdge(t *testing.T) {
	/*
		a-1  -->  a-2  --> a-4
		     \->  a-3  -/

		b-1  -->  b-2
	*/
	graph := NewGraph()
	graph.AddEdge("a-1", "a-2")
	graph.AddEdge("a-1", "a-3")
	graph.AddEdge("a-2", "a-4")
	graph.AddEdge("a-3", "a-4")
	graph.AddEdge("b-1", "b-2")

	target := "a-1"
	assert.Equal(t, 2, len(graph.edgesTo[target]))
	assert.Equal(t, "a-2", graph.edgesTo[target][0])
	assert.Equal(t, "a-3", graph.edgesTo[target][1])
	assert.Equal(t, 0, len(graph.edgesFrom[target]))

	target = "a-2"
	assert.Equal(t, 1, len(graph.edgesTo[target]))
	assert.Equal(t, "a-4", graph.edgesTo[target][0])
	assert.Equal(t, 1, len(graph.edgesFrom[target]))
	assert.Equal(t, "a-1", graph.edgesFrom[target][0])

	target = "a-3"
	assert.Equal(t, 1, len(graph.edgesTo[target]))
	assert.Equal(t, "a-4", graph.edgesTo[target][0])
	assert.Equal(t, 1, len(graph.edgesFrom[target]))
	assert.Equal(t, "a-1", graph.edgesFrom[target][0])

	target = "a-4"
	assert.Equal(t, 0, len(graph.edgesTo[target]))
	assert.Equal(t, 2, len(graph.edgesFrom[target]))
	assert.Equal(t, "a-2", graph.edgesFrom[target][0])
	assert.Equal(t, "a-3", graph.edgesFrom[target][1])
}

func TestGraph_GetPres(t *testing.T) {
	/*
		a-1  -->  a-2  --> a-4
		     \->  a-3  -/

		b-1  -->  b-2
	*/

	graph := NewGraph()
	graph.AddEdge("a-1", "a-2")
	graph.AddEdge("a-1", "a-3")
	graph.AddEdge("a-2", "a-4")
	graph.AddEdge("a-3", "a-4")
	graph.AddEdge("b-1", "b-2")

	target := "a-1"
	pres := graph.GetPres(target)
	assert.Equal(t, 0, len(pres))

	target = "a-2"
	pres = graph.GetPres(target)
	assert.Equal(t, 1, len(pres))
	assert.Equal(t, "a-1", pres[0])

	target = "a-3"
	pres = graph.GetPres(target)
	assert.Equal(t, 1, len(pres))
	assert.Equal(t, "a-1", pres[0])

	target = "a-4"
	pres = graph.GetPres(target)
	assert.Equal(t, 3, len(pres))
	assert.Equal(t, "a-2", pres[0])
	assert.Equal(t, "a-1", pres[1])
	assert.Equal(t, "a-3", pres[2])

	target = "b-1"
	pres = graph.GetPres(target)
	assert.Equal(t, 0, len(pres))

	target = "b-2"
	pres = graph.GetPres(target)
	assert.Equal(t, 1, len(pres))
	assert.Equal(t, "b-1", pres[0])
}

func TestGraph_IsCyclic(t *testing.T) {
	/*
		a-1  -->  a-2  --> a-4
		     \->  a-3  -/

		b-1  -->  b-2
	*/

	graph := NewGraph()
	graph.AddEdge("a-1", "a-2")
	graph.AddEdge("a-1", "a-3")
	graph.AddEdge("a-2", "a-4")
	graph.AddEdge("a-3", "a-4")
	graph.AddEdge("b-1", "b-2")

	assert.Equal(t, false, graph.IsCyclic())

	/*

		1  --> 2 --> 5
		|\     ↑     ↓
		|  \-> 3 <-  6
		|          \ ↓
		\-> 4       \7
	*/

	graph = NewGraph()
	graph.AddEdge("1", "2")
	graph.AddEdge("1", "3")
	graph.AddEdge("1", "4")
	graph.AddEdge("2", "5")
	graph.AddEdge("3", "2")
	graph.AddEdge("5", "6")
	graph.AddEdge("6", "7")
	graph.AddEdge("7", "3")

	assert.Equal(t, true, graph.IsCyclic())
}
