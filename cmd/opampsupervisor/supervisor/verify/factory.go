package verify

import (
	"context"
	"fmt"

	"github.com/sigstore/cosign/v2/cmd/cosign/cli/fulcio"
	"github.com/sigstore/cosign/v2/pkg/cosign"
	"github.com/sigstore/rekor/pkg/client"

	"github.com/open-telemetry/opentelemetry-collector-contrib/cmd/opampsupervisor/supervisor/config"
)

// Config holds settings for building a signature verifier.
type Config struct {
	Signature config.AgentSignature
}

// NewDefaultVerifier returns the default Sigstore-based verifier.
func NewDefaultVerifier(cfg Config) (SignatureVerifier, error) {
	checkOpts, err := createCosignCheckOpts(cfg.Signature)
	if err != nil {
		return nil, err
	}
	return &SigstoreVerifier{checkOpts: checkOpts}, nil
}

func createCosignCheckOpts(signatureOpts config.AgentSignature) (*cosign.CheckOpts, error) {
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
