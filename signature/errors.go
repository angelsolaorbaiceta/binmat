package signature

import "fmt"

// ErrInvalidVarName is returned when a variable name is invalid.
type ErrInvalidVarName struct {
	OffendingName string
}

func (e ErrInvalidVarName) Error() string {
	return fmt.Sprintf("invalid variable name: '%s'", e.OffendingName)
}

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
	ParseErrLogicError          ParseErrorReason = "called the parsing logic incorrectly"
	ParseErrReasonContigVars    ParseErrorReason = "found two contiguous variables"
	ParseErrUnaryNoLHS          ParseErrorReason = "unary operations don't have a left-hand-side operand"
	ParseErrReasonMissingLHSVar ParseErrorReason = "missing left-hand-side operand for condition"
	ParseErrExtraTrailVar       ParseErrorReason = "extra trailing variable"
	ParseErrIncompleteExpr      ParseErrorReason = "incomplete binary operation"
	ParseErrLHSOnUnary          ParseErrorReason = "unary operation doesn't expect a LHS"
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
// when there is an error appending a condition to another condition.
type errAppendToCond struct {
	Reason  ParseErrorReason
	Details string
}

func (e errAppendToCond) Error() string {
	return fmt.Sprintf(
		"can't append the condition. Reason: %s (%s)",
		e.Reason, e.Details,
	)
}
