// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package verify

import (
	"context"
	"fmt"

	"github.com/sigstore/cosign/v2/cmd/cosign/cli/fulcio"
	"github.com/sigstore/cosign/v2/pkg/cosign"
	"github.com/sigstore/rekor/pkg/client"
	"go.opentelemetry.io/collector/component"
)

// SigstoreVerifierBuilder builds Sigstore-based verifiers.
type SigstoreVerifierBuilder struct{}

// NewDefaultBuilder returns the default Sigstore-based builder.
func NewDefaultBuilder() SignatureVerifierBuilder { return SigstoreVerifierBuilder{} }

// Config returns the Sigstore verifier config struct.
func (SigstoreVerifierBuilder) Config() component.Config { return defaultSigstoreConfig() }

// NewVerifier builds a Sigstore verifier using the provided config.
func (SigstoreVerifierBuilder) NewVerifier(cfg component.Config) (SignatureVerifier, error) {
	opts, err := createCosignCheckOpts(cfg.(*SigstoreConfig))
	if err != nil {
		return nil, fmt.Errorf("create check opts: %w", err)
	}

	return &SigstoreVerifier{checkOpts: opts}, nil
}

// createCosignCheckOpts creates a cosign.CheckOpts from the signature options.
// These options provide information needed to verify the signature of the package.
// The options consist of public Fulcio certificates to verify the identity of the signature,
// a Rekor client to verify the integrity of the signature against a transparency log,
// and a set of identities that the signature must match. More information about the
// cosign.CheckOpts can be found in the specification (../specification/README.md#collector-executable-updates-flow).
func createCosignCheckOpts(signatureOpts *SigstoreConfig) (*cosign.CheckOpts, error) {
	rootCerts, err := fulcio.GetRoots()
	if err != nil {
		return nil, fmt.Errorf("fetch root certs: %w", err)
	}

	intermediateCerts, err := fulcio.GetIntermediates()
	if err != nil {
		return nil, fmt.Errorf("fetch intermediate certs: %w", err)
	}

	rekorClient, err := client.GetRekorClient("https://rekor.sigstore.dev")
	if err != nil {
		return nil, fmt.Errorf("create rekot client: %w", err)
	}

	rekorKeys, err := cosign.GetRekorPubs(context.Background())
	if err != nil {
		return nil, fmt.Errorf("get rekor public keys: %w", err)
	}

	ctLogPubKeys, err := cosign.GetCTLogPubs(context.Background())
	if err != nil {
		return nil, fmt.Errorf("get CT log public keys: %w", err)
	}

	identities := make([]cosign.Identity, 0, len(signatureOpts.Identities))
	for _, ident := range signatureOpts.Identities {
		identities = append(identities, cosign.Identity{
			Issuer:        ident.Issuer,
			IssuerRegExp:  ident.IssuerRegExp,
			Subject:       ident.Subject,
			SubjectRegExp: ident.SubjectRegExp,
		})
	}

	return &cosign.CheckOpts{
		RootCerts:                    rootCerts,
		IntermediateCerts:            intermediateCerts,
		CertGithubWorkflowRepository: signatureOpts.CertGithubWorkflowRepository,
		Identities:                   identities,
		RekorClient:                  rekorClient,
		RekorPubKeys:                 rekorKeys,
		CTLogPubKeys:                 ctLogPubKeys,
	}, nil
}
