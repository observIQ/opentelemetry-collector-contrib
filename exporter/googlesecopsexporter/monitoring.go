// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package googlesecopsexporter

import "go.opentelemetry.io/otel/attribute"

var (
	attrError = "error"

	attrErrorNone    attribute.KeyValue = attribute.String(attrError, "none")
	attrErrorUnknown attribute.KeyValue = attribute.String(attrError, "unknown")
)
