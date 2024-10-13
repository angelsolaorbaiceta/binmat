package signature

import "fmt"

// A notCondition is a unary operation that yields the opposite value of the source expression.
type notCondition struct {
	expr conditionExpr
}

func (c notCondition) append(expr conditionExpr) (conditionExpr, error) {
	return nil, nil
}

func (c *notCondition) apply(vars map[string]bool) bool {
	return !c.expr.apply(vars)
}

func (c *notCondition) hasOp() bool {
	return c.expr != nil
}

func (c *notCondition) setOp(expr conditionExpr) {
	c.expr = expr
}

func (c *notCondition) String() string {
	var expr string
	if c.expr == nil {
		expr = "??"
	} else {
		expr = c.expr.String()
	}

	return fmt.Sprintf("NOT %s", expr)
}
