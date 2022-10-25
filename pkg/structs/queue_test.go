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

type testType struct {
	body int
}

func (t *testType) DeepCopy() Item {
	if t == nil {
		return nil
	}
	return &testType{body: t.body}
}

func (t *testType) Equals(another Item) bool {
	if t == nil || another == nil {
		return false
	}
	b, ok := another.(*testType)
	if !ok {
		return false
	}
	return t.body == b.body
}

func compare(_a, _b Item) bool {
	if _a == nil || _b == nil {
		return false
	}
	a, aOk := _a.(*testType)
	b, bOk := _b.(*testType)
	if !aOk || !bOk {
		return false
	}
	return a.body < b.body
}

func initQueue() *sortedUniqueList {
	q := NewSortedUniqueQueue(compare)
	q.Add(&testType{body: 2})
	q.Add(&testType{body: 3})
	q.Add(&testType{body: 3})
	q.Add(&testType{body: 7})
	q.Add(&testType{body: 6})
	q.Add(&testType{body: 4})
	q.Add(&testType{body: 1})
	q.Add(&testType{body: 2})
	q.Add(&testType{body: 2})
	q.Add(&testType{body: 5})
	q.Add(&testType{body: 5})
	q.Add(&testType{body: 3})
	q.Add(&testType{body: 5})

	return q
}

func TestSortedUniqueQueue_Put(t *testing.T) {
	q := initQueue()

	n := q.nodes

	if n == nil {
		t.Fatal("nodes are nil")
	}

	i := 0
	for n != nil {
		it, ok := n.item.(*testType)
		assert.Equal(t, true, ok, "item is not a testType")
		assert.Equal(t, i+1, it.body, "items are not sorted")
		n = n.next
		i++
	}
}

func TestSortedUniqueList_First(t *testing.T) {
	q := initQueue()
	f := q.First()
	if f == nil {
		t.Fatal("first is nil")
	}

	ff, ok := f.(*testType)
	assert.Equal(t, true, ok, "item is not a testType")
	assert.Equal(t, 1, ff.body, "item is not sorted")
}

func TestSortedUniqueList_Delete(t *testing.T) {
	q := initQueue()

	var i *testType

	f := q.First()
	if f == nil {
		t.Fatal("first is nil")
	}
	ff, ok := f.(*testType)
	assert.Equal(t, true, ok, "item is not a testType")
	assert.Equal(t, 1, ff.body, "item is not sorted")

	i = &testType{body: 1}
	q.Delete(i)

	f = q.First()
	if f == nil {
		t.Fatal("first is nil")
	}
	ff, ok = f.(*testType)
	assert.Equal(t, true, ok, "item is not a testType")
	assert.Equal(t, 2, ff.body, "item is not deleted")
}

func TestSortedUniqueList_ForEach(t *testing.T) {
	q := initQueue()

	i := 1
	q.ForEach(func(item Item) {
		it, ok := item.(*testType)
		assert.Equal(t, true, ok, "item is not a testType")

		assert.Equal(t, i, it.body, "item is not looped correctly")
		i++
	})
}
