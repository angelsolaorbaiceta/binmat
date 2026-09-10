package bexpr

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCondition(t *testing.T) {

	type conditionTestCase struct {
		input map[string]bool
		want  bool
	}

	runConditionTestCase := func(condition string, tCase conditionTestCase) {
		cond, err := ParseCondition(condition)
		if err != nil {
			t.Fatalf("Want no error, got %s", err)
		}

		got, _ := cond(tCase.input)

		assert.Equal(t, tCase.want, got)
	}

	t.Run("Empty condition always returns false", func(t *testing.T) {
		cond, _ := ParseCondition("")

		if ok, _ := cond(map[string]bool{}); ok {
			t.Fatalf("expected false, got true")
		}
	})

	t.Run("Single variable condition", func(t *testing.T) {
		cond, _ := ParseCondition("a")

		if _, err := cond(map[string]bool{}); err == nil {
			t.Fatalf("expected error, got nil")
		}
		if ok, _ := cond(map[string]bool{"a": true}); !ok {
			t.Fatalf("expected true, got false")
		}
		if ok, _ := cond(map[string]bool{"a": false}); ok {
			t.Fatalf("expected false, got true")
		}
	})

	for _, input := range []string{
		"a a",
		"a b",
	} {
		t.Run(
			fmt.Sprintf("Two contiguous variables yield a parsing error (%s)", input),
			func(t *testing.T) {
				_, err := ParseCondition(input)

				if err == nil {
					t.Fatal("Expected parsing error, got none")
				}
				if err.Reason != ParseErrInvalidAppend {
					t.Fatal("Wrong reason")
				}
			})
	}

	t.Run("An extra trailing variable yields a parsing error", func(t *testing.T) {
		_, err := ParseCondition("a AND b c")
		if err == nil {
			t.Fatal("Expected parsing error, got none")
		}

		if err.Reason != ParseErrInvalidAppend {
			t.Fatal("Wrong reason")
		}
	})

	t.Run("A missing trailing variable yields a parsing error", func(t *testing.T) {
		_, err := ParseCondition("a AND")
		if err == nil {
			t.Fatal("Want parsing error, got none")
		}

		if err.Reason != ParseErrIncompleteExpr {
			t.Fatal("Wrong reason")
		}
	})

	for _, cond := range []string{
		"AND b",
		" AND b",
		"OR b",
		" OR b",
	} {
		t.Run(
			fmt.Sprintf("Missing LHS variable '%s'", cond),
			func(t *testing.T) {
				_, err := ParseCondition(cond)
				if err == nil {
					t.Fatal("Expected parse error")
				}

				if err.Reason != ParseErrIncompleteExpr {
					t.Fatal("Wrong reason")
				}
			})
	}

	for _, tCase := range []struct {
		input map[string]bool
		want  bool
	}{
		{input: map[string]bool{"a": true, "b": true}, want: true},
		{input: map[string]bool{"a": true, "b": false}, want: false},
		{input: map[string]bool{"a": false, "b": true}, want: false},
		{input: map[string]bool{"a": false, "b": false}, want: false},
	} {
		t.Run(
			fmt.Sprintf("Simple AND condition (a=%t, b=%t)", tCase.input["a"], tCase.input["b"]),
			func(t *testing.T) {
				cond, _ := ParseCondition("a AND b")
				got, _ := cond(tCase.input)

				if got != tCase.want {
					t.Errorf("With %v, want %t but got %t", tCase.input, tCase.want, got)
				}
			})
	}

	for _, tCase := range []struct {
		input map[string]bool
		want  bool
	}{
		{input: map[string]bool{"a": true, "b": true}, want: true},
		{input: map[string]bool{"a": true, "b": false}, want: true},
		{input: map[string]bool{"a": false, "b": true}, want: true},
		{input: map[string]bool{"a": false, "b": false}, want: false},
	} {
		t.Run(
			fmt.Sprintf("Simple OR condition (a=%t, b=%t)", tCase.input["a"], tCase.input["b"]),
			func(t *testing.T) {
				cond, _ := ParseCondition("a OR b")
				got, _ := cond(tCase.input)

				if got != tCase.want {
					t.Errorf("With %v, want %t but got %t", tCase.input, tCase.want, got)
				}
			})
	}

	t.Run("Simple NOT", func(t *testing.T) {
		cond, _ := ParseCondition("NOT a")
		if got, _ := cond(map[string]bool{"a": true}); got != false {
			t.Fatalf("Expected false, got true")
		}
		if got, _ := cond(map[string]bool{"a": false}); got != true {
			t.Fatalf("Expected true, got false")
		}
	})

	t.Run("NOT expression shouldn't find a LHS", func(t *testing.T) {
		_, err := ParseCondition("a NOT b")
		if err == nil {
			t.Fatal("Expected parse error")
		}

		if err.Reason != ParseErrInvalidAppend {
			t.Fatalf("Wrong reason: %s", err.Reason)
		}
	})

	for _, tCase := range []conditionTestCase{
		{input: map[string]bool{"a": true, "b": true}, want: false},
		{input: map[string]bool{"a": true, "b": false}, want: true},
		{input: map[string]bool{"a": false, "b": true}, want: false},
		{input: map[string]bool{"a": false, "b": false}, want: false},
	} {
		t.Run("Condition: 'a AND NOT b'", func(t *testing.T) {
			runConditionTestCase("a AND NOT b", tCase)
		})
	}

	for _, tCase := range []conditionTestCase{
		{input: map[string]bool{"a": false, "b": false, "c": false}, want: false},
		{input: map[string]bool{"a": false, "b": false, "c": true}, want: false},
		{input: map[string]bool{"a": false, "b": true, "c": false}, want: false},
		{input: map[string]bool{"a": false, "b": true, "c": true}, want: false},
		{input: map[string]bool{"a": true, "b": false, "c": false}, want: false},
		{input: map[string]bool{"a": true, "b": false, "c": true}, want: false},
		{input: map[string]bool{"a": true, "b": true, "c": false}, want: true},
		{input: map[string]bool{"a": true, "b": true, "c": true}, want: false},
	} {
		t.Run("Condition: 'a AND (b AND NOT c)'", func(t *testing.T) {
			runConditionTestCase("a AND (b AND NOT c)", tCase)
		})
	}
}

