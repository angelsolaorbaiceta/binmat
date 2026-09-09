package bexpr

import (
	"fmt"
	"testing"
)

func TestAndExpr(t *testing.T) {
	t.Run("without lhs or rhs", func(t *testing.T) {
		and := &andCondition{}

		if and.hasLhs() {
			t.Fatal("Want no lhs")
		}
		if got := and.getLhs(); got != nil {
			t.Fatalf("Want no lhs, got '%s'", got)
		}
		if and.hasRhs() {
			t.Fatal("Want no rhs")
		}
		if got := and.getRhs(); got != nil {
			t.Fatalf("Want no rhs, got '%s'", got)
		}
		if got := and.String(); got != "?? AND ??" {
			t.Fatalf("Want '?? AND ??', got '%s'", got)
		}
	})

	for _, tCase := range []struct {
		op   conditionExpr
		want string
	}{
		{op: &varCondition{varName: "a"}, want: "?? AND a"},
		{op: &groupCondition{}, want: "?? AND (??)"},
		{op: &notCondition{}, want: "?? AND NOT ??"},
	} {
		t.Run(
			fmt.Sprintf("set '%s' as RHS expecting '%s'", tCase.op, tCase.want),
			func(t *testing.T) {
				and := &andCondition{}
				and.setRhs(tCase.op)

				if !and.hasRhs() {
					t.Fatal("Want rhs, got none")
				}
				if got := and.getRhs(); got != tCase.op {
					t.Fatalf("Want '%s', got '%s'", tCase.op, got)
				}
				if got := and.String(); got != tCase.want {
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
				and := &andCondition{}
				err := and.setRhs(op)

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
			want: "?? AND NOT NOT ??",
		},
		{
			ops:  []conditionExpr{&notCondition{}, &notCondition{}, &varCondition{varName: "a"}},
			want: "?? AND NOT NOT a",
		},
		{
			ops:  []conditionExpr{&notCondition{}, &groupCondition{}, &varCondition{varName: "a"}},
			want: "?? AND NOT (a)",
		},
	} {
		t.Run(
			fmt.Sprintf("nested '%s'", tCase.want),
			func(t *testing.T) {
				and := &andCondition{}
				for _, op := range tCase.ops {
					and.setRhs(op)
				}

				if got := and.String(); got != tCase.want {
					t.Fatalf("Want '%s', got '%s'", tCase.want, got)
				}
			})
	}

	for _, and := range []*andCondition{
		{rhs: &varCondition{varName: "a"}},
		{rhs: &groupCondition{&varCondition{varName: "a"}}},
		{rhs: &groupCondition{&notCondition{&varCondition{varName: "a"}}}},
	} {
		t.Run(
			fmt.Sprintf("Can't append to complete %s", and),
			func(t *testing.T) {
				err := and.setRhs(&varCondition{varName: "xyz"})
				if err == nil {
					t.Fatal("Expected error")
				}
			})
	}
}
