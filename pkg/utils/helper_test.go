package utils

import (
	"testing"
)

// ---------- TABLE-DRIVEN ----------

func TestReduce_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		slice    []int
		initial  int
		expected int
	}{
		{"sum empty", []int{}, 0, 0},
		{"sum values", []int{1, 2, 3}, 0, 6},
		{"sum with initial", []int{1, 2}, 10, 13},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Reduce(
				tt.slice,
				func(a, b int) int { return a + b },
				tt.initial,
			)

			if result != tt.expected {
				t.Fatalf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestReduce_IntProduct(t *testing.T) {
	result := Reduce(
		[]int{1, 2, 3, 4},
		func(a, b int) int { return a * b },
		1,
	)

	if result != 24 {
		t.Fatalf("expected 24, got %d", result)
	}
}

// ---------- SINGLE ----------

func TestReduce_SingleElement(t *testing.T) {
	result := Reduce(
		[]int{5},
		func(a, b int) int { return a + b },
		10,
	)

	if result != 15 {
		t.Fatalf("expected 15, got %d", result)
	}
}

// ---------- STRINGS ----------

func TestReduce_StringConcat(t *testing.T) {
	result := Reduce(
		[]string{"a", "b", "c"},
		func(a, b string) string { return a + b },
		"",
	)

	if result != "abc" {
		t.Fatalf("expected 'abc', got '%s'", result)
	}
}

// ---------- STRUCTS ----------

type User struct {
	Name string
	Age  int
}

func TestReduce_StructAggregation(t *testing.T) {
	users := []User{
		{"Alice", 30},
		{"Bob", 20},
	}

	result := Reduce(
		users,
		func(a, b User) User {
			return User{
				Name: a.Name + b.Name,
				Age:  a.Age + b.Age,
			}
		},
		User{},
	)

	if result.Age != 50 {
		t.Fatalf("expected age 50, got %d", result.Age)
	}
}

// ---------- NON-COMMUTATIVE ----------

func TestReduce_NonCommutative(t *testing.T) {
	result := Reduce(
		[]int{1, 2, 3},
		func(a, b int) int { return a - b },
		10,
	)

	// (((10 - 1) - 2) - 3) = 4
	if result != 4 {
		t.Fatalf("expected 4, got %d", result)
	}
}

// ---------- ORDER SENSITIVITY ----------

func TestReduce_OrderMatters(t *testing.T) {
	left := Reduce(
		[]int{1, 2, 3},
		func(a, b int) int { return a*10 + b },
		0,
	)

	right := Reduce(
		[]int{3, 2, 1},
		func(a, b int) int { return a*10 + b },
		0,
	)

	if left == right {
		t.Fatal("expected order-sensitive results to differ")
	}
}

// ---------- NO SLICE MUTATION ----------

func TestReduce_DoesNotMutateSlice(t *testing.T) {
	input := []int{1, 2, 3}
	_ = Reduce(
		input,
		func(a, b int) int { return a + b },
		0,
	)

	expected := []int{1, 2, 3}
	for i := range input {
		if input[i] != expected[i] {
			t.Fatal("input slice was mutated")
		}
	}
}

func TestReduce_SumStructQuantity(t *testing.T) {
	type Item struct {
		Name     string
		Quantity int
	}

	items := []Item{
		{"Apple", 3},
		{"Banana", 5},
		{"Orange", 2},
	}

	result := Reduce(
		items,
		func(a int, b Item) int {
			return a + b.Quantity
		},
		0,
	)

	if result != 10 {
		t.Fatalf("expected total quantity 10, got %d", result)
	}
}
