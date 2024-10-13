package signature

import "fmt"

// A varCondition is a single boolean variable. The result of the condition is
// the value of the variable.
// This condition doesn't have operands: neither lhs, nor rhs.
type varCondition struct {
	varName string
}

// append in a variable condition follows these rules:
//   - Can't append to another variable: two contiguous variables don't make sense.
//   - Can't append to a unary condition, as these don't have a lhs.
//   - Appending to a binary condition is allowed. The returned value is the
//     binary operation, not the variable.
func (c *varCondition) append(expr conditionExpr) (conditionExpr, error) {
	switch typedExpr := expr.(type) {
	case *varCondition:
		return c, errAppendToCond{
			Reason:  ParseErrReasonContigVars,
			Details: fmt.Sprintf("'%s' and '%s'", c, typedExpr),
		}

	case unaryConditionExpr:
		return c, errAppendToCond{
			Reason:  ParseErrLHSOnUnary,
			Details: fmt.Sprintf("%s doesn't accept %s as lhs", typedExpr, c),
		}

	case binaryConditionExpr:
		typedExpr.setLhs(c)
		return typedExpr, nil
	}

	return c, nil
}

// apply simply returns the value of the variable as defined in the passed in map.
// Panics if the map doesn't include the required variable, thus, the presence
// of it should be verified before passing it here.
func (c *varCondition) apply(vars map[string]bool) bool {
	varVal, ok := vars[c.varName]
	if !ok {
		panic(fmt.Sprintf("'%s' variable not found in %v", c.varName, vars))
	}

	return varVal
}

func (c *varCondition) String() string {
	return c.varName
}
