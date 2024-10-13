package signature

import "fmt"

// A notCondition is a unary operation that yields the opposite value of the source expression.
type notCondition struct {
	expr conditionExpr
}

// append to a unary condition follows these rules:
//   - A variable condition can be appended (e.g. "NOT a")
//   - Another unary condition can be appended (e.g. "NOT NOT")
//   - A binary condition can't be appended (e.g. "NOT AND" doesn't make sense)
//
// In any case, the unary condition is returned.
func (c *notCondition) append(expr conditionExpr) (conditionExpr, *errAppendToCond) {
	switch expr.(type) {
	case *varCondition, unaryConditionExpr:
		c.setOp(expr)
		return c, nil

	case binaryConditionExpr:
		return c, &errAppendToCond{
			Reason:  ParseErrBinaryAfterUnary,
			Details: fmt.Sprintf("%s can't be appended to %s", expr, c),
		}
	}

	panic("Forgot to handle a condition type?")
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
