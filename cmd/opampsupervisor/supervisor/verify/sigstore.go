// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package verify

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/sigstore/cosign/v2/pkg/cosign"
	"github.com/sigstore/cosign/v2/pkg/oci/static"
)

// SigstoreVerifier verifies packages using Sigstore and Rekor.
type SigstoreVerifier struct {
	checkOpts *cosign.CheckOpts
}

// Verify verifies the provided package bytes against the signature.
func (v *SigstoreVerifier) Verify(ctx context.Context, packageBytes, signature []byte) error {
	b64Cert, b64Sig, err := parsePackageSignature(signature)
	if err != nil {
		return fmt.Errorf("parse package signature: %w", err)
	}
	return verifyPackageSignature(ctx, v.checkOpts, packageBytes, b64Cert, b64Sig)
}

func (v *SigstoreVerifier) Name() string { return "sigstore" }

func parsePackageSignature(signature []byte) (b64Cert, b64Signature []byte, err error) {
	splitSignature := bytes.SplitN(signature, []byte(" "), 2)
	if len(splitSignature) != 2 {
		return nil, nil, errors.New("signature must be formatted as a space separated cert and signature")
	}

	return splitSignature[0], splitSignature[1], nil
}

// verifyPackageSignature verifies that the b64Signature is a valid signature for packageBytes.
// b64Cert is used to validate the identity of the signature against the identities in the
// provided checkOpts.
func verifyPackageSignature(ctx context.Context, checkOpts *cosign.CheckOpts, packageBytes, b64Cert, b64Signature []byte) error {
	decodedCert, err := base64.StdEncoding.AppendDecode(nil, b64Cert)
	if err != nil {
		return fmt.Errorf("b64 decode cert: %w", err)
	}

	ociSig, err := static.NewSignature(packageBytes, string(b64Signature), static.WithCertChain(decodedCert, nil))
	if err != nil {
		return fmt.Errorf("create signature: %w", err)
	}

	// VerifyBlobSignature uses the provided checkOpts to verify the signature of the package.
	// Specifically it uses the public Fulcio certificates to verify the identity of the signature and
	// a Rekor client to verify the validity of the signature against a transparency log.
	_, err = cosign.VerifyBlobSignature(ctx, ociSig, checkOpts)
	if err != nil {
		return fmt.Errorf("verify blob: %w", err)
	}

	return nil
}
