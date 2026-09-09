package bexpr

import "fmt"

// A notCondition is a unary operation that yields the opposite value of the source expression.
type notCondition struct {
	op conditionExpr
}

func (c *notCondition) apply(vars map[string]bool) (bool, *ErrMissingVarValue) {
	a, err := c.op.apply(vars)
	if err != nil {
		return false, err
	}

	return !a, nil
}

// TODO: this shouldn't be necessary
func (c *notCondition) hasOp() bool {
	return c.op != nil
}

// setOp sets either a variable or another unary expression as the operand
// for this unary expression.
// Binary expressions can't be added as operands, so an errAppendToCond is
// returned in this case.
func (c *notCondition) setOp(expr conditionExpr) *errAppendToCond {
	if !canAppend(c, expr) {
		return &errAppendToCond{c, expr}
	}

	if c.op == nil {
		c.op = expr
		return nil
	}

	_, err := appendToCondition(c.op, expr)

	return err
}

func (c *notCondition) getOp() conditionExpr {
	return c.op
}

func (c *notCondition) String() string {
	var expr string
	if c.op == nil {
		expr = "??"
	} else {
		expr = c.op.String()
	}

	return fmt.Sprintf("NOT %s", expr)
}
