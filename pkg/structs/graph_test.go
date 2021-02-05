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
