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
	"sync"
)

// Item is an interface for the nodes to be stored in a queue
type Item interface {
	DeepCopy() Item
	Equals(Item) bool
}

type node struct {
	item Item

	prev *node
	next *node
}

func newNode(item Item) *node {
	return &node{item: item.DeepCopy()}
}

// CompareFunc is a function to sort nodes in a queue
type CompareFunc func(a Item, b Item) bool

// SortedUniqueList is an interface of sortedUniqueList
type SortedUniqueList interface {
	Add(item Item)
	First() Item
	ForEach(iteratorFunc IteratorFunc)
	Delete(i Item)
	Len() int
}

// sortedUniqueList is a kind of priority queues, whose nodes are sorted
// Also, uniqueness of the node is guaranteed
type sortedUniqueList struct {
	nodes *node
	lock  sync.Mutex

	compareFunc CompareFunc
}

// NewSortedUniqueQueue is a constructor for the sortedUniqueList
func NewSortedUniqueQueue(compareFunc CompareFunc) *sortedUniqueList {
	return &sortedUniqueList{lock: sync.Mutex{}, compareFunc: compareFunc}
}

// Add a node to the sortedUniqueList
func (q *sortedUniqueList) Add(item Item) {
	q.lock.Lock()
	defer q.lock.Unlock()

	var prevPtr *node
	nextPtr := q.nodes

	if q.compareFunc != nil {
		for nextPtr != nil {
			// Sort
			if q.compareFunc(item, nextPtr.item) {
				break
			}
			prevPtr = nextPtr
			nextPtr = nextPtr.next
		}
	}

	if prevPtr != nil {
		// Guarantee uniqueness
		if item.Equals(prevPtr.item) {
			return
		}
	}

	if nextPtr != nil {
		// Guarantee uniqueness
		if item.Equals(nextPtr.item) {
			return
		}
	}

	n := newNode(item)
	n.next = nextPtr
	n.prev = prevPtr
	if nextPtr != nil {
		nextPtr.prev = n
	}
	if prevPtr != nil {
		prevPtr.next = n
	} else {
		q.nodes = n
	}
}

// First retrieves the first node in the queue
func (q *sortedUniqueList) First() Item {
	q.lock.Lock()
	defer q.lock.Unlock()

	n := q.nodes

	if n == nil {
		return nil
	}

	return n.item
}

// IteratorFunc is a function to be used for each item in the queue
type IteratorFunc func(Item)

// ForEach runs IteratorFunc for each item in the queue
func (q *sortedUniqueList) ForEach(iteratorFunc IteratorFunc) {
	q.lock.Lock()
	defer q.lock.Unlock()

	n := q.nodes

	for n != nil {
		iteratorFunc(n.item)
		n = n.next
	}
}

// Delete deletes a node from the queue
func (q *sortedUniqueList) Delete(i Item) {
	q.lock.Lock()
	defer q.lock.Unlock()

	n := q.nodes

	for n != nil {
		if n.item.Equals(i) {
			if n.next != nil {
				n.next.prev = n.prev
			}
			if n.prev != nil {
				n.prev.next = n.next
			} else {
				q.nodes = n.next
			}
			return
		}
		n = n.next
	}
}

// Len returns the length of the queue
func (q *sortedUniqueList) Len() int {
	i := 0
	q.ForEach(func(_ Item) {
		i++
	})
	return i
}
