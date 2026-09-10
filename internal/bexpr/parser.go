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
// NOT binds tighter than AND, which binds tighter than OR, so "a OR b AND c"
// reads as "a OR (b AND c)" and "NOT a AND b" as "(NOT a) AND b". Operators of
// the same precedence associate to the left: "a AND b AND c" reads as
// "(a AND b) AND c". Parentheses override this order.
//
// Examples of valid conditions:
//   - "a AND b"
//   - "a OR b"
//   - "a AND b AND c"
//   - "a AND (b OR c)"
//   - "(a OR b) AND c"
//   - "a AND NOT b"
//   - "NOT a AND b"
//   - "a AND NOT (b OR c)"
//
// If the expression can't be parsed, an ErrConditionParse error is returned.
func ParseCondition(condition string) (Condition, *ErrConditionParse) {
	p := parserFromCond(condition)
	expr, err := parse(p, 0)
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

// parse consumes tokens until they run out or, when parsing the contents of a
// group (depth > 0), until the parenthesis closing it. The depth is the number
// of parentheses open at this point, which is what tells a closing parenthesis
// that was never opened from one that ends the current group.
func parse(p *parser, depth int) (conditionExpr, *ErrConditionParse) {
	var (
		token string
		expr  conditionExpr
		err   *errAppendToCond
		// closed records whether the loop ended at the parenthesis closing the
		// current group, as opposed to running out of tokens.
		closed bool
	)

outerLoop:
	for p.hasNext() {
		switch token = p.next(); token {
		case "":
			continue

		case tokenGroupStart:
			groupExpr, parseErr := parse(p, depth+1)
			if parseErr != nil {
				return nil, parseErr
			}

			group := &groupCondition{expr: groupExpr}
			expr, err = appendToCondition(expr, group)
			if err != nil {
				return nil, err.toParseErr(p.cond)
			}

		case tokenGroupEnd:
			if depth == 0 {
				return nil, &ErrConditionParse{
					OffendingCond: p.cond,
					Reason:        ParseErrUnbalancedParens,
					Details:       "unexpected ')'",
				}
			}

			// The current group is considered complete, so it can be returned from here
			closed = true
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

	if depth > 0 && !closed {
		return nil, &ErrConditionParse{
			OffendingCond: p.cond,
			Reason:        ParseErrUnbalancedParens,
			Details:       "missing ')'",
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
