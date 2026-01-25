package runtime

import "testing"

func TestIsTruthy(t *testing.T) {
	cases := []struct {
		name  string
		value any
		want  bool
	}{
		{"nil", nil, false},
		{"bool-false", false, false},
		{"bool-true", true, true},
		{"empty-string", "", false},
		{"string", "ok", true},
		{"safe-string-empty", SafeString(""), false},
		{"safe-string", SafeString("ok"), true},
		{"int-zero", 0, false},
		{"int", 7, true},
		{"float-zero", 0.0, false},
		{"float", 1.25, true},
		{"empty-slice", []int{}, false},
		{"slice", []int{1}, true},
		{"empty-map", map[string]int{}, false},
		{"map", map[string]int{"a": 1}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := IsTruthy(tc.value); got != tc.want {
				t.Fatalf("IsTruthy(%v) = %v", tc.value, got)
			}
		})
	}
}

func TestIterate(t *testing.T) {
	items := Iterate([]int{1, 2})
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].(int) != 1 || items[1].(int) != 2 {
		t.Fatalf("unexpected items: %v", items)
	}

	items = Iterate([]int{})
	if items != nil {
		t.Fatalf("expected nil for empty slice, got %v", items)
	}

	items = Iterate(map[string]int{"b": 2, "a": 1})
	if len(items) != 2 {
		t.Fatalf("expected 2 items from map, got %d", len(items))
	}
	if items[0].(int) != 1 || items[1].(int) != 2 {
		t.Fatalf("unexpected map items order: %v", items)
	}

	items = Iterate(map[int]int{1: 1})
	if items != nil {
		t.Fatalf("expected nil for non-string key map, got %v", items)
	}
}
