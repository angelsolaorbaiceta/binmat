package bexpr

import "fmt"

// ErrMissingVarValue is returned from the condition function when the map of
// variable values doesn't contain the value of a variable that's part of the
// condition expression.
type ErrMissingVarValue struct {
	OffendingName string
}

func (e ErrMissingVarValue) Error() string {
	return fmt.Sprintf("missing value for variable with name: '%s'", e.OffendingName)
}

type ParseErrorReason string

const (
	ParseErrInvalidAppend  ParseErrorReason = "invalid append attempt"
	ParseErrInvalidVarName ParseErrorReason = "invalid variable name"
	ParseErrIncompleteExpr ParseErrorReason = "incomplete binary operation"
)

// ErrConditionParse is returned when a condition expression can't be parsed due
// to some kind of syntax error, as detailed by the Reason field.
type ErrConditionParse struct {
	OffendingCond string
	Reason        ParseErrorReason
	Details       string
}

func (e ErrConditionParse) Error() string {
	return fmt.Sprintf(
		"can't parse the expression '%s'. Reason: %s (%s)",
		e.OffendingCond, e.Reason, e.Details,
	)
}

// errAppendToCond is the error returned by the appendToCondition() function
// when there is an error appending an expression to another expression.
type errAppendToCond struct {
	a, b conditionExpr
}

func (e errAppendToCond) Error() string {
	return fmt.Sprintf(
		"can't append %s and %s", e.a, e.b,
	)
}

func (e errAppendToCond) toParseErr(offendingCondition string) *ErrConditionParse {
	return &ErrConditionParse{
		OffendingCond: offendingCondition,
		Reason:        ParseErrInvalidAppend,
		Details:       e.a.String(),
	}
}
