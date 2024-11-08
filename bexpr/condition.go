package bexpr

import (
	"fmt"
)

// A condition is a function that takes a map of pattern names and returns true
// if the condition is met.
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
// All variables in the condition must be in the patterns map, otherwise an error
// will be returned.
//
// Here's a list of the possible errors the condition function can return:
//   - ErrMissingVarValue: when a variable in the expression isn't provided in the argument.
type condition func(map[string]bool) (bool, *ErrMissingVarValue)

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
func ParseCondition(condition string) (condition, *ErrConditionParse) {
	iter := makeTokenIter(condition)
	expr, err := parse(iter)
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

func parse(iter *tokenIter) (conditionExpr, *ErrConditionParse) {
	var (
		token string
		expr  conditionExpr
		err   *errAppendToCond
	)

outerLoop:
	for iter.hasNext() {
		switch token = iter.next(); token {
		case "":
			continue

		case tokenGroupStart:
			// Parse the entire group (recursively) and push it to the stack
			group, parseErr := parse(iter)
			if parseErr != nil {
				return nil, parseErr
			}

			expr, err = appendToCondition(expr, group)
			if err != nil {
				return nil, err.toParseErr(iter.condition)
			}

		case tokenGroupEnd:
			// The current group is considered complete, so it can be returned from here
			break outerLoop

		case tokenNot:
			expr, err = appendToCondition(expr, &notCondition{})
			if err != nil {
				return nil, err.toParseErr(iter.condition)
			}

		case tokenAnd:
			expr, err = appendToCondition(expr, &andCondition{})
			if err != nil {
				return nil, err.toParseErr(iter.condition)
			}

		case tokenOr:
			expr, err = appendToCondition(expr, &orCondition{})
			if err != nil {
				return nil, err.toParseErr(iter.condition)
			}

		default:
			// Check if token is a valid variable name
			// Invalid variable names directly trigger an error, as they are unrecoverable
			if isValidVarName(token) {
				expr, err = appendToCondition(expr, &varCondition{varName: token})
				if err != nil {
					return nil, err.toParseErr(iter.condition)
				}
			} else {
				return nil, &ErrConditionParse{
					OffendingCond: iter.condition,
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
			OffendingCond: iter.condition,
			Reason:        ParseErrIncompleteExpr,
			Details:       fmt.Sprintf("'%s'", expr),
		}
	}

	return expr, nil
}
