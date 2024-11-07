package bexpr

// import (
// 	"testing"
// )

// func TestConditions(t *testing.T) {
// 	t.Run("Adding a target to a nil source condition", func(t *testing.T) {
// 		var (
// 			a conditionExpr = nil
// 			b               = &varCondition{varName: "b"}
// 		)

// 		res, _ := appendToCondition(a, b)

// 		if res != b {
// 			t.Errorf("Expected the target condition to be set to the source")
// 		}
// 	})

// 	t.Run("A nil target condition is an error", func(t *testing.T) {
// 		var (
// 			a               = &varCondition{varName: "a"}
// 			b conditionExpr = nil
// 		)

// 		_, err := appendToCondition(a, b)

// 		if err == nil {
// 			t.Fatalf("Expected error")
// 		}
// 		if err.Reason != ParseErrLogicError {
// 			t.Fatalf("Expected errAppendToCond to contain a ParseLogicError reason")
// 		}
// 	})

// 	t.Run("Can't append a variable to a variable", func(t *testing.T) {
// 		var (
// 			a = &varCondition{varName: "a"}
// 			b = &varCondition{varName: "b"}
// 		)

// 		_, err := appendToCondition(a, b)

// 		if err == nil {
// 			t.Fatalf("Expected error")
// 		}
// 		if err.Reason != ParseErrContigVars {
// 			t.Fatalf("Expected errAppendToCond to contain a ReasonContigVars reason")
// 		}
// 	})

// 	t.Run("Can't append a unary condition to a variable", func(t *testing.T) {
// 		var (
// 			a = &varCondition{varName: "a"}
// 			b = &notCondition{}
// 		)

// 		_, err := appendToCondition(a, b)

// 		if err == nil {
// 			t.Fatalf("Expected error")
// 		}
// 		if err.Reason != ParseErrLHSOnUnary {
// 			t.Fatalf("Expected errAppendToCond to contain ParseErrLHSOnUnary a reason")
// 		}
// 	})

// 	t.Run("variable + AND", func(t *testing.T) {
// 		var (
// 			a    = &varCondition{varName: "a"}
// 			b    = &andCondition{}
// 			want = "a AND ??"
// 		)

// 		got, _ := appendToCondition(a, b)

// 		if got.String() != want {
// 			t.Fatalf("Want '%s', got '%s'", want, got.String())
// 		}
// 	})

// 	t.Run("variable + OR", func(t *testing.T) {
// 		var (
// 			a    = &varCondition{varName: "a"}
// 			b    = &orCondition{}
// 			want = "a OR ??"
// 		)

// 		got, _ := appendToCondition(a, b)

// 		if got.String() != want {
// 			t.Fatalf("Want '%s', got '%s'", want, got.String())
// 		}
// 	})

// 	t.Run("Append a variable to a unary condition", func(t *testing.T) {
// 		var (
// 			a    = &notCondition{}
// 			b    = &varCondition{varName: "b"}
// 			want = "NOT b"
// 		)

// 		got, _ := appendToCondition(a, b)

// 		if got.String() != want {
// 			t.Fatalf("Want '%s', got '%s'", want, got.String())
// 		}
// 	})

// 	t.Run("Append a unary condition to another unary condition", func(t *testing.T) {
// 		var (
// 			a    = &notCondition{}
// 			b    = &notCondition{}
// 			want = "NOT NOT ??"
// 		)

// 		got, _ := appendToCondition(a, b)

// 		if got.String() != want {
// 			t.Fatalf("Want '%s', got '%s'", want, got.String())
// 		}
// 	})

// 	for _, b := range []binaryConditionExpr{
// 		&andCondition{},
// 		&orCondition{},
// 	} {
// 		t.Run("Can't append a binary condition to a unary condition", func(t *testing.T) {
// 			a := &notCondition{}

// 			_, err := appendToCondition(a, b)

// 			if err == nil {
// 				t.Fatalf("Expected an error")
// 			}

// 			if err.Reason != ParseErrBinaryAfterUnary {
// 				t.Fatalf("Want reason ParseErrBinaryAfterUnary, got %v", err.Reason)
// 			}
// 		})
// 	}

// 	for _, tCase := range []struct {
// 		a    binaryConditionExpr
// 		want string
// 	}{
// 		{a: &andCondition{}, want: "?? AND b"},
// 		{a: &orCondition{}, want: "?? OR b"},
// 	} {
// 		t.Run("Append a variable to a binary condition", func(t *testing.T) {
// 			b := &varCondition{varName: "b"}
// 			got, _ := appendToCondition(tCase.a, b)

// 			if got.String() != tCase.want {
// 				t.Fatalf("Want '%s', got '%s'", tCase.want, got.String())
// 			}
// 		})

// 	}

// 	for _, tCase := range []struct {
// 		a    binaryConditionExpr
// 		want string
// 	}{
// 		{a: &andCondition{}, want: "?? AND NOT ??"},
// 		{a: &orCondition{}, want: "?? OR NOT ??"},
// 	} {
// 		t.Run("Append a unary condition to a binary condition", func(t *testing.T) {
// 			b := &notCondition{}
// 			got, _ := appendToCondition(tCase.a, b)

// 			if got.String() != tCase.want {
// 				t.Fatalf("Want '%s', got '%s'", tCase.want, got.String())
// 			}
// 		})
// 	}

// 	for _, a := range []binaryConditionExpr{
// 		&andCondition{},
// 		&orCondition{},
// 	} {
// 		t.Run("Can't append a binary condition to a binary condition", func(t *testing.T) {
// 			b := &andCondition{}
// 			_, err := appendToCondition(a, b)

// 			if err == nil {
// 				t.Fatal("Expected error")
// 			}

// 			if err.Reason != ParseErrContigBinary {
// 				t.Fatalf("Want reason ParseErrContigBinary, got %v", err.Reason)
// 			}
// 		})
// 	}

// 	// for _, a := range []conditionExpr{
// 	// 	// &andCondition{rhs: &varCondition{varName: "x"}},
// 	// 	// &orCondition{rhs: &varCondition{varName: "x"}},
// 	// 	&notCondition{op: &varCondition{varName: "x"}},
// 	// } {
// 	// 	t.Run(
// 	// 		fmt.Sprintf("Can't append to %s with a defined rhs which doesn't have a rhs left", a),
// 	// 		func(t *testing.T) {
// 	// 			b := &varCondition{varName: "b"}
// 	// 			_, err := a.append(b)

// 	// 			if err == nil {
// 	// 				t.Fatal("Expected error")
// 	// 			}
// 	// 		})
// 	// }

// 	t.Run("Compound condition: a AND NOT b", func(t *testing.T) {
// 		var (
// 			a    = &varCondition{varName: "a"}
// 			and  = &andCondition{}
// 			not  = &notCondition{}
// 			b    = &varCondition{varName: "b"}
// 			want = "a AND NOT b"
// 		)

// 		res, _ := appendAll([]conditionExpr{a, and, not, b})
// 		got := res.String()

// 		if got != want {
// 			t.Fatalf("Want %s, got %s", want, got)
// 		}
// 	})

// 	t.Run("Compound operation: a OR NOT b", func(t *testing.T) {
// 		var (
// 			a    = &varCondition{varName: "a"}
// 			or   = &orCondition{}
// 			not  = &notCondition{}
// 			b    = &varCondition{varName: "b"}
// 			want = "a OR NOT b"
// 		)

// 		res, _ := appendAll([]conditionExpr{a, or, not, b})
// 		got := res.String()

// 		if got != want {
// 			t.Fatalf("Want %s, got %s", want, got)
// 		}
// 	})
// }
