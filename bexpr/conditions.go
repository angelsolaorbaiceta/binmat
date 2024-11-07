package bexpr

import (
	"fmt"
)

// A conditionExpr is a boolean expression that can be evaluated given a map of
// variable names associated with their boolean value.
//
// All condition expressions must fall into one of the following three types:
//
//   - variable: a boolean variable
//   - unary: a boolean operation that operates on a single rhs operand
//   - binary: a boolean operation that operates on both lhs and rhs operands
type conditionExpr interface {
	fmt.Stringer

	// apply executes the boolean condition given the variable values in the map.
	// TODO: return error if a variable is missing to avoid searching for var names.
	apply(map[string]bool) bool
}

// A varConditionExpr is a boolean variable that can be evaluated as being either
// true or false.
type varConditionExpr interface {
	conditionExpr

	getName() string
}

// A binaryConditionExpr is a boolean expression that operates on two booleans,
// the lhs (left hand side) and rhs (right hand side).
type binaryConditionExpr interface {
	conditionExpr

	hasRhs() bool
	setRhs(expr conditionExpr)
	getRhs() conditionExpr

	hasLhs() bool
	setLhs(expr conditionExpr)
	getLhs() conditionExpr
}

// A unaryConditionExpr is a boolean expression that operates on just one boolean
// operand.
type unaryConditionExpr interface {
	conditionExpr

	hasOp() bool
	setOp(expr conditionExpr) *errAppendToCond
	getOp() conditionExpr
}

// isCondComplete returns whether the passed in condition has its operands
// defined (if should have them).
func isCondComplete(cond conditionExpr) bool {
	// A nil condition is considered "complete" (there isn't anything missing)
	if cond == nil {
		return true
	}

	switch typedCond := cond.(type) {
	case varConditionExpr:
		// A variable condition is always complete on its own
		return true

	case unaryConditionExpr:
		// A unary condition is complete if it has an operand
		return typedCond.hasOp()

	case binaryConditionExpr:
		// A binary condition if it has lhs and rhs
		return typedCond.hasLhs() && typedCond.hasRhs()
	}

	panic("Forgot to handle a condition type?")
}

// searchVarNames iterates the condition expression tree in a breadth-first search
// fashion, recording all variable names found.
func searchVarNames(expr conditionExpr) map[string]struct{} {
	var (
		varNames = make(map[string]struct{})
		q        = &queue[conditionExpr]{}
	)

	q.push(expr)

	for {
		item, ok := q.pop()
		if !ok {
			break
		}

		switch typedCond := item.(type) {
		case *varCondition:
			varNames[typedCond.varName] = struct{}{}

		case unaryConditionExpr:
			q.push(typedCond.getOp())

		case binaryConditionExpr:
			q.push(typedCond.getLhs())
			q.push(typedCond.getRhs())
		}
	}

	return varNames
}
