# The condition parser

Every signature has a `condition`: a boolean expression over the names of its
patterns that decides whether a file matches. This document explains how the
`internal/bexpr` package turns that text into something that can be evaluated,
one token at a time, and how it reports what went wrong when it can't.

```yaml
patterns:
  header: '{ 4d 5a ?? 00 }'
  marker: binmat
  kill: '{ de ad be ef }'
condition: header AND (marker OR NOT kill)
```

When a file is scanned, each pattern is searched for independently and the
result is a map from pattern name to "was it found". The condition is then
evaluated against that map.

## Table of contents

1. [The language](#the-language)
2. [Where the parser sits](#where-the-parser-sits)
3. [The pipeline](#the-pipeline)
4. [Tokenizer](#tokenizer)
5. [The expression tree](#the-expression-tree)
6. [Building the tree: the append algorithm](#building-the-tree-the-append-algorithm)
7. [Parentheses and groups](#parentheses-and-groups)
8. [Completeness](#completeness)
9. [Evaluation](#evaluation)
10. [Errors](#errors)
11. [Quirks and limitations](#quirks-and-limitations)
12. [Tests](#tests)

## The language

A condition is made of variables, three operators and parentheses.

**Variables** are pattern names: one to sixteen characters from lowercase
letters, digits and underscores (`^[a-z0-9_]{1,16}$`). `a`, `b1`, `c_2` and
`foo_bar` are valid names.

**Operators** are the uppercase keywords `NOT`, `AND` and `OR`. They bind with
different strength when parentheses don't say otherwise:

| Operator | Precedence | Meaning |
|---|---|---|
| `NOT` | highest | true when its operand is false |
| `AND` | middle | true when both operands are true |
| `OR` | lowest | true when either operand is true |

Operators of the same precedence associate to the left. **Parentheses** group a
sub-expression so it's treated as a single operand, overriding precedence.

| Written | Read as |
|---|---|
| `a OR b AND c` | `a OR (b AND c)` |
| `a AND b OR c` | `(a AND b) OR c` |
| `NOT a AND b` | `(NOT a) AND b` |
| `a AND b AND c` | `(a AND b) AND c` |
| `a OR NOT b AND c` | `a OR ((NOT b) AND c)` |
| `NOT (a OR b) AND c` | `(NOT (a OR b)) AND c` |

For reference, this is the grammar of the accepted language. The parser doesn't
implement it as a recursive descent over these rules (see below), but the set
of expressions it accepts is exactly this one:

```
condition = or_expr
or_expr   = and_expr { "OR" and_expr }
and_expr  = not_expr { "AND" not_expr }
not_expr  = "NOT" not_expr | primary
primary   = variable | "(" condition ")"
variable  = /[a-z0-9_]{1,16}/
```

Whitespace between tokens is ignored, so `a AND(b OR c)` and `  a   AND ( b OR c )  `
are the same condition. Variables and keywords do have to be separated from
each other by whitespace or a parenthesis, though: `a ANDb` is an error, not a
spelling of `a AND b`.

## Where the parser sits

The public surface of the package is small:

```go
// Turns the text into an evaluable function, or explains why it can't.
func ParseCondition(condition string) (Condition, *ErrConditionParse)

// Evaluates the condition against the pattern match results.
type Condition func(map[string]bool) (bool, *ErrMissingVarValue)

// Checks a single variable name.
func IsValidVarName(name string) bool
```

A `Signature` calls `ParseCondition` while compiling, right after parsing its
patterns. It then evaluates the resulting `Condition` once with every pattern
name set to `true`: if that returns an `ErrMissingVarValue`, the condition
refers to a pattern the signature doesn't define, and the signature is rejected
with `ErrSigMissingPattern`. This is why a condition can never fail at scan
time for referring to an unknown pattern.

At scan time, `Signature.CheckMatch` searches every pattern, builds the
`map[string]bool` of results, and calls the `Condition` with it.

## The pipeline

`ParseCondition` runs three steps:

```
"a OR b AND c"
      │
      ▼  tokenize()                        tokenizer.go
["a" "OR" "b" "AND" "c"]
      │
      ▼  parse(), one appendToCondition()  parser.go, append.go
      │  per token
    OR
   /  \                                    (the expression tree)
  a   AND
      / \
     b   c
      │
      ▼  isCondComplete()                  conditions.go
      │  balance and completeness checks
      ▼
Condition closure that calls tree.apply(vars)
```

The distinguishing feature is the second step. Instead of a recursive descent
that looks ahead to decide what to build, the parser walks the tokens left to
right and, for each one, creates a node and *appends* it to the tree built so
far. The tree is always a faithful picture of the tokens consumed up to that
point, which is what makes error messages like `'a AND ??'` possible: the `??`
marks the operand that never arrived.

## Tokenizer

`tokenize` in `tokenizer.go` finds the tokens with a single regular expression:

```
[a-z0-9_]+|AND|OR|NOT|\(|\)
```

`FindAllStringIndex` returns every match, in order, with its position.
Alternation is tried left to right at each position, so a run of lowercase
characters is always a variable, and the uppercase keywords are only recognised
as such because lowercase letters can't start them.

The regular expression alone would skip anything it doesn't match, so the
tokenizer also inspects the text *between* consecutive matches, and before the
first and after the last. That text may contain only whitespace, as defined by
`unicode.IsSpace`, which is how spaces, tabs and newlines are ignored and why
parentheses don't need surrounding spaces. The first character in a gap that
isn't whitespace stops tokenizing with `ParseErrUnexpectedChar`, reporting the
character and its 1-based column, counted in characters rather than bytes:

```
a & b        → unexpected character ('&' at column 3)
Foo AND bar  → unexpected character ('F' at column 1)
a AND b-c    → unexpected character ('-' at column 8)
ANDY         → unexpected character ('Y' at column 4)
```

Without this check `Foo AND bar` would tokenize to `["oo" "AND" "bar"]` and
parse as a different, valid condition.

A second check covers the opposite problem: two matches with *nothing* between
them. The regular expression happily splits `ANDb` into `AND` and `b`, so the
tokenizer rejects consecutive word tokens, variables or keywords, that touch.
Parentheses are exempt, since `(a OR b)` is the normal way to write a group.
The whole glued run is reported as one word, so the message shows what was
actually typed:

```
a ANDb       → missing whitespace between tokens ('ANDb' at column 3)
aAND b       → missing whitespace between tokens ('aAND' at column 1)
NOTNOT a     → missing whitespace between tokens ('NOTNOT' at column 1)
```

Variable length is not enforced here. The regex accepts runs of any length; the
parser checks each variable token against `IsValidVarName` and reports
`ParseErrInvalidVarName` for names longer than sixteen characters.

## The expression tree

Every node implements `conditionExpr` (`conditions.go`):

```go
type conditionExpr interface {
    fmt.Stringer
    apply(map[string]bool) (bool, *ErrMissingVarValue)
}
```

`apply` evaluates the subtree. `String` renders it, printing `??` wherever an
operand is missing. There are five node types in three families, and the
families, not the concrete types, are what the parser reasons about:

| Family | Interface | Types | Operands |
|---|---|---|---|
| variable | `varConditionExpr` | `varCondition` | none |
| unary | `unaryConditionExpr` | `notCondition`, `groupCondition` | one (`op`) |
| binary | `binaryConditionExpr` | `andCondition`, `orCondition` | two (`lhs`, `rhs`) |

A parenthesised group is a unary node whose single operand is the expression
inside the parentheses. Treating it as unary is what makes it behave as an
atomic operand once built.

The interfaces expose operand access:

```go
type unaryConditionExpr interface {
    conditionExpr
    hasOp() bool
    setOp(expr conditionExpr) *errAppendToCond
    getOp() conditionExpr
}

type binaryConditionExpr interface {
    conditionExpr
    precedence() int
    hasRhs() bool
    setRhs(expr conditionExpr) *errAppendToCond
    getRhs() conditionExpr
    hasLhs() bool
    setLhs(expr conditionExpr)
    getLhs() conditionExpr
}
```

Note the asymmetry between `setLhs` and `setRhs`. The left operand of a binary
operator is always known when the operator is created (it's whatever was
parsed before it), so `setLhs` is a plain assignment. The right operand arrives
later, token by token, so `setRhs` has *append* semantics: if the slot is
empty, it takes the expression, provided it's a variable or a unary node
(`isOperand`); if the slot is taken, it appends the expression to whatever is
there and stores the result back. `setOp` on unary nodes works the same way.
This recursion is how `a AND NOT NOT b` ends up as `a AND (NOT (NOT b))`: each
new token is pushed down into the deepest open slot.

`precedence()` returns `precedenceAnd` (2) for `AND` and `precedenceOr` (1) for
`OR`. Higher binds tighter.

## Building the tree: the append algorithm

`appendToCondition(base, toAppend)` in `append.go` is the heart of the parser.
It takes the tree built so far and one new node, and returns the new root of
the tree, which may or may not be the same node as before. The `parse` loop
calls it once per token and keeps the returned root.

The behaviour depends on the family of the base and of the new node:

| Base | New node | Result |
|---|---|---|
| nothing yet | anything | the new node is the root |
| variable | binary | the variable becomes the operator's `lhs`; the operator is the new root |
| unary | variable or unary | pushed into the unary's operand (`setOp`); root unchanged |
| unary, complete | binary | the unary node becomes the operator's `lhs`; the operator is the new root |
| binary | variable or unary | pushed into the `rhs` (`setRhs`); root unchanged |
| binary, complete | binary, tighter | pushed into the `rhs` (`setRhs`), where it takes the old `rhs` as its `lhs`; root unchanged |
| binary, complete | binary, same or looser | the whole base becomes the operator's `lhs`; the operator is the new root |
| anything else | | `errAppendToCond`, reported as `ParseErrInvalidAppend` |

"Complete" means `isCondComplete` is true for the base: no `??` anywhere in it.
A binary operator can't follow an expression that still has a hole, so
`a AND AND b` and `NOT AND b` are rejected.

### The right spine

An invariant makes this work. At any point during parsing, the only places in
the tree that can still be empty lie on its *right spine*: the path from the
root that always takes the right operand of a binary node or the operand of a
unary node. Everything to the left of that path is finished and will never
change.

- A **variable or unary node** fills the deepest empty slot on the spine. The
  `setRhs` and `setOp` recursion walks down the spine until it finds it.
- A **binary operator** climbs the spine looking for its place. Compared with
  the root, if it binds tighter it descends into the right operand and repeats
  the comparison there; otherwise it stops and takes the whole subtree at that
  level as its left operand. `NOT` and groups always stop the descent, since
  they bind tighter than any binary operator.

This is precedence climbing performed incrementally, and it is why the parser
needs no lookahead: the decision for a token depends only on the tree built so
far.

### Walkthroughs

Each step shows the token consumed and the tree afterwards.

**`a AND NOT b`**

```
a          AND        NOT          b

a          AND        AND          AND
          /   \      /   \        /   \
         a    ??    a    NOT     a    NOT
                          |            |
                         ??            b
```

`AND` takes `a` as its left operand and becomes the root. `NOT` is a unary
node, so it fills the empty right slot. `b` is pushed down through `setRhs`
into the `NOT`, whose slot is the deepest empty one.

**`a OR b AND c`** (the new operator binds tighter)

```
a OR b        AND            c

  OR           OR             OR
 /  \         /  \           /  \
a    b       a   AND        a   AND
                 /  \           /  \
                b    ??        b    c
```

When `AND` arrives, the root is a complete `OR`. `AND` binds tighter, so it is
pushed into the right operand: `appendToCondition(b, AND)` turns `b` into the
`AND`'s left operand and `setRhs` stores the resulting `AND` back as the `OR`'s
right operand. The root is still the `OR`, and the result reads `a OR (b AND c)`.

**`a AND b OR c`** (the new operator binds looser)

```
a AND b        OR              c

  AND           OR              OR
 /   \         /  \            /  \
a     b      AND   ??        AND   c
             / \             / \
            a   b           a   b
```

`OR` doesn't bind tighter than the `AND` at the root, so the whole `AND` becomes
its left operand and the `OR` is the new root: `(a AND b) OR c`.

**`a AND b AND c`** (same precedence, left associative)

Same shape as the previous one: the second `AND` takes the first as its left
operand, giving `(a AND b) AND c`.

**`NOT a AND b`**

```
NOT        a          AND          b

NOT        NOT        AND          AND
 |          |        /   \        /   \
??          a      NOT    ??    NOT    b
                    |            |
                    a            a
```

Once `NOT a` is complete, the `AND` takes it whole as its left operand. A
`NOT` never absorbs a binary operator into its operand, which is what gives it
the highest precedence.

**`(a OR b) AND c`**

```
( … )        AND          c

(OR)         AND          AND
 / \        /   \        /   \
a   b     (OR)   ??    (OR)   c
           / \          / \
          a   b        a   b
```

The parenthesised part is parsed on its own (next section) and wrapped in a
group node, which is unary and complete. The `AND` then treats it like any
other complete operand.

### Why the setters store what the append returns

Before precedence support, `setRhs` and `setOp` called `appendToCondition` on
their existing operand and discarded the returned root, because the only
appends that could reach a taken slot never changed the subtree's root. With
precedence, the `a OR b` + `AND` step above *does* change it: the right operand
goes from `b` to `AND(b, ??)`. So the setters now assign the returned root back
into the slot. Forgetting this would silently drop the new operator.

## Parentheses and groups

`parse(p, depth)` in `parser.go` is the token loop. `depth` is the number of
parentheses currently open; the top-level call uses zero.

- On `(`, `parse` calls itself with `depth + 1`. The recursive call consumes
  tokens until it meets the `)` that closes this group, returns the expression
  it built, and the caller wraps it in a `groupCondition` and appends that as a
  single node. Whatever was inside the parentheses is now opaque to the
  operators outside them.
- On `)`, if `depth` is zero there is no group to close and parsing fails with
  `ParseErrUnbalancedParens` and the detail `unexpected ')'`. Otherwise the
  loop records that the group was closed and stops, handing control back to
  the caller.
- When the tokens run out, a call with `depth > 0` that never saw its `)`
  fails with `ParseErrUnbalancedParens` and the detail `missing ')'`.

Nested groups are just nested recursive calls, each with its own depth, so
`((a OR b) AND c) OR d` works at any depth.

The balance check runs before the completeness check. `(a AND` is reported as
a missing parenthesis, not as an incomplete `AND`, since the structural problem
is the more fundamental one.

## Completeness

After the loop, `isCondComplete` walks the whole tree and checks that every
binary node has both operands and every unary node has its one, recursively.
A missing operand anywhere yields `ParseErrIncompleteExpr`, with the rendered
tree as detail so the `??` shows where the hole is:

```
a AND         → incomplete binary operation ('a AND ??')
AND b         → incomplete binary operation ('?? AND b')
a AND NOT NOT → incomplete binary operation ('a AND NOT NOT ??')
a AND ()      → incomplete binary operation ('a AND (??)')
```

The check has to be recursive. A shallow version that only inspected the root
would accept `a AND NOT NOT`, whose root `AND` has both operands, and the
resulting condition would dereference a nil operand when evaluated.

An empty condition produces no tree at all. `ParseCondition("")` succeeds and
returns a `Condition` that is always false. In practice a signature rejects an
empty condition before it reaches the parser, so this only matters when using
the package directly.

## Evaluation

The `Condition` returned by `ParseCondition` is a closure over the tree's root.
Calling it calls `root.apply(vars)`, and each node type evaluates itself:

- a variable looks its name up in the map, and returns `ErrMissingVarValue`
  if the name isn't there;
- `NOT` and groups evaluate their operand, negating it in the case of `NOT`;
- `AND` and `OR` evaluate their left operand, then their right one, and
  combine the results.

There is **no short-circuiting**: both operands of a binary node are always
evaluated. This is deliberate. It means a missing variable is reported no
matter where it sits in the tree, which is what lets `Signature` validate at
load time, with a single evaluation, that every variable the condition uses is
a defined pattern.

Evaluation never fails for a valid tree other than through a missing
variable, and after the load-time check that can't happen either, so
`Signature.CheckMatch` ignores the error.

## Errors

`ParseCondition` returns an `*ErrConditionParse` carrying the original text, a
`ParseErrorReason`, and a detail string. Its message reads
`can't parse the expression '<condition>'. Reason: <reason> (<detail>)`.

| Reason | Constant | Raised when | Detail | Example |
|---|---|---|---|---|
| unexpected character | `ParseErrUnexpectedChar` | the text between two tokens contains something other than whitespace | the character and its 1-based column | `a & b`, `Foo AND bar` |
| missing whitespace between tokens | `ParseErrMissingWhitespace` | two variables or keywords follow each other with nothing in between | the glued run and its 1-based column | `a ANDb`, `NOTNOT a` |
| invalid variable name | `ParseErrInvalidVarName` | a variable token fails `IsValidVarName` (in practice, it's longer than 16 characters) | the offending name and the rule | `a_very_long_variable_name AND b` |
| invalid append attempt | `ParseErrInvalidAppend` | a token can't follow what came before it | the expression it couldn't be appended to | `a b`, `a NOT b`, `a AND AND b`, `(a OR b) c` |
| incomplete binary operation | `ParseErrIncompleteExpr` | an operand is missing once the tokens run out | the rendered tree, with `??` for the hole | `a AND`, `OR b`, `a AND ()` |
| unbalanced parentheses | `ParseErrUnbalancedParens` | a `)` has no matching `(`, or a `(` is never closed | `unexpected ')'` or `missing ')'` | `a AND b)`, `a AND (b` |

Internally, `appendToCondition` returns an `errAppendToCond` naming the two
nodes involved; `parse` converts it into an `ErrConditionParse` with the
`ParseErrInvalidAppend` reason via `toParseErr`.

Evaluation has a single error, `ErrMissingVarValue`, described above.

## Quirks and limitations

- **Keywords are case-sensitive.** `a and b` treats `and` as a variable name
  and fails with an invalid append. Only uppercase `AND`, `OR` and `NOT` are
  operators.
- **Uppercase typos read as unexpected characters.** Only the three keywords
  contain uppercase letters, so `ANDB` reports an unexpected `B` at column 4
  while `ANDb` reports missing whitespace. Both are errors; the wording differs
  because the regular expression never produces a token for stray uppercase
  letters.
- **No short-circuit evaluation.** Cheap by design here, since evaluation is a
  handful of map lookups per file, but worth knowing if the tree ever grows
  expensive leaves.
- **Error positions are implicit.** Errors quote the whole condition and the
  partial tree, not a column. For the short conditions signatures use this is
  enough; longer ones might want a position.

## Tests

The package is tested at three levels, all in `internal/bexpr`:

- `tokenizer_test.go` checks the token stream for a few inputs, and that
  unexpected characters and glued tokens are reported with the right column.
- `cond_and_test.go`, `cond_or_test.go`, `cond_not_test.go` and
  `cond_group_test.go` exercise each node type in isolation: operand setters,
  `String()` rendering with `??`, and the rules on what can be set as an
  operand.
- `parser_test.go` drives `ParseCondition` end to end. `TestParseCondition`
  covers the basic shapes and error reasons. `TestParseChainedConditions`
  covers chained operators and precedence with explicit truth tables.
  `TestParseIncompleteChains` covers malformed chains,
  `TestParseUnbalancedParentheses` covers parenthesis balance in both
  directions plus deep nesting, and `TestParseUnexpectedCharacters` and
  `TestParseMissingWhitespace` check that tokenizer errors surface through
  `ParseCondition`.

The truth-table style is the recommended way to add coverage for a new shape:
list the condition, then every combination of variable values that matters and
the expected result. When two readings of an expression are possible, include
the inputs that distinguish them. For `a OR b AND c`, the case
`a=true, b=false, c=false` is what separates `a OR (b AND c)` (true) from
`(a OR b) AND c` (false).