// TestParseChainedConditions covers expressions where a group, a NOT or a
// completed binary operation is followed by another binary operator, and the
// precedence rules that apply when AND and OR are mixed without parentheses:
// NOT binds tightest, then AND, then OR, and equal operators associate left.
func TestParseChainedConditions(t *testing.T) {
	type truthCase struct {
		vars map[string]bool
		want bool
	}

	for _, tCase := range []struct {
		cond  string
		cases []truthCase
	}{
		{
			cond: "(a OR b) AND c",
			cases: []truthCase{
				{vars: map[string]bool{"a": false, "b": false, "c": false}, want: false},
				{vars: map[string]bool{"a": false, "b": false, "c": true}, want: false},
				{vars: map[string]bool{"a": false, "b": true, "c": false}, want: false},
				{vars: map[string]bool{"a": false, "b": true, "c": true}, want: true},
				{vars: map[string]bool{"a": true, "b": false, "c": false}, want: false},
				{vars: map[string]bool{"a": true, "b": false, "c": true}, want: true},
				{vars: map[string]bool{"a": true, "b": true, "c": false}, want: false},
				{vars: map[string]bool{"a": true, "b": true, "c": true}, want: true},
			},
		},
		{
			cond: "(a OR b) AND (c OR d)",
			cases: []truthCase{
				{vars: map[string]bool{"a": true, "b": false, "c": false, "d": true}, want: true},
				{vars: map[string]bool{"a": true, "b": false, "c": false, "d": false}, want: false},
				{vars: map[string]bool{"a": false, "b": false, "c": true, "d": true}, want: false},
				{vars: map[string]bool{"a": false, "b": true, "c": true, "d": false}, want: true},
			},
		},
		{
			cond: "a AND b AND c",
			cases: []truthCase{
				{vars: map[string]bool{"a": true, "b": true, "c": true}, want: true},
				{vars: map[string]bool{"a": true, "b": true, "c": false}, want: false},
				{vars: map[string]bool{"a": true, "b": false, "c": true}, want: false},
				{vars: map[string]bool{"a": false, "b": true, "c": true}, want: false},
			},
		},
		{
			cond: "a OR b OR c",
			cases: []truthCase{
				{vars: map[string]bool{"a": false, "b": false, "c": false}, want: false},
				{vars: map[string]bool{"a": false, "b": false, "c": true}, want: true},
				{vars: map[string]bool{"a": true, "b": false, "c": false}, want: true},
			},
		},
		{
			cond: "a AND (b OR c) AND d",
			cases: []truthCase{
				{vars: map[string]bool{"a": true, "b": false, "c": true, "d": true}, want: true},
				{vars: map[string]bool{"a": true, "b": false, "c": false, "d": true}, want: false},
				{vars: map[string]bool{"a": true, "b": true, "c": true, "d": false}, want: false},
				{vars: map[string]bool{"a": false, "b": true, "c": true, "d": true}, want: false},
			},
		},
		{
			cond: "NOT a AND b",
			cases: []truthCase{
				{vars: map[string]bool{"a": false, "b": false}, want: false},
				{vars: map[string]bool{"a": false, "b": true}, want: true},
				{vars: map[string]bool{"a": true, "b": false}, want: false},
				{vars: map[string]bool{"a": true, "b": true}, want: false},
			},
		},
		{
			cond: "NOT a OR b",
			cases: []truthCase{
				{vars: map[string]bool{"a": true, "b": false}, want: false},
				{vars: map[string]bool{"a": true, "b": true}, want: true},
				{vars: map[string]bool{"a": false, "b": false}, want: true},
			},
		},
		{
			cond: "NOT (a OR b) AND c",
			cases: []truthCase{
				{vars: map[string]bool{"a": false, "b": false, "c": true}, want: true},
				{vars: map[string]bool{"a": false, "b": false, "c": false}, want: false},
				{vars: map[string]bool{"a": true, "b": false, "c": true}, want: false},
				{vars: map[string]bool{"a": false, "b": true, "c": true}, want: false},
			},
		},
		{
			// AND binds tighter than OR: a OR (b AND c)
			cond: "a OR b AND c",
			cases: []truthCase{
				{vars: map[string]bool{"a": false, "b": false, "c": false}, want: false},
				{vars: map[string]bool{"a": false, "b": false, "c": true}, want: false},
				{vars: map[string]bool{"a": false, "b": true, "c": false}, want: false},
				{vars: map[string]bool{"a": false, "b": true, "c": true}, want: true},
				{vars: map[string]bool{"a": true, "b": false, "c": false}, want: true},
				{vars: map[string]bool{"a": true, "b": false, "c": true}, want: true},
				{vars: map[string]bool{"a": true, "b": true, "c": false}, want: true},
				{vars: map[string]bool{"a": true, "b": true, "c": true}, want: true},
			},
		},
		{
			// AND binds tighter than OR: (a AND b) OR c
			cond: "a AND b OR c",
			cases: []truthCase{
				{vars: map[string]bool{"a": false, "b": false, "c": false}, want: false},
				{vars: map[string]bool{"a": false, "b": false, "c": true}, want: true},
				{vars: map[string]bool{"a": false, "b": true, "c": false}, want: false},
				{vars: map[string]bool{"a": false, "b": true, "c": true}, want: true},
				{vars: map[string]bool{"a": true, "b": false, "c": false}, want: false},
				{vars: map[string]bool{"a": true, "b": false, "c": true}, want: true},
				{vars: map[string]bool{"a": true, "b": true, "c": false}, want: true},
				{vars: map[string]bool{"a": true, "b": true, "c": true}, want: true},
			},
		},
		{
			// NOT binds tighter than AND, which binds tighter than OR: a OR ((NOT b) AND c)
			cond: "a OR NOT b AND c",
			cases: []truthCase{
				{vars: map[string]bool{"a": false, "b": false, "c": true}, want: true},
				{vars: map[string]bool{"a": false, "b": true, "c": true}, want: false},
				{vars: map[string]bool{"a": false, "b": false, "c": false}, want: false},
				{vars: map[string]bool{"a": true, "b": true, "c": false}, want: true},
			},
		},
	} {
		t.Run(fmt.Sprintf("Condition: '%s'", tCase.cond), func(t *testing.T) {
			cond, err := ParseCondition(tCase.cond)
			if err != nil {
				t.Fatalf("Want no error, got %s", err)
			}

			for _, c := range tCase.cases {
				got, applyErr := cond(c.vars)
				if applyErr != nil {
					t.Fatalf("With %v, want no error, got %s", c.vars, applyErr)
				}
				if got != c.want {
					t.Errorf("With %v, want %t but got %t", c.vars, c.want, got)
				}
			}
		})
	}
}

