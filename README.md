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

Patterns are either sequences of hexadecimal numbers (byte sequences) or strings.

**Byte sequences**.
Byte sequences are represented by hexadecimal numbers.
They appear enclosed between curly brackets, as a string in the yaml file.
For example:

```
a: '{ 74 fc ff ff c6 05 19 45 }'
```

Byte sequences may contain two question marks (`??`) that match any byte at that position.
Here's an example:

```
b: '{ 51 67 ?? ?? 44 }'
```

The previous pattern matches any sequence of bytes that starts with `0x51 0x67`, then has two arbitrary bytes, and last, a `0x44` byte.
Examples of byte sequences that would be matched:

- `0x51 0x67 0xab 0xcd 0x44`
- `0x51 0x67 0x11 0x22 0x44`
- `0x51 0x67 0x33 0xff 0x44`

**Strings**.
Only ASCII strings are supported at the moment.
