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
