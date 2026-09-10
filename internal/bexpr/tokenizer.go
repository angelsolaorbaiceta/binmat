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
//
// Variables and keywords must be separated from each other by whitespace or a
// parenthesis, so "a ANDb" is an error too instead of being read as "a AND b".
func tokenize(condition string) ([]string, *ErrConditionParse) {
	var (
		locs   = tokensRe.FindAllStringIndex(condition, -1)
		tokens = make([]string, 0, len(locs))
		// End of the previous token. The text between it and the start of the
		// next one must be whitespace only.
		prevEnd = 0
	)

	for i, loc := range locs {
		if err := checkBetweenTokens(condition, prevEnd, loc[0]); err != nil {
			return nil, err
		}

		if i > 0 && areGluedWords(condition, locs[i-1], loc) {
			return nil, gluedWordsErr(condition, locs, i-1)
		}

		tokens = append(tokens, condition[loc[0]:loc[1]])
		prevEnd = loc[1]
	}

	if err := checkBetweenTokens(condition, prevEnd, len(condition)); err != nil {
		return nil, err
	}

	return tokens, nil
}

// isWord returns whether the token is a variable name or a keyword, as opposed
// to a parenthesis. Words need whitespace or a parenthesis between them, while
// parentheses can touch anything.
func isWord(token string) bool {
	return token != tokenGroupStart && token != tokenGroupEnd
}

// areGluedWords returns whether the tokens at the prev and next locations are
// both words and follow each other with nothing in between.
func areGluedWords(condition string, prev, next []int) bool {
	return prev[1] == next[0] &&
		isWord(condition[prev[0]:prev[1]]) &&
		isWord(condition[next[0]:next[1]])
}

// gluedWordsErr reports the run of glued words that starts at the token with
// the given index as if it were a single word, with its 1-based column, so the
// error shows what was actually written (e.g. 'ANDb' rather than 'AND' and 'b').
func gluedWordsErr(condition string, locs [][]int, from int) *ErrConditionParse {
	to := from
	for to+1 < len(locs) && areGluedWords(condition, locs[to], locs[to+1]) {
		to++
	}

	start, end := locs[from][0], locs[to][1]

	return &ErrConditionParse{
		OffendingCond: condition,
		Reason:        ParseErrMissingWhitespace,
		Details: fmt.Sprintf(
			"'%s' at column %d",
			condition[start:end], utf8.RuneCountInString(condition[:start])+1,
		),
	}
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
