package bexpr

import "fmt"

// An andCondition is a binary operation that yields true if both operands are true.
type andCondition struct {
	lhs, rhs conditionExpr
}

func (c *andCondition) apply(vars map[string]bool) (bool, *ErrMissingVarValue) {
	a, err := c.lhs.apply(vars)
	if err != nil {
		return false, err
	}
	b, err := c.rhs.apply(vars)
	if err != nil {
		return false, err
	}

	return a && b, nil
}

func (c *andCondition) precedence() int {
	return precedenceAnd
}

func (c *andCondition) hasRhs() bool {
	return c.rhs != nil
}

// setRhs sets the expression as the rhs operand or, if there is one already,
// appends the expression to it. Only variables and unary expressions can be
// set as the operand; binary operators are chained through appendToCondition.
func (c *andCondition) setRhs(expr conditionExpr) *errAppendToCond {
	if c.rhs == nil {
		if !isOperand(expr) {
			return &errAppendToCond{c, expr}
		}

		c.rhs = expr
		return nil
	}

	// Appending might restructure the rhs subtree (e.g. "?? OR a" + "AND" yields
	// "?? OR (a AND ??)"), so the returned parent replaces the current rhs.
	rhs, err := appendToCondition(c.rhs, expr)
	if err != nil {
		return err
	}

	c.rhs = rhs
	return nil
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
