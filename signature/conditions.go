package signature

import "fmt"

// A conditionExpr is a boolean expression that can be evaluated given a map of
// variable names associated with their boolean value.
type conditionExpr interface {
	fmt.Stringer

	apply(map[string]bool) bool
}

// A binaryConditionExpr is a boolean expression that operates on two booleans,
// the lhs (left hand side) and rhs (right hand side).
type binaryConditionExpr interface {
	conditionExpr

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
	var lhs, rhs string

	if c.lhs == nil {
		lhs = "??"
	} else {
		lhs = c.lhs.String()
	}

	if c.rhs == nil {
		rhs = "??"
	} else {
		rhs = c.rhs.String()
	}

	return fmt.Sprintf("%s AND %s", lhs, rhs)
}

// An orCondition is a binary operation that yields true if at least one operant is true.
type orCondition struct {
	lhs, rhs conditionExpr
}

func (c *orCondition) apply(vars map[string]bool) bool {
	return c.lhs.apply(vars) || c.rhs.apply(vars)
}

func (c *orCondition) hasRhs() bool {
	return c.rhs != nil
}

func (c *orCondition) setRhs(expr conditionExpr) {
	c.rhs = expr
}

func (c *orCondition) String() string {
	var lhs, rhs string

	if c.lhs == nil {
		lhs = "??"
	} else {
		lhs = c.lhs.String()
	}

	if c.rhs == nil {
		rhs = "??"
	} else {
		rhs = c.rhs.String()
	}

	return fmt.Sprintf("%s OR %s", lhs, rhs)
}
