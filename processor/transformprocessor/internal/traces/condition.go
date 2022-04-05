// Copyright  The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package traces // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor/internal/traces"

import (
	"fmt"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/transformprocessor/internal/common"
)

type condFunc = func(ctx spanTransformContext) bool

var alwaysTrue = func(ctx spanTransformContext) bool {
	return true
}

func newConditionEvaluator(cond *common.Condition, functions map[string]interface{}) (condFunc, error) {
	if cond == nil {
		return alwaysTrue, nil
	}
	left, err := newGetter(cond.Left, functions)
	if err != nil {
		return nil, err
	}
	right, err := newGetter(cond.Right, functions)
	// TODO(anuraaga): Check if both left and right are literals and const-evaluate
	if err != nil {
		return nil, err
	}

	switch cond.Op {
	case "==":
		return func(ctx spanTransformContext) bool {
			a := left.get(ctx)
			b := right.get(ctx)
			return a == b
		}, nil
	case "!=":
		return func(ctx spanTransformContext) bool {
			a := left.get(ctx)
			b := right.get(ctx)
			return a != b
		}, nil
	}

	return nil, fmt.Errorf("unrecognized boolean operation %v", cond.Op)
}
