package runtime

import "testing"

func TestEvalArg(t *testing.T) {
	ctx := NewContext(map[string]any{
		"name": "Ada",
	})

	if got := EvalArg(ctx, ArgPath, "name"); got != "Ada" {
		t.Fatalf("ArgPath name = %v", got)
	}
	if got := EvalArg(ctx, ArgString, "text"); got != "text" {
		t.Fatalf("ArgString text = %v", got)
	}
	if got := EvalArg(ctx, ArgBool, "true"); got != true {
		t.Fatalf("ArgBool true = %v", got)
	}
	if got := EvalArg(ctx, ArgNull, ""); got != nil {
		t.Fatalf("ArgNull = %v", got)
	}
	if got := EvalArg(ctx, ArgNumber, "42"); got != int64(42) {
		t.Fatalf("ArgNumber 42 = %T(%v)", got, got)
	}
	if got := EvalArg(ctx, ArgNumber, "3.14"); got != float64(3.14) {
		t.Fatalf("ArgNumber 3.14 = %T(%v)", got, got)
	}
	if got := EvalArg(ctx, ArgPath, "missing"); got != nil {
		t.Fatalf("ArgPath missing = %v", got)
	}
}
