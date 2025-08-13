package verify

import "context"

// SignatureVerifier verifies the signature of a package at pkgPath using
// an accompanying signature string or file path, depending on the caller.
type SignatureVerifier interface {
	Verify(ctx context.Context, pkgPath string, signature string) error
	Name() string
}
