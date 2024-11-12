# Binary Matcher

A CLI to match binary files using signatures.
A signature is defined in a _.yaml_ file, and contains the following fields:

- `name`: The name given to the signature.
- `description`: An optional description of what the signature matches.
- `patterns`: Named sequences of bytes or strings to be searched for in the binary.
  Pattern names can include between 1 and 16 lowercase letters, numbers and underscores.
- `condition`: A boolean expression of the defined patterns, joined using the following operations:
  - `AND`: true when both operands are true.
  - `OR`: true when either operand is true.
  - `NOT`: negates an expression.

Here's an example of a signature:

```yaml
name: example signature
description: a signature for demonstration purposes
patterns:
  a: '{ 74 fc ff ff c6 05 19 45 }'
  b: '{ 51 67 ?? ?? 44 }'
  c: this is a string
condition: a AND (b OR c)
```

Patterns are either sequences of hexadecimal numbers (bytes) or strings.
In the case of hexadecimal numbers, they appear enclosed between curly brackets.
Hexadecimal sequences may contain two question marks (`??`) that match any byte at that position.
