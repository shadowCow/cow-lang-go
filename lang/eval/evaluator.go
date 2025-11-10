// Package eval implements an interpreter for the Cow language.
// It evaluates an Abstract Syntax Tree (AST) and executes the program.
package eval

import (
	"fmt"
	"io"
	"strings"

	"github.com/shadowCow/cow-lang-go/lang/ast"
)

// Environment stores variable bindings with support for scope chaining.
type Environment struct {
	store  map[string]interface{}
	parent *Environment // Parent environment for scope chain (nil for global scope)
}

// NewEnvironment creates a new environment with an optional parent.
// Pass nil for parent to create a global scope environment.
func NewEnvironment(parent *Environment) *Environment {
	return &Environment{
		store:  make(map[string]interface{}),
		parent: parent,
	}
}

// Get retrieves a variable value from the environment.
// Searches up the scope chain if not found in current environment.
func (env *Environment) Get(name string) (interface{}, bool) {
	// Check local scope first
	if value, exists := env.store[name]; exists {
		return value, true
	}
	// Check parent scope if it exists
	if env.parent != nil {
		return env.parent.Get(name)
	}
	return nil, false
}

// Set stores a variable value in the environment.
func (env *Environment) Set(name string, value interface{}) {
	env.store[name] = value
}

// Function represents a user-defined function at runtime.
// Functions are first-class values that can be stored in variables.
type Function struct {
	Parameters []string  // Parameter names
	Body       *ast.Block // Function body
	// Note: No Env field - we don't capture closures, only access globals
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
		env:    NewEnvironment(nil), // nil parent = global scope
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

	case *ast.FunctionDef:
		return e.evalFunctionDef(s)

	case *ast.ReturnStatement:
		// Return statements should only be evaluated inside blocks
		// This case shouldn't be hit at the top level
		return fmt.Errorf("return statement outside function")

	case *ast.Block:
		_, err := e.evalBlock(s)
		return err

	case *ast.IndexAssignment:
		return e.evalIndexAssignment(s)

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

	case *ast.BoolLiteral:
		return ex.Value, nil

	case *ast.StringLiteral:
		return ex.Value, nil

	case *ast.Identifier:
		return e.evalIdentifier(ex)

	case *ast.FunctionCall:
		return e.evalFunctionCall(ex)

	case *ast.UnaryExpression:
		return e.evalUnaryExpression(ex)

	case *ast.BinaryExpression:
		return e.evalBinaryExpression(ex)

	case *ast.FunctionLiteral:
		return e.evalFunctionLiteral(ex)

	case *ast.ArrayLiteral:
		return e.evalArrayLiteral(ex)

	case *ast.IndexAccess:
		return e.evalIndexAccess(ex)

	case *ast.MemberAccess:
		return e.evalMemberAccess(ex)

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
	// Check for built-in println function
	if call.Name == "println" {
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

	// Check for array method calls (like len, push, pop)
	// These are represented as function calls where the first argument is the array object
	if len(call.Arguments) > 0 {
		// Evaluate first argument to see if it's an ArrayMethod
		firstArg, err := e.evalExpression(call.Arguments[0])
		if err == nil {
			if arrayMethod, ok := firstArg.(*ArrayMethod); ok {
				// This is a method call on an array
				return e.callArrayMethod(arrayMethod, call.Arguments[1:])
			}
		}
	}

	// Look up user-defined function
	fnValue, exists := e.env.Get(call.Name)
	if !exists {
		return nil, fmt.Errorf("undefined function: %s", call.Name)
	}

	// Check if it's a Function
	fn, ok := fnValue.(*Function)
	if !ok {
		return nil, fmt.Errorf("%s is not a function (it's a %T)", call.Name, fnValue)
	}

	// Call the user-defined function
	return e.callUserFunction(fn, call.Arguments)
}

// callArrayMethod calls an array method (len, push, pop)
func (e *Evaluator) callArrayMethod(method *ArrayMethod, args []ast.Expression) (interface{}, error) {
	switch method.Method {
	case "len":
		// len() takes no arguments and returns the length
		if len(args) != 0 {
			return nil, fmt.Errorf("len() takes no arguments, got %d", len(args))
		}
		return int64(len(method.Array)), nil

	case "push":
		// push(item) appends an item and returns nil
		if len(args) != 1 {
			return nil, fmt.Errorf("push() takes exactly 1 argument, got %d", len(args))
		}

		// Evaluate the item to push
		item, err := e.evalExpression(args[0])
		if err != nil {
			return nil, fmt.Errorf("error evaluating push argument: %v", err)
		}

		// Append to the array
		newArray := append(method.Array, item)

		// Update the array in the environment
		// We need to get the identifier name from the object expression
		if ident, ok := method.Object.(*ast.Identifier); ok {
			e.env.Set(ident.Name, newArray)
		} else {
			return nil, fmt.Errorf("push() only supported on simple identifiers, not complex expressions")
		}

		return nil, nil

	case "pop":
		// pop() removes and returns the last element
		if len(args) != 0 {
			return nil, fmt.Errorf("pop() takes no arguments, got %d", len(args))
		}

		if len(method.Array) == 0 {
			return nil, fmt.Errorf("cannot pop from empty array")
		}

		// Get the last element
		lastIndex := len(method.Array) - 1
		lastElement := method.Array[lastIndex]

		// Remove the last element
		newArray := method.Array[:lastIndex]

		// Update the array in the environment
		if ident, ok := method.Object.(*ast.Identifier); ok {
			e.env.Set(ident.Name, newArray)
		} else {
			return nil, fmt.Errorf("pop() only supported on simple identifiers, not complex expressions")
		}

		return lastElement, nil

	default:
		return nil, fmt.Errorf("unknown array method: %s", method.Method)
	}
}

// println prints a value to the output writer.
func (e *Evaluator) println(value interface{}) error {
	var str string

	switch v := value.(type) {
	case int64:
		str = fmt.Sprintf("%d\n", v)
	case float64:
		str = fmt.Sprintf("%g\n", v)
	case bool:
		str = fmt.Sprintf("%t\n", v)
	case string:
		str = fmt.Sprintf("%s\n", v)
	case []interface{}:
		str = e.formatArray(v) + "\n"
	default:
		return fmt.Errorf("cannot print value of type %T", value)
	}

	_, err := e.output.Write([]byte(str))
	return err
}

// formatArray formats an array for printing.
func (e *Evaluator) formatArray(arr []interface{}) string {
	if len(arr) == 0 {
		return "[]"
	}

	parts := make([]string, len(arr))
	for i, elem := range arr {
		switch v := elem.(type) {
		case int64:
			parts[i] = fmt.Sprintf("%d", v)
		case float64:
			parts[i] = fmt.Sprintf("%g", v)
		case bool:
			parts[i] = fmt.Sprintf("%t", v)
		case string:
			parts[i] = fmt.Sprintf("%q", v)
		case []interface{}:
			parts[i] = e.formatArray(v)
		default:
			parts[i] = fmt.Sprintf("%v", v)
		}
	}

	return "[" + strings.Join(parts, ", ") + "]"
}

// evalUnaryExpression evaluates a unary expression (e.g., !true, -5).
func (e *Evaluator) evalUnaryExpression(expr *ast.UnaryExpression) (interface{}, error) {
	// Evaluate the operand
	operand, err := e.evalExpression(expr.Operand)
	if err != nil {
		return nil, fmt.Errorf("error evaluating operand of %s: %v", expr.Operator, err)
	}

	switch expr.Operator {
	case "NOT", "!":
		// Logical NOT
		boolVal, ok := operand.(bool)
		if !ok {
			return nil, fmt.Errorf("logical NOT operator requires boolean operand, got %T", operand)
		}
		return !boolVal, nil

	case "MINUS", "-":
		// Unary minus
		switch v := operand.(type) {
		case int64:
			return -v, nil
		case float64:
			return -v, nil
		default:
			return nil, fmt.Errorf("unary minus operator requires numeric operand, got %T", operand)
		}

	default:
		return nil, fmt.Errorf("unknown unary operator: %s", expr.Operator)
	}
}

// evalBinaryExpression evaluates a binary expression.
// Handles arithmetic, comparison, equality, and logical operators.
func (e *Evaluator) evalBinaryExpression(expr *ast.BinaryExpression) (interface{}, error) {
	// For logical operators (&&, ||), use short-circuit evaluation
	if expr.Operator == "AND" || expr.Operator == "&&" {
		return e.evalLogicalAnd(expr)
	}
	if expr.Operator == "OR" || expr.Operator == "||" {
		return e.evalLogicalOr(expr)
	}

	// For other operators, evaluate both operands
	leftVal, err := e.evalExpression(expr.Left)
	if err != nil {
		return nil, fmt.Errorf("error evaluating left operand of %s: %v", expr.Operator, err)
	}

	rightVal, err := e.evalExpression(expr.Right)
	if err != nil {
		return nil, fmt.Errorf("error evaluating right operand of %s: %v", expr.Operator, err)
	}

	// Handle equality operators (work on any type)
	if expr.Operator == "EQUAL_EQUAL" || expr.Operator == "==" {
		return e.evalEquality(leftVal, rightVal), nil
	}
	if expr.Operator == "NOT_EQUAL" || expr.Operator == "!=" {
		return !e.evalEquality(leftVal, rightVal), nil
	}

	// Handle string operations
	leftStr, leftIsStr := leftVal.(string)
	rightStr, rightIsStr := rightVal.(string)
	if leftIsStr || rightIsStr {
		// If one operand is a string, both must be strings
		if !leftIsStr || !rightIsStr {
			return nil, fmt.Errorf("type mismatch: cannot use %s operator with %T and %T",
				expr.Operator, leftVal, rightVal)
		}
		return e.evalStringBinaryOp(leftStr, rightStr, expr.Operator)
	}

	// Handle comparison and arithmetic operators (require numeric types)
	leftInt, leftIsInt := leftVal.(int64)
	leftFloat, leftIsFloat := leftVal.(float64)
	rightInt, rightIsInt := rightVal.(int64)
	rightFloat, rightIsFloat := rightVal.(float64)

	if !leftIsInt && !leftIsFloat {
		return nil, fmt.Errorf("left operand of %s has invalid type: %T (expected number)", expr.Operator, leftVal)
	}
	if !rightIsInt && !rightIsFloat {
		return nil, fmt.Errorf("right operand of %s has invalid type: %T (expected number)", expr.Operator, rightVal)
	}

	// If either operand is float, do float operations
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

	// Both are integers
	return e.evalIntBinaryOp(leftInt, rightInt, expr.Operator)
}

// evalIntBinaryOp performs integer binary operations (arithmetic and comparison).
func (e *Evaluator) evalIntBinaryOp(left, right int64, operator string) (interface{}, error) {
	switch operator {
	// Arithmetic operators
	case "PLUS", "+":
		return left + right, nil
	case "MINUS", "-":
		return left - right, nil
	case "MULTIPLY", "*":
		return left * right, nil
	case "DIVIDE", "/":
		if right == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return left / right, nil
	case "MODULO", "%":
		if right == 0 {
			return nil, fmt.Errorf("modulo by zero")
		}
		return left % right, nil

	// Comparison operators
	case "LESS_THAN", "<":
		return left < right, nil
	case "LESS_EQUAL", "<=":
		return left <= right, nil
	case "GREATER_THAN", ">":
		return left > right, nil
	case "GREATER_EQUAL", ">=":
		return left >= right, nil

	default:
		return nil, fmt.Errorf("unknown binary operator: %s", operator)
	}
}

// evalFloatBinaryOp performs floating-point binary operations (arithmetic and comparison).
func (e *Evaluator) evalFloatBinaryOp(left, right float64, operator string) (interface{}, error) {
	switch operator {
	// Arithmetic operators
	case "PLUS", "+":
		return left + right, nil
	case "MINUS", "-":
		return left - right, nil
	case "MULTIPLY", "*":
		return left * right, nil
	case "DIVIDE", "/":
		if right == 0.0 {
			return nil, fmt.Errorf("division by zero")
		}
		return left / right, nil
	case "MODULO", "%":
		return nil, fmt.Errorf("modulo operator not supported for floating-point numbers")

	// Comparison operators
	case "LESS_THAN", "<":
		return left < right, nil
	case "LESS_EQUAL", "<=":
		return left <= right, nil
	case "GREATER_THAN", ">":
		return left > right, nil
	case "GREATER_EQUAL", ">=":
		return left >= right, nil

	default:
		return nil, fmt.Errorf("unknown binary operator: %s", operator)
	}
}

// evalLogicalAnd evaluates logical AND with short-circuit evaluation.
func (e *Evaluator) evalLogicalAnd(expr *ast.BinaryExpression) (interface{}, error) {
	// Evaluate left operand
	leftVal, err := e.evalExpression(expr.Left)
	if err != nil {
		return nil, fmt.Errorf("error evaluating left operand of &&: %v", err)
	}

	leftBool, ok := leftVal.(bool)
	if !ok {
		return nil, fmt.Errorf("logical AND requires boolean operands, got %T", leftVal)
	}

	// Short-circuit: if left is false, return false without evaluating right
	if !leftBool {
		return false, nil
	}

	// Left is true, evaluate right
	rightVal, err := e.evalExpression(expr.Right)
	if err != nil {
		return nil, fmt.Errorf("error evaluating right operand of &&: %v", err)
	}

	rightBool, ok := rightVal.(bool)
	if !ok {
		return nil, fmt.Errorf("logical AND requires boolean operands, got %T", rightVal)
	}

	return rightBool, nil
}

// evalLogicalOr evaluates logical OR with short-circuit evaluation.
func (e *Evaluator) evalLogicalOr(expr *ast.BinaryExpression) (interface{}, error) {
	// Evaluate left operand
	leftVal, err := e.evalExpression(expr.Left)
	if err != nil {
		return nil, fmt.Errorf("error evaluating left operand of ||: %v", err)
	}

	leftBool, ok := leftVal.(bool)
	if !ok {
		return nil, fmt.Errorf("logical OR requires boolean operands, got %T", leftVal)
	}

	// Short-circuit: if left is true, return true without evaluating right
	if leftBool {
		return true, nil
	}

	// Left is false, evaluate right
	rightVal, err := e.evalExpression(expr.Right)
	if err != nil {
		return nil, fmt.Errorf("error evaluating right operand of ||: %v", err)
	}

	rightBool, ok := rightVal.(bool)
	if !ok {
		return nil, fmt.Errorf("logical OR requires boolean operands, got %T", rightVal)
	}

	return rightBool, nil
}

// evalStringBinaryOp performs string binary operations (concatenation and comparison).
func (e *Evaluator) evalStringBinaryOp(left, right string, operator string) (interface{}, error) {
	switch operator {
	// Concatenation
	case "PLUS", "+":
		return left + right, nil

	// Comparison operators (lexicographic)
	case "LESS_THAN", "<":
		return left < right, nil
	case "LESS_EQUAL", "<=":
		return left <= right, nil
	case "GREATER_THAN", ">":
		return left > right, nil
	case "GREATER_EQUAL", ">=":
		return left >= right, nil

	default:
		return nil, fmt.Errorf("operator %s not supported for strings", operator)
	}
}

// evalEquality checks if two values are equal.
// Works on any type.
func (e *Evaluator) evalEquality(left, right interface{}) bool {
	return left == right
}

// evalFunctionDef evaluates a function definition statement.
// Creates a Function value and stores it in the environment.
func (e *Evaluator) evalFunctionDef(stmt *ast.FunctionDef) error {
	// Create function value
	fn := &Function{
		Parameters: stmt.Parameters,
		Body:       stmt.Body,
	}

	// Store in environment (global scope)
	e.env.Set(stmt.Name, fn)
	return nil
}

// evalFunctionLiteral evaluates a function literal expression.
// Returns a Function value that can be assigned to variables or passed as arguments.
func (e *Evaluator) evalFunctionLiteral(expr *ast.FunctionLiteral) (interface{}, error) {
	return &Function{
		Parameters: expr.Parameters,
		Body:       expr.Body,
	}, nil
}

// returnValue is a special type used to propagate return statements up through block evaluation.
type returnValue struct {
	value interface{}
}

// evalBlock evaluates a block of statements.
// Returns the value from a return statement, or nil if no return.
func (e *Evaluator) evalBlock(block *ast.Block) (interface{}, error) {
	for _, stmt := range block.Statements {
		// Check if it's a return statement
		if retStmt, ok := stmt.(*ast.ReturnStatement); ok {
			// Evaluate return expression
			val, err := e.evalExpression(retStmt.Value)
			if err != nil {
				return nil, err
			}
			// Return the value wrapped to signal a return
			return &returnValue{value: val}, nil
		}

		// Evaluate other statements
		if err := e.evalStatement(stmt); err != nil {
			return nil, err
		}
	}

	// No return statement found
	return nil, fmt.Errorf("function must end with return statement")
}

// callUserFunction calls a user-defined function with the given arguments.
func (e *Evaluator) callUserFunction(fn *Function, args []ast.Expression) (interface{}, error) {
	// Check argument count
	if len(args) != len(fn.Parameters) {
		return nil, fmt.Errorf("function expects %d arguments, got %d",
			len(fn.Parameters), len(args))
	}

	// Evaluate arguments
	argValues := make([]interface{}, len(args))
	for i, arg := range args {
		val, err := e.evalExpression(arg)
		if err != nil {
			return nil, fmt.Errorf("error evaluating argument %d: %v", i, err)
		}
		argValues[i] = val
	}

	// Create new environment for function scope
	// Parent is current environment (typically global)
	fnEnv := NewEnvironment(e.env)

	// Bind parameters to argument values
	for i, param := range fn.Parameters {
		fnEnv.Set(param, argValues[i])
	}

	// Save current environment and switch to function environment
	savedEnv := e.env
	e.env = fnEnv
	defer func() { e.env = savedEnv }()

	// Execute function body
	result, err := e.evalBlock(fn.Body)
	if err != nil {
		return nil, err
	}

	// Unwrap return value
	if retVal, ok := result.(*returnValue); ok {
		return retVal.value, nil
	}

	// This shouldn't happen if evalBlock works correctly
	return nil, fmt.Errorf("function did not return properly")
}

// evalArrayLiteral evaluates an array literal expression.
// Returns a Go slice ([]interface{}) containing the evaluated elements.
func (e *Evaluator) evalArrayLiteral(expr *ast.ArrayLiteral) (interface{}, error) {
	elements := make([]interface{}, len(expr.Elements))

	for i, elemExpr := range expr.Elements {
		val, err := e.evalExpression(elemExpr)
		if err != nil {
			return nil, fmt.Errorf("error evaluating array element %d: %v", i, err)
		}
		elements[i] = val
	}

	return elements, nil
}

// evalIndexAccess evaluates array indexing: arr[index]
func (e *Evaluator) evalIndexAccess(expr *ast.IndexAccess) (interface{}, error) {
	// Evaluate the object being indexed
	obj, err := e.evalExpression(expr.Object)
	if err != nil {
		return nil, fmt.Errorf("error evaluating indexed object: %v", err)
	}

	// Check if it's an array
	arr, ok := obj.([]interface{})
	if !ok {
		return nil, fmt.Errorf("cannot index non-array type: %T", obj)
	}

	// Evaluate the index expression
	indexVal, err := e.evalExpression(expr.Index)
	if err != nil {
		return nil, fmt.Errorf("error evaluating index: %v", err)
	}

	// Convert index to int64
	index, ok := indexVal.(int64)
	if !ok {
		return nil, fmt.Errorf("array index must be an integer, got %T", indexVal)
	}

	// Bounds check
	if index < 0 || index >= int64(len(arr)) {
		return nil, fmt.Errorf("array index out of bounds: index %d, length %d", index, len(arr))
	}

	return arr[index], nil
}

// evalMemberAccess evaluates member access: obj.member
// For arrays, this is used to access methods like len, push, pop
func (e *Evaluator) evalMemberAccess(expr *ast.MemberAccess) (interface{}, error) {
	// Evaluate the object
	obj, err := e.evalExpression(expr.Object)
	if err != nil {
		return nil, fmt.Errorf("error evaluating object for member access: %v", err)
	}

	// Check if it's an array
	arr, ok := obj.([]interface{})
	if !ok {
		return nil, fmt.Errorf("member access only supported on arrays, got %T", obj)
	}

	// Return a callable that represents the method bound to this array
	// We'll return a special function value that knows about the array and method
	return &ArrayMethod{
		Array:  arr,
		Method: expr.Member,
		Object: expr.Object, // Keep object expression for updates
	}, nil
}

// ArrayMethod represents a method bound to an array instance
type ArrayMethod struct {
	Array  []interface{}
	Method string
	Object ast.Expression // The original object expression (for mutations)
}

// evalIndexAssignment evaluates assignment to an array index: arr[index] = value
func (e *Evaluator) evalIndexAssignment(stmt *ast.IndexAssignment) error {
	// Get the array from the environment
	arrValue, exists := e.env.Get(stmt.Name)
	if !exists {
		return fmt.Errorf("undefined variable: %s", stmt.Name)
	}

	// Navigate through the index chain to find the target array and final index
	currentArray := arrValue
	for i := 0; i < len(stmt.Indices)-1; i++ {
		// Evaluate this index
		indexVal, err := e.evalExpression(stmt.Indices[i])
		if err != nil {
			return fmt.Errorf("error evaluating index %d: %v", i, err)
		}

		index, ok := indexVal.(int64)
		if !ok {
			return fmt.Errorf("array index must be an integer, got %T", indexVal)
		}

		// Get the nested array
		arr, ok := currentArray.([]interface{})
		if !ok {
			return fmt.Errorf("cannot index non-array type: %T", currentArray)
		}

		if index < 0 || index >= int64(len(arr)) {
			return fmt.Errorf("array index out of bounds: index %d, length %d", index, len(arr))
		}

		currentArray = arr[index]
	}

	// Now currentArray is the final array to modify
	arr, ok := currentArray.([]interface{})
	if !ok {
		return fmt.Errorf("cannot index non-array type: %T", currentArray)
	}

	// Evaluate the final index
	finalIndexVal, err := e.evalExpression(stmt.Indices[len(stmt.Indices)-1])
	if err != nil {
		return fmt.Errorf("error evaluating final index: %v", err)
	}

	finalIndex, ok := finalIndexVal.(int64)
	if !ok {
		return fmt.Errorf("array index must be an integer, got %T", finalIndexVal)
	}

	// Bounds check
	if finalIndex < 0 || finalIndex >= int64(len(arr)) {
		return fmt.Errorf("array index out of bounds: index %d, length %d", finalIndex, len(arr))
	}

	// Evaluate the value to assign
	value, err := e.evalExpression(stmt.Value)
	if err != nil {
		return fmt.Errorf("error evaluating assignment value: %v", err)
	}

	// Perform the assignment
	arr[finalIndex] = value

	// Note: The slice is modified in place, so we don't need to update the environment
	// unless this was a single-level array (no nested indices)
	if len(stmt.Indices) == 1 {
		e.env.Set(stmt.Name, arr)
	}

	return nil
}
