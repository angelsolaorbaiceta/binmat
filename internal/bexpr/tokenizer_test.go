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
		{cond: "", want: []string{}},
		{cond: "a AND (b OR c)", want: []string{"a", "AND", "(", "b", "OR", "c", ")"}},
		{cond: "  a   AND (  b OR c )  ", want: []string{"a", "AND", "(", "b", "OR", "c", ")"}},
		{cond: "foo78 OR NOT bar23", want: []string{"foo78", "OR", "NOT", "bar23"}},
		{cond: "a\tAND\n(b OR\tc)", want: []string{"a", "AND", "(", "b", "OR", "c", ")"}},
		{cond: "a AND(b OR c)", want: []string{"a", "AND", "(", "b", "OR", "c", ")"}},
		{cond: "NOT(a)AND(b)", want: []string{"NOT", "(", "a", ")", "AND", "(", "b", ")"}},
	} {

		t.Run(
			fmt.Sprintf("tokenize '%s'", tCase.cond),
			func(t *testing.T) {
				got, err := tokenize(tCase.cond)
				assert.Nil(t, err)
				assert.Equal(t, tCase.want, got)
			})
	}
}

// TestTokenizeUnexpectedCharacters makes sure characters that aren't part of
// the language are reported, with their position, instead of silently dropped.
func TestTokenizeUnexpectedCharacters(t *testing.T) {
	for _, tCase := range []struct {
		cond   string
		char   string
		column int
	}{
		{cond: "a & b", char: "&", column: 3},
		{cond: "Foo AND bar", char: "F", column: 1},
		{cond: "a AND b-c", char: "-", column: 8},
		{cond: "ANDY", char: "Y", column: 4},
		{cond: "a OR !b", char: "!", column: 6},
		{cond: "a AND b.", char: ".", column: 8},
		{cond: "(a OR b) AND c;", char: ";", column: 15},
		// Multi-byte character: the column counts characters, not bytes.
		{cond: "a \u2013 b", char: "\u2013", column: 3},
	} {
		t.Run(
			fmt.Sprintf("tokenize '%s'", tCase.cond),
			func(t *testing.T) {
				tokens, err := tokenize(tCase.cond)

				assert.Nil(t, tokens)
				if assert.NotNil(t, err) {
					assert.Equal(t, ParseErrUnexpectedChar, err.Reason)
					assert.Equal(t, tCase.cond, err.OffendingCond)
					assert.Equal(t, fmt.Sprintf("'%s' at column %d", tCase.char, tCase.column), err.Details)
				}
			})
	}
}

// TestTokenizeMissingWhitespace makes sure variables and keywords can't be
// glued together: two consecutive word tokens must be separated by whitespace
// or a parenthesis. The glued run is reported as a whole, with its position.
func TestTokenizeMissingWhitespace(t *testing.T) {
	for _, tCase := range []struct {
		cond   string
		word   string
		column int
	}{
		{cond: "a ANDb", word: "ANDb", column: 3},
		{cond: "aAND b", word: "aAND", column: 1},
		{cond: "aANDb", word: "aANDb", column: 1},
		{cond: "NOTa", word: "NOTa", column: 1},
		{cond: "NOT NOTa", word: "NOTa", column: 5},
		{cond: "a ORb", word: "ORb", column: 3},
		{cond: "NOTNOT a", word: "NOTNOT", column: 1},
		{cond: "a ANDNOT b", word: "ANDNOT", column: 3},
		{cond: "(a ORb) AND c", word: "ORb", column: 4},
		{cond: "a AND\tbOR c", word: "bOR", column: 7},
	} {
		t.Run(
			fmt.Sprintf("tokenize '%s'", tCase.cond),
			func(t *testing.T) {
				tokens, err := tokenize(tCase.cond)

				assert.Nil(t, tokens)
				if assert.NotNil(t, err) {
					assert.Equal(t, ParseErrMissingWhitespace, err.Reason)
					assert.Equal(t, tCase.cond, err.OffendingCond)
					assert.Equal(t, fmt.Sprintf("'%s' at column %d", tCase.word, tCase.column), err.Details)
				}
			})
	}
}
