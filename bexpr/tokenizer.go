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
	return tokensRe.FindAllString(condition, -1)
}
