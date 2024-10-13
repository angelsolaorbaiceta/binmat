package signature

import "fmt"

// An orCondition is a binary operation that yields true if at least one operant is true.
type orCondition struct {
	lhs, rhs conditionExpr
}

// append to a binary operation follows these rules:
//   - A variable condition can be appended (e.g. "?? AND b")
//   - A unary condition can be appended (e.g. "?? AND NOT ??")
//   - A binary condition can't be appended (e.g. "?? AND OR ??")
//
// In every case, this binary condition (the receiver of the method) is returned.
func (c *orCondition) append(expr conditionExpr) (conditionExpr, *errAppendToCond) {
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
