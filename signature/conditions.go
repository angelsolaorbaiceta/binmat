package signature

import (
	"fmt"
)

// A conditionExpr is a boolean expression that can be evaluated given a map of
// variable names associated with their boolean value.
type conditionExpr interface {
	fmt.Stringer

	// append appends—when possible—the given condition to this condition.
	// Appending means to add as a left-hand-side (lhs) operand.
	//
	// For example:
	//
	// 	a.append(b)
	//
	// Means to add "a" as the lhs of "b".
	//
	// It returns the top-level condition, whether that is this condition or the
	// condition that's been appended.
	// For example, if "a" is a variable and "b" is an AND operation, a.append(b)
	// returns b, the AND operation.
	//
	// An error is returned in the cases where it doesn't make sense to append
	// the condition.
	// Implementors should return the target expression beside the error, when
	// there is one.
	// In other words, errors are returned together with the condition upon which
	// the method is called.
	append(expr conditionExpr) (conditionExpr, *errAppendToCond)

	// apply executes the boolean condition given the variable values in the map.
	apply(map[string]bool) bool
}

// A binaryConditionExpr is a boolean expression that operates on two booleans,
// the lhs (left hand side) and rhs (right hand side).
type binaryConditionExpr interface {
	conditionExpr

	hasRhs() bool
	setLhs(expr conditionExpr)
	setRhs(expr conditionExpr)
}

// A unaryConditionExpr is a boolean expression that operates on just one boolean
// operand.
type unaryConditionExpr interface {
	conditionExpr

	hasOp() bool
	setOp(expr conditionExpr)
}

// appendToCondition appends the toAppend condition to the base condition.
// Depending on the nature of the base condition, adding a condition to it means
// a different thing:
//   - A variable condition can't be added to another variable condition.
//   - A condition added to a unary condition adds it as its operand, overwriting
//     the previous operand if there was any.
//   - A condition added to a binary condition adds it as its right-hand-side
//     operand, overwriting the previous operand if there was any.
//
// The base might be nil, in which case appending to it yields the target condition.
// The toAppend condition, by contrast, can't be nil.
// Passing a nil toAppend condition returns an errAppendToCond error.
//
// If there is an error appending a condition, an errAppendToCond error is
// returned specifying the reason why the operation failed.
func appendToCondition(baseCond, toAppend conditionExpr) (conditionExpr, *errAppendToCond) {
	if toAppend == nil {
		return nil, &errAppendToCond{
			Reason:  ParseErrLogicError,
			Details: "the condition to be appended is nil",
		}
	}
	if baseCond == nil {
		return toAppend, nil
	}

	return baseCond.append(toAppend)
}
