package signature

import "fmt"

type ErrSignatureReason string

const (
	ErrSigEmptyName      ErrSignatureReason = "the name can't be empty"
	ErrSigEmptyPatterns  ErrSignatureReason = "the patterns map can't be empty"
	ErrSigWrongCondition ErrSignatureReason = "the condition is either empty or invalid"
	ErrSigMissingPattern ErrSignatureReason = "missing pattern for condition"
)

// An ErrSignature is an error originating from an ill-formed signature.
type ErrSignature struct {
	cause  error
	reason ErrSignatureReason
}

func (e ErrSignature) Error() string {
	return fmt.Sprintf("Invalid signature (%s). Cause: %s", e.reason, e.cause)
}

func (e ErrSignature) Unwrap() error {
	return e.cause
}
