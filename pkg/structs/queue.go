package structs

import (
	"sync"
)

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

type CompareFunc func(a Item, b Item) bool

type SortedUniqueList struct {
	nodes *node
	lock  sync.Mutex

	compareFunc CompareFunc
}

func (q *SortedUniqueList) Add(item Item) {
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

func (q *SortedUniqueList) First() Item {
	q.lock.Lock()
	defer q.lock.Unlock()

	n := q.nodes

	if n == nil {
		return nil
	}

	return n.item
}

type IteratorFunc func(Item)

func (q *SortedUniqueList) ForEach(iteratorFunc IteratorFunc) {
	q.lock.Lock()
	defer q.lock.Unlock()

	n := q.nodes

	for n != nil {
		iteratorFunc(n.item)
		n = n.next
	}
}

func (q *SortedUniqueList) Delete(i Item) {
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

func (q *SortedUniqueList) Len() int {
	i := 0
	q.ForEach(func(_ Item) {
		i++
	})
	return i
}

func NewSortedUniqueQueue(compareFunc CompareFunc) *SortedUniqueList {
	return &SortedUniqueList{lock: sync.Mutex{}, compareFunc: compareFunc}
}
