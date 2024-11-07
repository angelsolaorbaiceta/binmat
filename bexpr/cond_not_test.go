package bexpr

import (
	"fmt"
	"testing"
)

func TestNotExpr(t *testing.T) {
	t.Run("without operation", func(t *testing.T) {
		not := &notCondition{}

		if not.hasOp() {
			t.Fatal("Want no op, got op")
		}
		if got := not.getOp(); got != nil {
			t.Fatal("Want no op, got op")
		}
		if got := not.String(); got != "NOT ??" {
			t.Fatalf("Want 'NOT ??', got %s", got)
		}
	})

	for _, tCase := range []struct {
		op   conditionExpr
		want string
	}{
		{op: &varCondition{varName: "a"}, want: "NOT a"},
		{op: &notCondition{}, want: "NOT NOT ??"},
	} {
		t.Run(
			fmt.Sprintf("set '%s' expecting '%s'", tCase.op, tCase.want),
			func(t *testing.T) {
				not := &notCondition{}
				not.setOp(tCase.op)

				if !not.hasOp() {
					t.Fatal("Want operation, got none")
				}
				if got := not.getOp(); got != tCase.op {
					t.Fatalf("Want %s, got %s", tCase.op, got)
				}
				if got := not.String(); got != tCase.want {
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
				not := &notCondition{}
				err := not.setOp(op)

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
			want: "NOT NOT NOT ??",
		},
		{
			ops:  []conditionExpr{&notCondition{}, &notCondition{}, &varCondition{varName: "a"}},
			want: "NOT NOT NOT a",
		},
	} {
		t.Run(
			fmt.Sprintf("nested '%s'", tCase.want),
			func(t *testing.T) {
				not := &notCondition{}
				for _, op := range tCase.ops {
					not.setOp(op)
				}

				if got := not.String(); got != tCase.want {
					t.Fatalf("Want '%s', got '%s'", tCase.want, got)
				}
			})
	}

	for _, not := range []*notCondition{
		{op: &varCondition{varName: "a"}},
		{op: &notCondition{&varCondition{varName: "a"}}},
	} {
		t.Run(
			fmt.Sprintf("Can't append to complete %s", not),
			func(t *testing.T) {
				err := not.setOp(&varCondition{varName: "xyz"})
				if err == nil {
					t.Fatal("Expected error")
				}
			})
	}
}
