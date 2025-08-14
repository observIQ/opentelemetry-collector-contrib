// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package verify

import (
	"errors"
	"fmt"
)

// SigstoreConfig holds settings used for verifying package signatures with Sigstore.
// You can read more about Cosign and signing here.
// https://docs.sigstore.dev/cosign/signing/overview/
type SigstoreConfig struct {
	// TODO: The Fulcio root certificate can be specified via SIGSTORE_ROOT_FILE for now
	// But we should add it as a config option.
	// https://github.com/open-telemetry/opentelemetry-collector-contrib/issues/35931

	// github_workflow_repository defines the expected repository field on the sigstore certificate.
	CertGithubWorkflowRepository string `mapstructure:"github_workflow_repository"`

	// Identities is a list of valid identities to use when verifying the agent.
	// Only one needs to match the identity on the certificate for signature verification to pass.
	Identities []SigstoreIdentity `mapstructure:"identities"`
}

func (c *SigstoreConfig) Validate() error {
	for i, ident := range c.Identities {
		if err := ident.Validate(); err != nil {
			return fmt.Errorf("agent::identities[%d]: %w", i, err)
		}
	}
	return nil
}

// SigstoreIdentity represents an Issuer/Subject pair that identifies
// the signer of the agent. This allows restricting the valid signers, such
// that only very specific sources are trusted as signers for an agent.
// You can read more about Cosign and signing here.
// Issuer and Subject are used to strictly match their values.
// IssuerRegExp and SubjectRegExp can be used instead to match the values using a regular expression.
// These are the values that are used to verify the identity of the signer.
// https://docs.sigstore.dev/cosign/signing/overview/
type SigstoreIdentity struct {
	// Issuer is the OIDC Issuer for the identity.
	Issuer string `mapstructure:"issuer"`
	// Subject is the OIDC Subject for the identity.
	Subject string `mapstructure:"subject"`
	// IssuerRegExp is a regular expression for matching the OIDC Issuer for the identity.
	IssuerRegExp string `mapstructure:"issuer_regex"`
	// SubjectRegExp is a regular expression for matching the OIDC Subject for the identity.
	SubjectRegExp string `mapstructure:"subject_regex"`
}

func (i SigstoreIdentity) Validate() error {
	if i.Issuer != "" && i.IssuerRegExp != "" {
		return errors.New("cannot specify both issuer and issuer_regex")
	}
	if i.Subject != "" && i.SubjectRegExp != "" {
		return errors.New("cannot specify both subject and subject_regex")
	}
	if i.Issuer == "" && i.IssuerRegExp == "" {
		return errors.New("must specify one of issuer or issuer_regex")
	}
	if i.Subject == "" && i.SubjectRegExp == "" {
		return errors.New("must specify one of subject or subject_regex")
	}
	return nil
}

func defaultSigstoreConfig() *SigstoreConfig {
	return &SigstoreConfig{
		CertGithubWorkflowRepository: "open-telemetry/opentelemetry-collector-releases",
		Identities: []SigstoreIdentity{
			{
				Issuer:        "https://token.actions.githubusercontent.com",
				SubjectRegExp: `^https://github.com/open-telemetry/opentelemetry-collector-releases/.github/workflows/base-release.yaml@refs/tags/[^/]*$`,
			},
		},
	}
}
