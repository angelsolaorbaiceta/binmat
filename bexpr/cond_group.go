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

func (c *groupCondition) setOp(expr conditionExpr) {
	c.expr = expr
}

func (c *groupCondition) getOp() conditionExpr {
	return c.expr
}

func (c *groupCondition) String() string {
	var expr string
	if c.expr == nil {
		expr = "??"
	} else {
		c.expr.String()
	}

	return fmt.Sprintf("(%s)", expr)
}
