package bexpr

import (
	"fmt"
	"testing"
)

func TestOrExpr(t *testing.T) {
	t.Run("without lhs or rhs", func(t *testing.T) {
		or := &orCondition{}

		if or.hasLhs() {
			t.Fatal("Want no lhs")
		}
		if got := or.getLhs(); got != nil {
			t.Fatalf("Want no lhs, got '%s'", got)
		}
		if or.hasRhs() {
			t.Fatal("Want no rhs")
		}
		if got := or.getRhs(); got != nil {
			t.Fatalf("Want no rhs, got '%s'", got)
		}
		if got := or.String(); got != "?? OR ??" {
			t.Fatalf("Want '?? OR ??', got '%s'", got)
		}
	})

	for _, tCase := range []struct {
		op   conditionExpr
		want string
	}{
		{op: &varCondition{varName: "a"}, want: "?? OR a"},
		{op: &groupCondition{}, want: "?? OR (??)"},
		{op: &notCondition{}, want: "?? OR NOT ??"},
	} {
		t.Run(
			fmt.Sprintf("set '%s' as RHS expecting '%s'", tCase.op, tCase.want),
			func(t *testing.T) {
				or := &orCondition{}
				or.setRhs(tCase.op)

				if !or.hasRhs() {
					t.Fatal("Want rhs, got none")
				}
				if got := or.getRhs(); got != tCase.op {
					t.Fatalf("Want '%s', got '%s'", tCase.op, got)
				}
				if got := or.String(); got != tCase.want {
					t.Fatalf("Want '%s', got '%s'", tCase.want, got)
				}
			})
	}

	for _, op := range []conditionExpr{
		&andCondition{},
		&orCondition{},
	} {
		t.Run(
			fmt.Sprintf("can't set binary op '%s' as RHS", op),
			func(t *testing.T) {
				or := &orCondition{}
				err := or.setRhs(op)

				if err == nil {
					t.Fatal("Expected error")
				}
			})
	}

	for _, tCase := range []struct {
		ops  []conditionExpr
		want string
	}{
		{
			ops:  []conditionExpr{&notCondition{}, &notCondition{}},
			want: "?? OR NOT NOT ??",
		},
		{
			ops:  []conditionExpr{&notCondition{}, &notCondition{}, &varCondition{varName: "a"}},
			want: "?? OR NOT NOT a",
		},
		{
			ops:  []conditionExpr{&notCondition{}, &groupCondition{}, &varCondition{varName: "a"}},
			want: "?? OR NOT (a)",
		},
	} {
		t.Run(
			fmt.Sprintf("nested '%s'", tCase.want),
			func(t *testing.T) {
				or := &orCondition{}
				for _, op := range tCase.ops {
					or.setRhs(op)
				}

				if got := or.String(); got != tCase.want {
					t.Fatalf("Want '%s', got '%s'", tCase.want, got)
				}
			})
	}

	for _, or := range []*orCondition{
		{rhs: &varCondition{varName: "a"}},
		{rhs: &groupCondition{&varCondition{varName: "a"}}},
		{rhs: &groupCondition{&notCondition{&varCondition{varName: "a"}}}},
	} {
		t.Run(
			fmt.Sprintf("Can't append to complete %s", or),
			func(t *testing.T) {
				err := or.setRhs(&varCondition{varName: "xyz"})
				if err == nil {
					t.Fatal("Expected error")
				}
			})
	}
}
