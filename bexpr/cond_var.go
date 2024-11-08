package bexpr

import (
	"regexp"
)

var varNameRe = regexp.MustCompile(`^[a-z0-9_]{1,16}$`)

// isValidVarName returns true if the name is a valid variable name.
//
// A variable name is valid if the following conditions are met:
//   - It's not empty
//   - It uses only lowercase letters, numbers, and underscores
//   - It doesn't contain spaces (this should be handled by the tokenizer)
//   - It's length is between 1 and 16 characters
func isValidVarName(name string) bool {
	return varNameRe.MatchString(name)
}

// A varCondition is a single boolean variable. The result of the condition is
// the value of the variable.
// This condition doesn't have operands: neither lhs, nor rhs.
type varCondition struct {
	varName string
}

// apply simply returns the value of the variable as defined in the passed in map.
func (c *varCondition) apply(vars map[string]bool) (bool, *ErrMissingVarValue) {
	varVal, ok := vars[c.varName]
	if !ok {
		return false, &ErrMissingVarValue{OffendingName: c.varName}
	}

	return varVal, nil
}

func (c *varCondition) getName() string {
	return c.varName
}

func (c *varCondition) String() string {
	return c.varName
}
