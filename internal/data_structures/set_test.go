package data_structures

import (
	"reflect"
	"testing"
)

func sliceToMap[T comparable](items []T) map[T]struct{} {
	m := make(map[T]struct{}, len(items))
	for _, v := range items {
		m[v] = struct{}{}
	}
	return m
}

func assertSameItems[T comparable](t *testing.T, got, want []T) {
	t.Helper()
	if !reflect.DeepEqual(sliceToMap(got), sliceToMap(want)) {
		t.Fatalf("items mismatch\n got:  %v\n want: %v", got, want)
	}
}

func TestNewSet(t *testing.T) {
	s := NewSet[int]()
	if s == nil {
		t.Fatal("New returned nil")
	}
	if s.Size() != 0 {
		t.Fatalf("expected empty set, got size %d", s.Size())
	}
}

func TestAddAndContains(t *testing.T) {
	s := NewSet[int]()

	s.Add(1)
	s.Add(2)
	s.Add(2) // duplicate

	if !s.Contains(1) || !s.Contains(2) {
		t.Fatal("expected set to contain added elements")
	}
	if s.Size() != 2 {
		t.Fatalf("expected size 2, got %d", s.Size())
	}
}

func TestRemove(t *testing.T) {
	s := NewSet[int]()
	s.Add(1)
	s.Add(2)

	s.Remove(1)
	if s.Contains(1) {
		t.Fatal("expected element to be removed")
	}

	// removing non-existent element should be safe
	s.Remove(42)

	if s.Size() != 1 {
		t.Fatalf("expected size 1, got %d", s.Size())
	}
}

func TestClear(t *testing.T) {
	s := NewSet[string]()
	s.Add("a")
	s.Add("b")

	s.Clear()

	if s.Size() != 0 {
		t.Fatalf("expected empty set after Clear, got %d", s.Size())
	}
	if len(s.Items()) != 0 {
		t.Fatal("expected Items to be empty after Clear")
	}
}

func TestItems(t *testing.T) {
	s := NewSet[int]()
	values := []int{1, 2, 3}

	for _, v := range values {
		s.Add(v)
	}

	items := s.Items()
	assertSameItems(t, items, values)

	// modifying returned slice must not affect set
	items[0] = 99
	if s.Contains(99) {
		t.Fatal("Items leaked internal state")
	}
}

func TestUnion(t *testing.T) {
	a := NewSet[int]()
	b := NewSet[int]()

	a.Add(1)
	a.Add(2)
	b.Add(2)
	b.Add(3)

	u := a.Union(b)

	assertSameItems(t, u.Items(), []int{1, 2, 3})

	// original sets must remain unchanged
	assertSameItems(t, a.Items(), []int{1, 2})
	assertSameItems(t, b.Items(), []int{2, 3})
}

func TestIntersection(t *testing.T) {
	a := NewSet[int]()
	b := NewSet[int]()

	a.Add(1)
	a.Add(2)
	a.Add(3)
	b.Add(2)
	b.Add(3)
	b.Add(4)

	i := a.Intersection(b)

	assertSameItems(t, i.Items(), []int{2, 3})
}

func TestDifference(t *testing.T) {
	a := NewSet[int]()
	b := NewSet[int]()

	a.Add(1)
	a.Add(2)
	a.Add(3)
	b.Add(2)

	d := a.Difference(b)

	assertSameItems(t, d.Items(), []int{1, 3})
}

func TestIsSubset(t *testing.T) {
	a := NewSet[int]()
	b := NewSet[int]()

	a.Add(1)
	a.Add(2)
	b.Add(1)
	b.Add(2)
	b.Add(3)

	if !a.IsSubset(b) {
		t.Fatal("expected a to be subset of b")
	}
	if b.IsSubset(a) {
		t.Fatal("did not expect b to be subset of a")
	}

	empty := NewSet[int]()
	if !empty.IsSubset(b) {
		t.Fatal("empty set should be subset of any set")
	}
}

func TestSetWithStructType(t *testing.T) {
	type key struct {
		ID int
	}

	s := NewSet[key]()
	k1 := key{ID: 1}
	k2 := key{ID: 2}

	s.Add(k1)
	s.Add(k2)
	s.Add(k1)

	if s.Size() != 2 {
		t.Fatalf("expected size 2, got %d", s.Size())
	}
}
