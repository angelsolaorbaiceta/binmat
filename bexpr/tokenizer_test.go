package bexpr

import "testing"

func TestTokenize(t *testing.T) {
	t.Run("tokenize 'a AND (b OR c)'", func(t *testing.T) {
		var (
			iter = makeTokenIter("a AND (b OR c)")
			got  = iter.getAll()
			want = []string{"a", "AND", "(", "b", "OR", "c", ")"}
		)

		if len(got) != len(want) {
			t.Fatalf("Want size %d, got %d", len(want), len(got))
		}
		for i, gotToken := range got {
			wantToken := want[i]
			if gotToken != wantToken {
				t.Fatalf("Want token '%s', got '%s'", wantToken, gotToken)
			}
		}
	})

	t.Run("tokenize removes whitespace", func(t *testing.T) {
		var (
			iter = makeTokenIter("  a   AND (  b OR c )  ")
			got  = iter.getAll()
			want = []string{"a", "AND", "(", "b", "OR", "c", ")"}
		)

		if len(got) != len(want) {
			t.Fatalf("Want size %d, got %d", len(want), len(got))
		}
		for i, gotToken := range got {
			wantToken := want[i]
			if gotToken != wantToken {
				t.Fatalf("Want token '%s', got '%s'", wantToken, gotToken)
			}
		}
	})
}
