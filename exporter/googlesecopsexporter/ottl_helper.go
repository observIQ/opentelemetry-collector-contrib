// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package googlesecopsexporter

import (
	"context"
	"errors"
	"fmt"

	"go.opentelemetry.io/collector/component"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/contexts/ottllog"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/ottlfuncs"
)

// ottlLogRecordExpression evaluates an OTTL expression, returning a resultant value.
type ottlLogRecordExpression struct {
	statement *ottl.Statement[*ottllog.TransformContext]
}

// Execute executes the expression with the given context, returning the value of the expression.
func (e *ottlLogRecordExpression) Execute(ctx context.Context, tCtx *ottllog.TransformContext) (any, error) {
	val, _, err := e.statement.Execute(ctx, tCtx)
	return val, err
}

// newOTTLLogRecordExpression creates a new OTTL expression for log records.
// The expression is wrapped in a value function, so only converter functions and target expressions can be used.
func newOTTLLogRecordExpression(expression string, set component.TelemetrySettings) (*ottlLogRecordExpression, error) {
	// Wrap the expression in the "value" function, since the ottl grammar expects a function first.
	statementStr := fmt.Sprintf("value(%s)", expression)
	statement, err := newOTTLLogRecordStatement(statementStr, set)
	if err != nil {
		return nil, err
	}

	return &ottlLogRecordExpression{
		statement: statement,
	}, nil
}

// newOTTLLogRecordStatement parses the given statement into an ottl.Statement for a log transform context.
func newOTTLLogRecordStatement(statementStr string, set component.TelemetrySettings) (*ottl.Statement[*ottllog.TransformContext], error) {
	parser, err := ottllog.NewParser(logRecordFunctions(), set)
	if err != nil {
		return nil, err
	}

	statement, err := parser.ParseStatement(statementStr)
	if err != nil {
		return nil, err
	}

	return statement, nil
}

// logRecordFunctions returns the list of available functions for OTTL statements.
// We include all the converter functions here (functions that do not edit telemetry),
// as well as a custom value function.
func logRecordFunctions() map[string]ottl.Factory[*ottllog.TransformContext] {
	valueFactory := newValueFactory()

	factories := ottlfuncs.StandardConverters[*ottllog.TransformContext]()
	factories[valueFactory.Name()] = valueFactory

	return factories
}

type valueArguments struct {
	Target ottl.Getter[*ottllog.TransformContext] `ottlarg:"0"`
}

// newValueFactory returns a factory for the value function, which returns the value of its first argument.
// We need this function because OTTL does not allow direct access to fields on the context, instead
// expecting a function as the first token.
func newValueFactory() ottl.Factory[*ottllog.TransformContext] {
	return ottl.NewFactory("value", &valueArguments{}, createValueFunction)
}

func createValueFunction(_ ottl.FunctionContext, a ottl.Arguments) (ottl.ExprFunc[*ottllog.TransformContext], error) {
	args, ok := a.(*valueArguments)
	if !ok {
		return nil, errors.New("valueFactory args must be of type *valueArguments")
	}

	return valueFn(args)
}

func valueFn(c *valueArguments) (ottl.ExprFunc[*ottllog.TransformContext], error) {
	return func(ctx context.Context, tCtx *ottllog.TransformContext) (any, error) {
		return c.Target.Get(ctx, tCtx)
	}, nil
}
