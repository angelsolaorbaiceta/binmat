package signature

import (
	"fmt"
	"regexp"
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
type condition func(map[string]bool) (bool, error)

const (
	condAnd = "AND"
	condOr  = "OR"
	condNot = "NOT"
)

var (
	tokenizerRe = regexp.MustCompile(`[\s()]+`)
	varNameRe   = regexp.MustCompile(`^[a-z0-9_]{1,16}$`)
)

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
func ParseCondition(condition string) (condition, error) {
	var (
		varNames = make(map[string]struct{})
		lhsVar   *varCondition
		expr     conditionExpr
	)

	for _, token := range tokenizerRe.Split(condition, -1) {
		switch token {
		case "":
			continue
		case condAnd:
			// Check there is a LHS variable to be used in the condition.
			if lhsVar == nil {
				return nil, ErrConditionParse{
					OffendingCond: condition,
					Reason:        ParseErrReasonMissingLHSVar,
					Details:       "AND requires a LHS variable",
				}
			}

			expr = &andCondition{lhs: lhsVar}
			lhsVar = nil
		case condOr:
			// do something
			panic("not implemented")
		case condNot:
			// do something
			panic("not implemented")
		default:
			// Check if token is a valid variable name.
			// Invalid variable names directly trigger an error, as they are unrecoverable.
			if !isValidVarName(token) {
				return nil, ErrInvalidVarName{OffendingName: token}
			}

			// Check that there isn't already a left-hand-side unused variable.
			if lhsVar != nil {
				return nil, ErrConditionParse{
					OffendingCond: condition,
					Reason:        ParseErrReasonContigVars,
					Details:       fmt.Sprintf("variables '%s' and '%s'", lhsVar, token),
				}
			}

			// If there is a binary expression without RHS, add it there
			if binExpr, ok := expr.(binaryConditionExpr); ok {
				if binExpr.hasRhs() {
					// TODO:
					// Error! What do we do with the new token?
				} else {
					binExpr.setRhs(&varCondition{varName: token})
				}
			} else {
				lhsVar = &varCondition{varName: token}
			}

			varNames[token] = struct{}{}
		}
	}

	// There might be an unused single lhs. Check if there is no expression, in
	// which case the lhs is the expression. If there was an expression in place
	// this is an error (an extra trailing variable).
	if lhsVar != nil {
		if expr == nil {
			expr = lhsVar
			lhsVar = nil
		} else {
			// TODO
			// Error: trailing extra variable
		}
	}

	cond := func(vars map[string]bool) (bool, error) {
		for name := range varNames {
			if ok := vars[name]; !ok {
				return false, ErrMissingVarValue{OffendingName: name}
			}
		}

		if expr == nil {
			return false, nil
		}

		return expr.apply(vars), nil
	}

	return cond, nil
}

// isValidVarName returns true if the name is a valid variable name.
//
// A variable name is valid if the following conditions are met:
//   - It's not empty
//   - It uses only lowercase letters, numbers, and underscores
//   - It doesn't contain spaces (this should be handled by the tokenizer)
//   - It's length is between 1 and 16 characters
func isValidVarName(name string) bool {
	return varNameRe.MatchString(name)
}