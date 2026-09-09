package bexpr

// appendToCondition appends the toAppend condition to the base condition and
// returns the top-level expression resulting from the append.
//
// The base might be nil, in which case appending to it yields the target condition.
// The toAppend condition, by contrast, can't be nil.
// Passing a nil toAppend condition returns an errAppendToCond error.
//
// If there is an error appending a condition, an errAppendToCond error is
// returned specifying the operands that failed to be appended.
func appendToCondition(baseCond, toAppend conditionExpr) (conditionExpr, *errAppendToCond) {
	if toAppend == nil {
		return nil, &errAppendToCond{baseCond, toAppend}
	}
	if baseCond == nil {
		return toAppend, nil
	}

	switch a := baseCond.(type) {
	case varConditionExpr:
		switch b := toAppend.(type) {
		// Only binary ops can be appended to variables (e.g. "a AND").
		// The variable is set as the lhs of the binary expression, and the latter
		// is returned as the parent.
		// If there was an lhs already, it returns an error.
		case binaryConditionExpr:
			if b.hasLhs() {
				return nil, &errAppendToCond{baseCond, toAppend}
			}

			b.setLhs(a)
			return b, nil
		}

	case unaryConditionExpr:
		switch b := toAppend.(type) {
		// Both variables and unary expressions can be appended to unary expressions
		// (e.g. "NOT a", "NOT NOT").
		// In both cases the first unary expression is returned as the parent.
		case varConditionExpr, unaryConditionExpr:
			if err := a.setOp(b); err != nil {
				return nil, err
			} else {
				return a, nil
			}
		}

	case binaryConditionExpr:
		switch b := toAppend.(type) {
		// Both variables and unary expressions can be appended to binary expressions
		// (e.g. "AND a", "AND NOT").
		// In both cases "b" is added as the rhs of the binary expression.
		// In both cases the binary expression is returned as the parent.
		// If there was a rhs already, it returns an error.
		case varConditionExpr, unaryConditionExpr:
			if err := a.setRhs(b); err != nil {
				return nil, err
			} else {
				return a, nil
			}
		}
	}

	return nil, &errAppendToCond{baseCond, toAppend}
}

// TODO: unify with logic in previous method to avoid duplicating the logic.
// canAppend determines if b can appended to a (a.append(b)).
func canAppend(a, b conditionExpr) bool {
	result := false

	switch a.(type) {
	// only binary ops can be appended to variables (e.g. "a AND")
	case varConditionExpr:
		switch b.(type) {
		case binaryConditionExpr:
			result = true
		}

	// both variables and unary expressions can be appended to unary and binary
	// expressions (e.g. "NOT a", "NOT NOT", "AND b", "AND NOT")
	case unaryConditionExpr, binaryConditionExpr:
		switch b.(type) {
		case varConditionExpr, unaryConditionExpr:
			result = true
		}
	}

	return result
}
