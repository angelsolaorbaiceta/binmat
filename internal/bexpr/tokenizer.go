package bexpr

import (
	"fmt"
	"regexp"
	"unicode"
	"unicode/utf8"
)

const (
	tokenAnd        = "AND"
	tokenOr         = "OR"
	tokenNot        = "NOT"
	tokenGroupStart = "("
	tokenGroupEnd   = ")"
)

// tokensRe recognizes the tokens of the language: variable names, the operator
// keywords and parentheses.
var tokensRe = regexp.MustCompile(`[a-z0-9_]+|AND|OR|NOT|\(|\)`)

// tokenize splits the condition into its tokens.
//
// Whitespace between tokens is skipped. Any other character that isn't part of
// a token is an error, reported with its 1-based column, so that a condition
// like "Foo AND bar" fails at the 'F' instead of quietly becoming "oo AND bar".
func tokenize(condition string) ([]string, *ErrConditionParse) {
	var (
		locs   = tokensRe.FindAllStringIndex(condition, -1)
		tokens = make([]string, 0, len(locs))
		// End of the previous token. The text between it and the start of the
		// next one must be whitespace only.
		prevEnd = 0
	)

	for _, loc := range locs {
		if err := checkBetweenTokens(condition, prevEnd, loc[0]); err != nil {
			return nil, err
		}

		tokens = append(tokens, condition[loc[0]:loc[1]])
		prevEnd = loc[1]
	}

	if err := checkBetweenTokens(condition, prevEnd, len(condition)); err != nil {
		return nil, err
	}

	return tokens, nil
}

// checkBetweenTokens makes sure the text of the condition between the from and
// to byte offsets, which lies between two tokens (or before the first one, or
// after the last one), contains nothing but whitespace.
// The first character that isn't whitespace is reported as unexpected.
func checkBetweenTokens(condition string, from, to int) *ErrConditionParse {
	for i, r := range condition[from:to] {
		if unicode.IsSpace(r) {
			continue
		}

		return &ErrConditionParse{
			OffendingCond: condition,
			Reason:        ParseErrUnexpectedChar,
			Details: fmt.Sprintf(
				"'%c' at column %d",
				r, utf8.RuneCountInString(condition[:from+i])+1,
			),
		}
	}

	return nil
}
