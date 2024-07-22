package supervisor

import (
	"bytes"
	_ "embed"
	"text/template"
	"time"

	"github.com/open-telemetry/opamp-go/protobufs"
)

var (
	//go:embed templates/bootstrap_pipeline.yaml
	bootstrapConfTpl string

	//go:embed templates/extraconfig.yaml
	extraConfigTpl string

	//go:embed templates/opampextension.yaml
	opampextensionTpl string

	//go:embed templates/owntelemetry.yaml
	ownTelemetryTpl string
)

var (
	parsedBootstrapConfTpl  = template.Must(template.New("bootstrap").Parse(bootstrapConfTpl))
	parsedExtraConfigTpl    = template.Must(template.New("extraconfig").Parse(extraConfigTpl))
	parsedOpampextensionTpl = template.Must(template.New("opampextension").Parse(opampextensionTpl))
	parsedOwnTelemetryTpl   = template.Must(template.New("owntelemetry").Parse(ownTelemetryTpl))
)

func composeBootstrapConfig(instanceID string, supervisorPort int) ([]byte, error) {
	var cfg bytes.Buffer
	err := parsedBootstrapConfTpl.Execute(&cfg, map[string]any{
		"InstanceUid":    instanceID,
		"SupervisorPort": supervisorPort,
	})
	if err != nil {
		return nil, err
	}

	return cfg.Bytes(), nil
}

func composeExtraConfig(healthcheckEndpoint string, ad *protobufs.AgentDescription) ([]byte, error) {
	resourceAttrs := map[string]string{}
	for _, attr := range ad.GetIdentifyingAttributes() {
		resourceAttrs[attr.Key] = attr.Value.GetStringValue()
	}
	for _, attr := range ad.GetNonIdentifyingAttributes() {
		resourceAttrs[attr.Key] = attr.Value.GetStringValue()
	}

	var cfg bytes.Buffer
	err := parsedExtraConfigTpl.Execute(&cfg, map[string]any{
		"Healthcheck":        healthcheckEndpoint,
		"ResourceAttributes": resourceAttrs,
	})
	if err != nil {
		return nil, err
	}

	return cfg.Bytes(), nil
}

func composeOpAMPExtensionConfig(instanceID string, supervisorPort, pid int, orphanPollInterval time.Duration) ([]byte, error) {
	var cfg bytes.Buffer
	err := parsedOpampextensionTpl.Execute(&cfg, map[string]any{
		"InstanceUid":      instanceID,
		"SupervisorPort":   supervisorPort,
		"PID":              pid,
		"PPIDPollInterval": orphanPollInterval,
	})
	if err != nil {
		return nil, err
	}

	return cfg.Bytes(), nil
}

func composeOwnTelemetryConfig(destinationEndpoint string) ([]byte, error) {
	var cfg bytes.Buffer
	err := parsedOwnTelemetryTpl.Execute(&cfg, map[string]any{
		"MetricsEndpoint": destinationEndpoint,
	})
	if err != nil {
		return nil, err
	}

	return cfg.Bytes(), nil
}
