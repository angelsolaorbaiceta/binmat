package signature

import "fmt"

// An orCondition is a binary operation that yields true if at least one operant is true.
type orCondition struct {
	lhs, rhs conditionExpr
}

func (c orCondition) append(expr conditionExpr) (conditionExpr, error) {
	return nil, nil
}

func (c *orCondition) apply(vars map[string]bool) bool {
	return c.lhs.apply(vars) || c.rhs.apply(vars)
}

func (c *orCondition) hasRhs() bool {
	return c.rhs != nil
}

func (c *orCondition) setLhs(expr conditionExpr) {
	c.lhs = expr
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
