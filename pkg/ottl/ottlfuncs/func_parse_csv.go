// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package ottlfuncs // import "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl/ottlfuncs"

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go.opentelemetry.io/collector/pdata/pcommon"

	"github.com/open-telemetry/opentelemetry-collector-contrib/internal/coreinternal/parseutils"
	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
)

const (
	parseCSVModeStrict       = "strict"
	parseCSVModeLazyQuotes   = "lazyQuotes"
	parseCSVModeIgnoreQuotes = "ignoreQuotes"
)

const (
	parseCSVDefaultDelimiter = ','
	parseCSVDefaultMode      = parseCSVModeStrict
)

type ParseCSVArguments[K any] struct {
	Target          ottl.StringGetter[K]
	Header          ottl.StringGetter[K]
	Delimiter       ottl.Optional[string]
	HeaderDelimiter ottl.Optional[string]
	Mode            ottl.Optional[string]
}

func (p ParseCSVArguments[K]) validate() error {
	if !p.Delimiter.IsEmpty() {
		if len([]rune(p.Delimiter.Get())) != 1 {
			return errors.New("delimiter must be a single character")
		}
	}

	if !p.HeaderDelimiter.IsEmpty() {
		if len([]rune(p.HeaderDelimiter.Get())) != 1 {
			return errors.New("header_delimiter must be a single character")
		}
	}

	if !p.Mode.IsEmpty() {
		mode := p.Mode.Get()
		switch mode {
		case parseCSVModeStrict, parseCSVModeLazyQuotes, parseCSVModeIgnoreQuotes: //OK
		default:
			return fmt.Errorf("unknown mode: %s", mode)
		}
	}

	return nil
}

func NewParseCSVFactory[K any]() ottl.Factory[K] {
	return ottl.NewFactory("ParseCSV", &ParseCSVArguments[K]{}, createParseCSVFunction[K])
}

func createParseCSVFunction[K any](_ ottl.FunctionContext, oArgs ottl.Arguments) (ottl.ExprFunc[K], error) {
	args, ok := oArgs.(*ParseCSVArguments[K])
	if !ok {
		return nil, fmt.Errorf("ParseCSVFactory args must be of type *ParseCSVArguments[K]")
	}

	if err := args.validate(); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	delimiter := parseCSVDefaultDelimiter
	if !args.Delimiter.IsEmpty() {
		delimiter = []rune(args.Delimiter.Get())[0]
	}

	// headerDelimiter defaults to the chosen delimter,
	// since in most cases headerDelimiter == delmiter.
	headerDelimiter := string(delimiter)
	if !args.HeaderDelimiter.IsEmpty() {
		headerDelimiter = args.HeaderDelimiter.Get()
	}

	mode := parseCSVDefaultMode
	if !args.Mode.IsEmpty() {
		mode = args.Mode.Get()
	}

	return parseCSV(args.Target, args.Header, delimiter, headerDelimiter, mode), nil
}

func parseCSV[K any](target, header ottl.StringGetter[K], delimiter rune, headerDelimiter string, mode string) ottl.ExprFunc[K] {
	return func(ctx context.Context, tCtx K) (any, error) {
		targetStr, err := target.Get(ctx, tCtx)
		if err != nil {
			return nil, err
		}

		headerStr, err := header.Get(ctx, tCtx)
		if err != nil {
			return nil, err
		}

		if headerStr == "" {
			return nil, errors.New("headers must not be an empty string")
		}

		headers := strings.Split(headerStr, headerDelimiter)

		var fields []string
		switch mode {
		case parseCSVModeStrict, parseCSVModeLazyQuotes:
			lazyQuotes := mode == parseCSVModeLazyQuotes
			fields, err = parseutils.ReadCSVRow(targetStr, delimiter, lazyQuotes)
			if err != nil {
				return nil, err
			}
		case parseCSVModeIgnoreQuotes:
			// Ignoring quotes makes CSV parseable with just string.Split
			fields = strings.Split(targetStr, string([]rune{delimiter}))
		}

		headersToFields, err := parseutils.MapCSVHeaders(headers, fields)
		if err != nil {
			return nil, fmt.Errorf("map csv headers: %w", err)
		}

		pMap := pcommon.NewMap()
		err = pMap.FromRaw(headersToFields)
		return pMap, err
	}
}
