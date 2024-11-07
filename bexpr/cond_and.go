package bexpr

import "fmt"

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

func (c *andCondition) setRhs(expr conditionExpr) *errAppendToCond {
	if !canAppend(c, expr) {
		return &errAppendToCond{c, expr}
	}

	if c.rhs == nil {
		c.rhs = expr
		return nil
	}

	_, err := appendToCondition(c.rhs, expr)

	return err
}

func (c *andCondition) getRhs() conditionExpr {
	return c.rhs
}

func (c *andCondition) hasLhs() bool {
	return c.lhs != nil
}

func (c *andCondition) setLhs(expr conditionExpr) {
	c.lhs = expr
}

func (c *andCondition) getLhs() conditionExpr {
	return c.lhs
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
