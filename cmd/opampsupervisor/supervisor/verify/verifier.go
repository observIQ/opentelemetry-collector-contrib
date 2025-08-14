// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package verify

import (
	"context"

	"go.opentelemetry.io/collector/component"
)

// SignatureVerifier verifies the signature of a package using the provided
// package and signature bytes.
type SignatureVerifier interface {
	Verify(ctx context.Context, packageBytes, signature []byte) error
	Name() string
}

type SignatureVerifierBuilder interface {
	Config() component.Config
	NewVerifier(component.Config) (SignatureVerifier, error)
}
