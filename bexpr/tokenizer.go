package bexpr

import (
	"regexp"
)

const (
	tokenAnd        = "AND"
	tokenOr         = "OR"
	tokenNot        = "NOT"
	tokenGroupStart = "("
	tokenGroupEnd   = ")"
)

var tokensRe = regexp.MustCompile(`[a-z0-9_]+|AND|OR|NOT|\(|\)`)

func tokenize(condition string) []string {
	tokens := make([]string, 0)
	for _, token := range tokensRe.FindAllString(condition, -1) {
		tokens = append(tokens, token)
	}

	return tokens
}
