# Cow Language Examples

This directory contains example Cow language programs.

## hello_numbers.cow

A simple program that demonstrates printing various number literals:
- Decimal integers
- Hexadecimal integers (0xFF)
- Binary integers (0b1010)
- Floating-point numbers
- Scientific notation
- Numbers with underscores for readability

## Running Examples

To run these examples, you'll need to create a main program that:
1. Reads the source file
2. Compiles the lexical grammar using `langdef.GetLexical()`
3. Tokenizes using `lexer.NewLexer(dfa, source)`
4. Parses using `parser.NewParser(tokens)`
5. Evaluates using `eval.NewEvaluator(os.Stdout)`

See the integration tests in `langdef/interpreter_integration_test.go` for complete examples.
