package bexpr

import (
	"fmt"
	"testing"
)

func TestParseCondition(t *testing.T) {
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

	for _, tCase := range []struct {
		input map[string]bool
		want  bool
	}{
		{input: map[string]bool{"a": true, "b": true}, want: false},
		{input: map[string]bool{"a": true, "b": false}, want: true},
		{input: map[string]bool{"a": false, "b": true}, want: false},
		{input: map[string]bool{"a": false, "b": false}, want: false},
	} {
		t.Run("Condition: 'a AND NOT b'", func(t *testing.T) {
			cond, _ := ParseCondition("a AND NOT b")
			got, _ := cond(tCase.input)

			if tCase.want != got {
				t.Fatalf("Want %t, got %t with %v", tCase.want, got, tCase.input)
			}
		})
	}
}
