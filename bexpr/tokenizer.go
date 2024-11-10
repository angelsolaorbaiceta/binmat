package bexpr

import "regexp"

var tokensRe = regexp.MustCompile(`\s*([a-zA-Z]+|AND|OR|NOT|\(|\))\s*`)

const (
	tokenAnd        = "AND"
	tokenOr         = "OR"
	tokenNot        = "NOT"
	tokenGroupStart = "("
	tokenGroupEnd   = ")"
)

// A tokenIter is a token iterator.
// Instances maintain the state of the current token and whether the iteration
// finished.
type tokenIter struct {
	condition string
	tokens    []string
	nextIdx   int
	done      bool
}

func makeTokenIter(condition string) *tokenIter {
	var (
		tokens = tokensRe.FindAllString(condition, -1)
		done   = false
	)

	if len(tokens) < 1 {
		done = true
	}

	return &tokenIter{
		condition: condition,
		tokens:    tokens,
		nextIdx:   0,
		done:      done,
	}
}

func (iter *tokenIter) hasNext() bool {
	return !iter.done
}

func (iter *tokenIter) next() string {
	if !iter.hasNext() {
		panic("Called next() on exhausted token iterator")
	}

	next := iter.tokens[iter.nextIdx]
	iter.nextIdx += 1
	if iter.nextIdx >= len(iter.tokens) {
		iter.done = true
	}

	return next
}

func (iter *tokenIter) getAll() []string {
	tokens := make([]string, 0, len(iter.tokens))

	for iter.hasNext() {
		tokens = append(tokens, iter.next())
	}

	return tokens
}
