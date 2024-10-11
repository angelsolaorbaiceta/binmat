package signature

import (
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
		t.Run("Two contiguous variables yield a parsing error", func(t *testing.T) {
			_, err := ParseCondition(input)

			if err == nil {
				t.Fatalf("Expected parsing error, got none")
			}
			parseErr, ok := err.(ErrConditionParse)
			if !ok {
				t.Fatalf("Expected ErrConditionParse, got %v", err)
			}
			if parseErr.Reason != ParseErrReasonContigVars {
				t.Fatalf("Expected ErrConditionParse to contain a ReasonContigVars reason")
			}
		})
	}

	t.Run("An extra trailing variable yields a parsing error", func(t *testing.T) {
		_, err := ParseCondition("a AND b c")
		if err == nil {
			t.Fatalf("Expected parsing error, got none")
		}

		parseErr, ok := err.(ErrConditionParse)
		if !ok {
			t.Fatalf("Expected ErrConditionParse, got %v", err)
		}
		if parseErr.Reason != ParseErrExtraTrailVar {
			t.Fatalf("Expected ErrConditionParse to contain a ParseErrExtraTrailVar reason")
		}
	})

	t.Run("A missing trailing variable yields a parsing error", func(t *testing.T) {
		_, err := ParseCondition("a AND")
		if err == nil {
			t.Fatalf("Expected parsing error, got none")
		}

		parseErr, ok := err.(ErrConditionParse)
		if !ok {
			t.Fatalf("Expected ErrConditionParse, got %v", err)
		}
		if parseErr.Reason != ParseErrIncompleteExpr {
			t.Fatalf("Expected ErrConditionParse to contain a ParseErrIncompleteExpr reason")
		}
	})

	t.Run("AND with missing LHS variable", func(t *testing.T) {
		_, err := ParseCondition(" AND b")
		if err == nil {
			t.Fatalf("Expected parse error")
		}

		parseErr, ok := err.(ErrConditionParse)
		if !ok {
			t.Fatalf("Expected ErrConditionParse, got %v", err)
		}
		if parseErr.Reason != ParseErrReasonMissingLHSVar {
			t.Fatalf("Expected ErrConditionParse to contain a ParseErrReasonMissingLHSVar reason")
		}
	})

	for _, tCase := range []struct {
		input map[string]bool
		want  bool
	}{
		{input: map[string]bool{"a": true, "b": true}, want: true},
		{input: map[string]bool{"a": true, "b": false}, want: false},
		{input: map[string]bool{"a": false, "b": true}, want: false},
		{input: map[string]bool{"a": false, "b": false}, want: false},
	} {
		t.Run("Simple AND condition", func(t *testing.T) {
			cond, _ := ParseCondition("a AND b")
			got, _ := cond(tCase.input)

			if got != tCase.want {
				t.Errorf("With %v, want %t but got %t", tCase.input, tCase.want, got)
			}
		})
	}

}
