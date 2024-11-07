package bexpr

import (
	"fmt"
	"testing"
)

func TestGroupExpr(t *testing.T) {
	t.Run("without operation", func(t *testing.T) {
		group := &groupCondition{}

		if group.hasOp() {
			t.Fatal("Want no op, got op")
		}
		if got := group.getOp(); got != nil {
			t.Fatal("Want no op, got op")
		}
		if got := group.String(); got != "(??)" {
			t.Fatalf("Want '(??)', got %s", got)
		}
	})

	for _, tCase := range []struct {
		op   conditionExpr
		want string
	}{
		{op: &varCondition{varName: "a"}, want: "(a)"},
		{op: &groupCondition{}, want: "((??))"},
		{op: &notCondition{}, want: "(NOT ??)"},
	} {
		t.Run(
			fmt.Sprintf("set '%s' expecting '%s'", tCase.op, tCase.want),
			func(t *testing.T) {
				group := &groupCondition{}
				group.setOp(tCase.op)

				if !group.hasOp() {
					t.Fatal("Want operation, got none")
				}
				if got := group.getOp(); got != tCase.op {
					t.Fatalf("Want %s, got %s", tCase.op, got)
				}
				if got := group.String(); got != tCase.want {
					t.Fatalf("Want '%s', got '%s'", tCase.want, got)
				}
			})
	}

	for _, op := range []conditionExpr{
		&andCondition{},
		&orCondition{},
	} {
		t.Run(
			fmt.Sprintf("can't set binary op '%s'", op),
			func(t *testing.T) {
				group := &groupCondition{}
				err := group.setOp(op)

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
			want: "(NOT NOT ??)",
		},
		{
			ops:  []conditionExpr{&notCondition{}, &notCondition{}, &varCondition{varName: "a"}},
			want: "(NOT NOT a)",
		},
		{
			ops:  []conditionExpr{&notCondition{}, &groupCondition{}, &varCondition{varName: "a"}},
			want: "(NOT (a))",
		},
	} {
		t.Run(
			fmt.Sprintf("nested '%s'", tCase.want),
			func(t *testing.T) {
				group := &groupCondition{}
				for _, op := range tCase.ops {
					group.setOp(op)
				}

				if got := group.String(); got != tCase.want {
					t.Fatalf("Want '%s', got '%s'", tCase.want, got)
				}
			})
	}

	for _, group := range []*groupCondition{
		{expr: &varCondition{varName: "a"}},
		{expr: &notCondition{&varCondition{varName: "a"}}},
	} {
		t.Run(
			fmt.Sprintf("Can't append to complete %s", group),
			func(t *testing.T) {
				err := group.setOp(&varCondition{varName: "xyz"})
				if err == nil {
					t.Fatal("Expected error")
				}
			})
	}
}