// TestParseIncompleteChains makes sure malformed chains are rejected instead
// of yielding a condition that would fail when applied.
func TestParseIncompleteChains(t *testing.T) {
	for _, tCase := range []struct {
		cond   string
		reason ParseErrorReason
	}{
		{cond: "a AND AND b", reason: ParseErrInvalidAppend},
		{cond: "a OR AND b", reason: ParseErrInvalidAppend},
		{cond: "NOT AND b", reason: ParseErrInvalidAppend},
		{cond: "(a OR b) c", reason: ParseErrInvalidAppend},
		{cond: "a AND NOT NOT", reason: ParseErrIncompleteExpr},
		{cond: "a AND ()", reason: ParseErrIncompleteExpr},
		{cond: "a AND (b OR)", reason: ParseErrIncompleteExpr},
		{cond: "(a OR b) AND", reason: ParseErrIncompleteExpr},
	} {
		t.Run(fmt.Sprintf("Condition: '%s'", tCase.cond), func(t *testing.T) {
			_, err := ParseCondition(tCase.cond)
			if err == nil {
				t.Fatal("Want parsing error, got none")
			}
			if err.Reason != tCase.reason {
				t.Fatalf("Want reason '%s', got '%s'", tCase.reason, err.Reason)
			}
		})
	}
}

