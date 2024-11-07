package bexpr

import "fmt"

type groupCondition struct {
	expr conditionExpr
}

func (c *groupCondition) apply(vars map[string]bool) bool {
	return c.expr.apply(vars)
}

func (c *groupCondition) hasOp() bool {
	return c.expr != nil
}

// setOp sets either a variable or another unary expression as the operand
// for this unary expression.
// Binary expressions can't be added as operands, so an errAppendToCond is
// returned in this case.
func (c *groupCondition) setOp(expr conditionExpr) *errAppendToCond {
	if !canAppend(c, expr) {
		return &errAppendToCond{c, expr}
	}

	if c.expr == nil {
		c.expr = expr
		return nil
	}

	_, err := appendToCondition(c.expr, expr)

	return err
}

func (c *groupCondition) getOp() conditionExpr {
	return c.expr
}

func (c *groupCondition) String() string {
	var expr string
	if c.expr == nil {
		expr = "??"
	} else {
		expr = c.expr.String()
	}

	return fmt.Sprintf("(%s)", expr)
}
