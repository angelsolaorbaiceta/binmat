package bexpr

import "fmt"

// An orCondition is a binary operation that yields true if at least one operant is true.
type orCondition struct {
	lhs, rhs conditionExpr
}

func (c *orCondition) apply(vars map[string]bool) (bool, *ErrMissingVarValue) {
	a, err := c.lhs.apply(vars)
	if err != nil {
		return false, err
	}
	b, err := c.rhs.apply(vars)
	if err != nil {
		return false, err
	}

	return a || b, nil
}

func (c *orCondition) hasRhs() bool {
	return c.rhs != nil
}

func (c *orCondition) setRhs(expr conditionExpr) *errAppendToCond {
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

func (c *orCondition) getRhs() conditionExpr {
	return c.rhs
}

func (c *orCondition) hasLhs() bool {
	return c.lhs != nil
}

func (c *orCondition) setLhs(expr conditionExpr) {
	c.lhs = expr
}

func (c *orCondition) getLhs() conditionExpr {
	return c.lhs
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
