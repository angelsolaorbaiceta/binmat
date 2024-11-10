package bexpr

import (
	"fmt"
	"testing"
)

func TestTokenize(t *testing.T) {
	for _, tCase := range []struct {
		cond string
		want []string
	}{
		{cond: "a AND (b OR c)", want: []string{"a", "AND", "(", "b", "OR", "c", ")"}},
		{cond: "  a   AND (  b OR c )  ", want: []string{"a", "AND", "(", "b", "OR", "c", ")"}},
		{cond: "foo78 OR NOT bar23", want: []string{"foo78", "OR", "NOT", "bar23"}},
	} {

		t.Run(
			fmt.Sprintf("tokenize '%s'", tCase.cond),
			func(t *testing.T) {
				var (
					iter = makeTokenIter(tCase.cond)
					got  = iter.getAll()
				)

				if len(got) != len(tCase.want) {
					t.Fatalf("Want size %d, got %d", len(tCase.want), len(got))
				}
				for i, gotToken := range got {
					wantToken := tCase.want[i]
					if gotToken != wantToken {
						t.Fatalf("Want token '%s', got '%s'", wantToken, gotToken)
					}
				}
			})
	}
}
