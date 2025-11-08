// Package eval implements an interpreter for the Cow language.
// It evaluates an Abstract Syntax Tree (AST) and executes the program.
package eval

import (
	"fmt"
	"io"

	"github.com/shadowCow/cow-lang-go/lang/ast"
)

// Environment stores variable bindings.
type Environment struct {
	store map[string]interface{}
}

// NewEnvironment creates a new environment.
func NewEnvironment() *Environment {
	return &Environment{
		store: make(map[string]interface{}),
	}
}

// Get retrieves a variable value from the environment.
func (env *Environment) Get(name string) (interface{}, bool) {
	value, exists := env.store[name]
	return value, exists
}

// Set stores a variable value in the environment.
func (env *Environment) Set(name string, value interface{}) {
	env.store[name] = value
}

// Evaluator holds the state during evaluation.
type Evaluator struct {
	output io.Writer     // Where to write println output
	env    *Environment  // Variable storage
}

// NewEvaluator creates a new evaluator.
// The output writer is where println statements will write to.
func NewEvaluator(output io.Writer) *Evaluator {
	return &Evaluator{
		output: output,
		env:    NewEnvironment(),
	}
}

// Eval evaluates a program AST.
func (e *Evaluator) Eval(program *ast.Program) error {
	for _, stmt := range program.Statements {
		if err := e.evalStatement(stmt); err != nil {
			return err
		}
	}
	return nil
}

// evalStatement evaluates a single statement.
func (e *Evaluator) evalStatement(stmt ast.Statement) error {
	switch s := stmt.(type) {
	case *ast.LetStatement:
		return e.evalLetStatement(s)

	case *ast.ExpressionStatement:
		_, err := e.evalExpression(s.Expression)
		return err

	default:
		return fmt.Errorf("unknown statement type: %T", stmt)
	}
}

// evalLetStatement evaluates a let statement (variable declaration).
func (e *Evaluator) evalLetStatement(stmt *ast.LetStatement) error {
	// Evaluate the value expression
	value, err := e.evalExpression(stmt.Value)
	if err != nil {
		return fmt.Errorf("error evaluating let statement for '%s': %v", stmt.Name, err)
	}

	// Store the variable in the environment
	e.env.Set(stmt.Name, value)
	return nil
}

// evalExpression evaluates an expression and returns its value.
// For now, values are represented as interface{} and can be int64 or float64.
func (e *Evaluator) evalExpression(expr ast.Expression) (interface{}, error) {
	switch ex := expr.(type) {
	case *ast.IntLiteral:
		return ex.Value, nil

	case *ast.FloatLiteral:
		return ex.Value, nil

	case *ast.Identifier:
		return e.evalIdentifier(ex)

	case *ast.FunctionCall:
		return e.evalFunctionCall(ex)

	case *ast.BinaryExpression:
		return e.evalBinaryExpression(ex)

	default:
		return nil, fmt.Errorf("unknown expression type: %T", expr)
	}
}

// evalIdentifier evaluates an identifier (variable reference).
func (e *Evaluator) evalIdentifier(id *ast.Identifier) (interface{}, error) {
	value, exists := e.env.Get(id.Name)
	if !exists {
		return nil, fmt.Errorf("undefined variable: %s", id.Name)
	}
	return value, nil
}

// evalFunctionCall evaluates a function call.
func (e *Evaluator) evalFunctionCall(call *ast.FunctionCall) (interface{}, error) {
	// For now, we only support the built-in println function
	if call.Name != "println" {
		return nil, fmt.Errorf("unknown function: %s", call.Name)
	}

	// Evaluate all arguments
	for i, arg := range call.Arguments {
		value, err := e.evalExpression(arg)
		if err != nil {
			return nil, fmt.Errorf("error evaluating argument %d to println: %v", i, err)
		}

		// Print the value
		if err := e.println(value); err != nil {
			return nil, err
		}
	}

	// println returns void/nil
	return nil, nil
}

// println prints a value to the output writer.
func (e *Evaluator) println(value interface{}) error {
	var str string

	switch v := value.(type) {
	case int64:
		str = fmt.Sprintf("%d\n", v)
	case float64:
		str = fmt.Sprintf("%g\n", v)
	default:
		return fmt.Errorf("cannot print value of type %T", value)
	}

	_, err := e.output.Write([]byte(str))
	return err
}

// evalBinaryExpression evaluates a binary expression (e.g., 1 + 2, x * y).
func (e *Evaluator) evalBinaryExpression(expr *ast.BinaryExpression) (interface{}, error) {
	// Evaluate left operand
	leftVal, err := e.evalExpression(expr.Left)
	if err != nil {
		return nil, fmt.Errorf("error evaluating left operand of %s: %v", expr.Operator, err)
	}

	// Evaluate right operand
	rightVal, err := e.evalExpression(expr.Right)
	if err != nil {
		return nil, fmt.Errorf("error evaluating right operand of %s: %v", expr.Operator, err)
	}

	// Type coercion: if one operand is float, convert both to float
	leftInt, leftIsInt := leftVal.(int64)
	leftFloat, leftIsFloat := leftVal.(float64)
	rightInt, rightIsInt := rightVal.(int64)
	rightFloat, rightIsFloat := rightVal.(float64)

	if !leftIsInt && !leftIsFloat {
		return nil, fmt.Errorf("left operand of %s has invalid type: %T", expr.Operator, leftVal)
	}
	if !rightIsInt && !rightIsFloat {
		return nil, fmt.Errorf("right operand of %s has invalid type: %T", expr.Operator, rightVal)
	}

	// If either operand is float, do float arithmetic
	if leftIsFloat || rightIsFloat {
		var left, right float64
		if leftIsFloat {
			left = leftFloat
		} else {
			left = float64(leftInt)
		}
		if rightIsFloat {
			right = rightFloat
		} else {
			right = float64(rightInt)
		}
		return e.evalFloatBinaryOp(left, right, expr.Operator)
	}

	// Both are integers - do integer arithmetic
	return e.evalIntBinaryOp(leftInt, rightInt, expr.Operator)
}

// evalIntBinaryOp performs integer binary operations.
func (e *Evaluator) evalIntBinaryOp(left, right int64, operator string) (interface{}, error) {
	switch operator {
	case "+":
		return left + right, nil
	case "-":
		return left - right, nil
	case "*":
		return left * right, nil
	case "/":
		if right == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return left / right, nil
	case "%":
		if right == 0 {
			return nil, fmt.Errorf("modulo by zero")
		}
		return left % right, nil
	default:
		return nil, fmt.Errorf("unknown binary operator: %s", operator)
	}
}

// evalFloatBinaryOp performs floating-point binary operations.
func (e *Evaluator) evalFloatBinaryOp(left, right float64, operator string) (interface{}, error) {
	switch operator {
	case "+":
		return left + right, nil
	case "-":
		return left - right, nil
	case "*":
		return left * right, nil
	case "/":
		if right == 0.0 {
			return nil, fmt.Errorf("division by zero")
		}
		return left / right, nil
	case "%":
		return nil, fmt.Errorf("modulo operator not supported for floating-point numbers")
	default:
		return nil, fmt.Errorf("unknown binary operator: %s", operator)
	}
}
