package signature

import "fmt"

// An andCondition is a binary operation that yields true if both operands are true.
type andCondition struct {
	lhs, rhs conditionExpr
}

func (c *andCondition) append(expr conditionExpr) (conditionExpr, *errAppendToCond) {
	switch typedExpr := expr.(type) {
	case *varCondition, unaryConditionExpr:
		c.setRhs(typedExpr)
		return c, nil

	case binaryConditionExpr:
		return c, &errAppendToCond{
			Reason:  ParseErrContigBinary,
			Details: fmt.Sprintf("can't append %s to %s", typedExpr, c),
		}
	}

	panic("Forgot to handle a condition type?")
}

func (c *andCondition) apply(vars map[string]bool) bool {
	return c.lhs.apply(vars) && c.rhs.apply(vars)
}

func (c *andCondition) hasRhs() bool {
	return c.rhs != nil
}

func (c *andCondition) setLhs(expr conditionExpr) {
	c.lhs = expr
}

func (c *andCondition) setRhs(expr conditionExpr) {
	if c.rhs == nil {
		c.rhs = expr
	} else {
		switch exprType := expr.(type) {
		case binaryConditionExpr:
			exprType.setRhs(expr)
		case unaryConditionExpr:
			exprType.setOp(expr)
		}
	}
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
