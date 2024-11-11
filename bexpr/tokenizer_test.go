package bexpr

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
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

				assert.Equal(t, tCase.want, got)
			})
	}
}
