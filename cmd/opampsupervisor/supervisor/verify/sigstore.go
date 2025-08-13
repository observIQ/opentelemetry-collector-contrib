package verify

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/sigstore/cosign/v2/pkg/cosign"
	"github.com/sigstore/cosign/v2/pkg/oci/static"
)

// SigstoreVerifier verifies packages using Sigstore and Rekor.
type SigstoreVerifier struct {
	checkOpts *cosign.CheckOpts
}

// Verify verifies the package at pkgPath using the provided signature.
func (v *SigstoreVerifier) Verify(ctx context.Context, pkgPath, signature string) error {
	packageBytes, err := os.ReadFile(pkgPath)
	if err != nil {
		return fmt.Errorf("read package bytes: %w", err)
	}
	b64Cert, b64Sig, err := parsePackageSignature(signature)
	if err != nil {
		return fmt.Errorf("parse package signature: %w", err)
	}
	return verifyPackageSignature(ctx, v.checkOpts, packageBytes, b64Cert, b64Sig)
}

func (v *SigstoreVerifier) Name() string { return "sigstore" }

func parsePackageSignature(signature string) (b64Cert, b64Signature string, err error) {
	parts := strings.SplitN(signature, " ", 2)
	if len(parts) != 2 {
		return "", "", errors.New("signature must be formatted as a space separated cert and signature")
	}
	return parts[0], parts[1], nil
}

func verifyPackageSignature(ctx context.Context, checkOpts *cosign.CheckOpts, packageBytes []byte, b64Cert, b64Signature string) error {
	decodedCert, err := base64.StdEncoding.DecodeString(b64Cert)
	if err != nil {
		return fmt.Errorf("b64 decode cert: %w", err)
	}

	ociSig, err := static.NewSignature(packageBytes, b64Signature, static.WithCertChain(decodedCert, nil))
	if err != nil {
		return fmt.Errorf("create signature: %w", err)
	}

	_, err = cosign.VerifyBlobSignature(ctx, ociSig, checkOpts)
	if err != nil {
		return fmt.Errorf("verify blob: %w", err)
	}

	return nil
}