// TestParseUnbalancedParentheses makes sure every opening parenthesis is
// closed and every closing one was opened, while balanced nesting keeps
// working at any depth.
func TestParseUnbalancedParentheses(t *testing.T) {
	for _, cond := range []string{
		"(",
		"(a",
		"(a OR b",
		"a AND (b",
		"a AND ((b OR c)",
		")",
		"a)",
		"a AND b)",
		"a AND b) OR c",
		"(a OR b))",
		"a AND (b OR c))",
	} {
		t.Run(fmt.Sprintf("Condition: '%s'", cond), func(t *testing.T) {
			_, err := ParseCondition(cond)
			if err == nil {
				t.Fatal("Want parsing error, got none")
			}
			if err.Reason != ParseErrUnbalancedParens {
				t.Fatalf("Want reason '%s', got '%s'", ParseErrUnbalancedParens, err.Reason)
			}
		})
	}

	for _, tCase := range []struct {
		cond string
		vars map[string]bool
		want bool
	}{
		{cond: "((a))", vars: map[string]bool{"a": true}, want: true},
		{cond: "((a))", vars: map[string]bool{"a": false}, want: false},
		{cond: "(a AND (b OR (c AND NOT d)))", vars: map[string]bool{"a": true, "b": false, "c": true, "d": false}, want: true},
		{cond: "(a AND (b OR (c AND NOT d)))", vars: map[string]bool{"a": true, "b": false, "c": true, "d": true}, want: false},
		{cond: "((a OR b) AND c) OR d", vars: map[string]bool{"a": false, "b": true, "c": true, "d": false}, want: true},
		{cond: "((a OR b) AND c) OR d", vars: map[string]bool{"a": false, "b": true, "c": false, "d": false}, want: false},
		{cond: "((a OR b) AND c) OR d", vars: map[string]bool{"a": false, "b": false, "c": false, "d": true}, want: true},
	} {
		t.Run(fmt.Sprintf("Balanced '%s' with %v", tCase.cond, tCase.vars), func(t *testing.T) {
			cond, err := ParseCondition(tCase.cond)
			if err != nil {
				t.Fatalf("Want no error, got %s", err)
			}

			got, applyErr := cond(tCase.vars)
			if applyErr != nil {
				t.Fatalf("Want no error, got %s", applyErr)
			}
			if got != tCase.want {
				t.Errorf("Want %t, got %t", tCase.want, got)
			}
		})
	}
}

// TestParseUnexpectedCharacters makes sure tokenizer errors surface through
// ParseCondition, so a typo can't turn into a different, valid condition.
func TestParseUnexpectedCharacters(t *testing.T) {
	for _, cond := range []string{
		"a & b",
		"Foo AND bar",
		"a AND b-c",
		"ANDY",
	} {
		t.Run(fmt.Sprintf("Condition: '%s'", cond), func(t *testing.T) {
			_, err := ParseCondition(cond)
			if err == nil {
				t.Fatal("Want parsing error, got none")
			}
			if err.Reason != ParseErrUnexpectedChar {
				t.Fatalf("Want reason '%s', got '%s'", ParseErrUnexpectedChar, err.Reason)
			}
		})
	}
}
