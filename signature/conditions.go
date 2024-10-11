package signature

import "fmt"

// A conditionExpr is a boolean expression that can be evaluated given a map of
// variable names associated with their boolean value.
type conditionExpr interface {
	apply(map[string]bool) bool
}

// A binaryConditionExpr is a boolean expression that operates on two booleans,
// the lhs (left hand side) and rhs (right hand side).
type binaryConditionExpr interface {
	hasRhs() bool
	setRhs(expr conditionExpr)
}

// A varCondition is a single boolean variable. The result of the condition is
// the value of the variable.
type varCondition struct {
	varName string
}

func (c varCondition) apply(vars map[string]bool) bool {
	varVal, ok := vars[c.varName]
	if !ok {
		panic("variable not found")
	}

	return varVal
}

func (c varCondition) String() string {
	return c.varName
}

// An andCondition is a binary operation that yields true if both operands are true.
type andCondition struct {
	lhs, rhs conditionExpr
}

func (c *andCondition) apply(vars map[string]bool) bool {
	return c.lhs.apply(vars) && c.rhs.apply(vars)
}

func (c *andCondition) hasRhs() bool {
	return c.rhs != nil
}

func (c *andCondition) setRhs(expr conditionExpr) {
	c.rhs = expr
}

func (c *andCondition) String() string {
	return fmt.Sprintf("%s AND %s", c.lhs, c.rhs)
}
