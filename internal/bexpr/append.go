package bexpr

// appendToCondition appends the toAppend condition to the base condition and
// returns the top-level expression resulting from the append.
//
// The base might be nil, in which case appending to it yields the target condition.
// The toAppend condition, by contrast, can't be nil.
// Passing a nil toAppend condition returns an errAppendToCond error.
//
// Appending a variable or a unary expression fills the next empty operand
// (e.g. "a AND" + "b", or "NOT" + "a"). Appending a binary operator to a
// complete expression chains them (e.g. "a AND b" + "OR", "NOT a" + "AND", or
// "(a OR b)" + "AND") following the operators' precedence: the operator that
// binds tightest takes the operands closest to it, and operators of equal
// precedence associate to the left.
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
		// Only binary ops can be appended to variables (e.g. "a AND").
		// The variable is set as the lhs of the binary expression, and the latter
		// is returned as the parent.
		if b, ok := toAppend.(binaryConditionExpr); ok {
			return chainBinary(a, b)
		}

	case unaryConditionExpr:
		switch b := toAppend.(type) {
		// Both variables and unary expressions can be appended to unary expressions
		// (e.g. "NOT a", "NOT NOT").
		// In both cases the first unary expression is returned as the parent.
		case varConditionExpr, unaryConditionExpr:
			if err := a.setOp(b); err != nil {
				return nil, err
			}
			return a, nil

		// Binary ops can be appended to complete unary expressions (e.g. "NOT a AND",
		// "(a OR b) AND"). NOT binds tighter than any binary operator and groups are
		// self-contained, so the unary expression becomes the lhs of the binary one,
		// which is returned as the parent.
		case binaryConditionExpr:
			if !isCondComplete(a) {
				return nil, &errAppendToCond{baseCond, toAppend}
			}
			return chainBinary(a, b)
		}

	case binaryConditionExpr:
		switch b := toAppend.(type) {
		// Both variables and unary expressions can be appended to binary expressions
		// (e.g. "AND a", "AND NOT").
		// In both cases "b" is added as the rhs of the binary expression.
		// In both cases the binary expression is returned as the parent.
		// If there was a rhs already, it's appended to the rhs.
		case varConditionExpr, unaryConditionExpr:
			if err := a.setRhs(b); err != nil {
				return nil, err
			}
			return a, nil

		// Binary ops can be appended to complete binary expressions (e.g. "a AND b OR",
		// "a OR b AND").
		case binaryConditionExpr:
			if !isCondComplete(a) {
				return nil, &errAppendToCond{baseCond, toAppend}
			}

			if b.precedence() > a.precedence() {
				// The new operator binds tighter, so it takes the base's rhs as its
				// own lhs: "a OR b" + "AND" yields "a OR (b AND ??)".
				if err := a.setRhs(b); err != nil {
					return nil, err
				}
				return a, nil
			}

			// Same or lower precedence: operators associate to the left, so the whole
			// base becomes the lhs: "a AND b" + "OR" yields "(a AND b) OR ??".
			return chainBinary(a, b)
		}
	}

	return nil, &errAppendToCond{baseCond, toAppend}
}

// chainBinary sets the lhs expression as the left operand of the binary
// operator and returns the operator as the new parent expression.
// If the operator already has a lhs, an errAppendToCond error is returned.
func chainBinary(lhs conditionExpr, op binaryConditionExpr) (conditionExpr, *errAppendToCond) {
	if op.hasLhs() {
		return nil, &errAppendToCond{lhs, op}
	}

	op.setLhs(lhs)
	return op, nil
}

// isOperand returns whether the expression can be set as the operand of a
// unary expression or as the empty rhs of a binary one. Variables and unary
// expressions can; binary operators can't, as they are chained through
// appendToCondition instead, which accounts for their precedence.
func isOperand(expr conditionExpr) bool {
	switch expr.(type) {
	case varConditionExpr, unaryConditionExpr:
		return true
	}

	return false
}
