package bexpr

import (
	"fmt"
)

// A Condition is a function that takes a map of pattern names and returns true
// if the Condition is met.
//
// Example:
//
//	cond, err := ParseCondition("a AND (b OR c)")
//	if err != nil {
//		panic(err)
//	}
//
//	cond(map[string]bool{"a": true, "b": false, "c": true}) // false
//	cond(map[string]bool{"a": true, "b": true, "c": false}) // true
//
// All variables in the Condition must be in the patterns map, otherwise an error
// will be returned.
//
// Here's a list of the possible errors the Condition function can return:
//   - ErrMissingVarValue: when a variable in the expression isn't provided in the argument.
type Condition func(map[string]bool) (bool, *ErrMissingVarValue)

// parser holds the state while parsing a condition.
type parser struct {
	tokens []string
	idx    int
	cond   string // original condition for error messages
}

func parserFromCond(condition string) *parser {
	tokens := tokenize(condition)
	return &parser{
		tokens: tokens,
		idx:    0,
		cond:   condition,
	}
}

func (p *parser) hasNext() bool {
	return p.idx < len(p.tokens)
}

func (p *parser) next() string {
	if !p.hasNext() {
		panic("unexpected end of tokens")
	}
	s := p.tokens[p.idx]
	p.idx++
	return s
}

// ParseCondition parses a condition string and returns a condition function.
//
// A condition is made of variable names and operators.
// Variables are always lowercase letters, numbers, and underscores with a length
// between 1 and 16 characters.
//
// Examples of valid variable names:
//   - "a"
//   - "b1"
//   - "c_2"
//   - "foo_bar"
//
// Operators are:
//   - AND
//   - OR
//   - NOT
//   - Parentheses (for grouping)
//
// Examples of valid conditions:
//   - "a AND b"
//   - "a OR b"
//   - "a AND (b OR c)"
//   - "a AND NOT b"
//   - "a AND NOT (b OR c)"
//
// If the expression can't be parsed, an ErrConditionParse error is returned.
func ParseCondition(condition string) (Condition, *ErrConditionParse) {
	p := parserFromCond(condition)
	expr, err := parse(p)
	if err != nil {
		return nil, err
	}

	cond := func(vars map[string]bool) (bool, *ErrMissingVarValue) {
		if expr == nil {
			return false, nil
		}

		return expr.apply(vars)
	}

	return cond, nil
}

func parse(p *parser) (conditionExpr, *ErrConditionParse) {
	var (
		token string
		expr  conditionExpr
		err   *errAppendToCond
	)

outerLoop:
	for p.hasNext() {
		switch token = p.next(); token {
		case "":
			continue

		case tokenGroupStart:
			groupExpr, parseErr := parse(p)
			if parseErr != nil {
				return nil, parseErr
			}

			group := &groupCondition{expr: groupExpr}
			expr, err = appendToCondition(expr, group)
			if err != nil {
				return nil, err.toParseErr(p.cond)
			}

		case tokenGroupEnd:
			// The current group is considered complete, so it can be returned from here
			break outerLoop

		case tokenNot:
			expr, err = appendToCondition(expr, &notCondition{})
			if err != nil {
				return nil, err.toParseErr(p.cond)
			}

		case tokenAnd:
			expr, err = appendToCondition(expr, &andCondition{})
			if err != nil {
				return nil, err.toParseErr(p.cond)
			}

		case tokenOr:
			expr, err = appendToCondition(expr, &orCondition{})
			if err != nil {
				return nil, err.toParseErr(p.cond)
			}

		default:
			// Check if token is a valid variable name
			// Invalid variable names directly trigger an error, as they are unrecoverable
			if IsValidVarName(token) {
				expr, err = appendToCondition(expr, &varCondition{varName: token})
				if err != nil {
					return nil, err.toParseErr(p.cond)
				}
			} else {
				return nil, &ErrConditionParse{
					OffendingCond: p.cond,
					Reason:        ParseErrInvalidVarName,
					Details: fmt.Sprintf(
						"'%s' must contain between 1 and 16 lowercase letters, numbers and underscores",
						token,
					),
				}
			}
		}
	}

	if !isCondComplete(expr) {
		return nil, &ErrConditionParse{
			OffendingCond: p.cond,
			Reason:        ParseErrIncompleteExpr,
			Details:       fmt.Sprintf("'%s'", expr),
		}
	}

	return expr, nil
}
