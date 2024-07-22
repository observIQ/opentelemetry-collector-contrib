package supervisor

import (
	"fmt"
	"sort"
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/open-telemetry/opamp-go/protobufs"
)

// configRenderer defines an interface for a struct that renders a config into a single config file
type configRenderer interface {
	// Render renders the effective config given the instanceID, remote config, connection settings,
	// and agent description
	Render(
		instanceID string,
		config *protobufs.AgentRemoteConfig,
		telSettings *protobufs.TelemetryConnectionSettings,
		agentDesc *protobufs.AgentDescription,
	) ([]byte, error)
}

type otelConfigRenderer struct {
	healthcheckEndpoint string
	supervisorPort      int
	orphanPollInterval  time.Duration
	pidProvider         pidProvider
}

func newConfigRenderer(
	healthcheckEndpoint string,
	supervisorPort int,
	orphanPollInterval time.Duration,
	pidProvider pidProvider,
) otelConfigRenderer {
	return otelConfigRenderer{
		healthcheckEndpoint: healthcheckEndpoint,
		supervisorPort:      supervisorPort,
		orphanPollInterval:  orphanPollInterval,
		pidProvider:         pidProvider,
	}
}

func (o otelConfigRenderer) Render(
	instanceID string,
	config *protobufs.AgentRemoteConfig,
	telSettings *protobufs.TelemetryConnectionSettings,
	agentDesc *protobufs.AgentDescription,
) ([]byte, error) {
	var k = koanf.New("::")

	if err := o.mergeRemoteConfig(k, config); err != nil {
		return nil, fmt.Errorf("merge remote config: %w", err)
	}

	// Merge own metrics config.
	if err := o.mergeOwnTelemetryConfig(k, telSettings); err != nil {
		return nil, fmt.Errorf("merge remote config: %w", err)
	}

	// Merge local config last since it has the highest precedence.
	if err := o.mergeExtraConfig(k, agentDesc); err != nil {
		return nil, fmt.Errorf("merge remote config: %w", err)
	}

	if err := o.mergeOpAMPExtensionConfig(k, instanceID); err != nil {
		return nil, err
	}

	// The merged final result is our new merged config.
	return k.Marshal(yaml.Parser())
}

func (o otelConfigRenderer) mergeRemoteConfig(k *koanf.Koanf, config *protobufs.AgentRemoteConfig) error {
	c := config.GetConfig()
	if c == nil {
		// TODO: Merge bootstrap config if agent remote config is not present
		return nil
	}

	// Sort to make sure the order of merging is stable.
	var names []string
	var hasInstanceConfig bool
	for name := range c.ConfigMap {
		if name == "" {
			// skip instance config
			hasInstanceConfig = true
			continue
		}
		names = append(names, name)
	}

	sort.Strings(names)

	if hasInstanceConfig {
		// Append instance config as the last item.
		names = append(names, "")
	}

	// Merge received configs.
	for _, name := range names {
		item := c.ConfigMap[name]
		if item == nil {
			continue
		}
		var k2 = koanf.New("::")
		err := k2.Load(rawbytes.Provider(item.Body), yaml.Parser())
		if err != nil {
			return fmt.Errorf("cannot parse config named %s: %w", name, err)
		}

		err = k.Merge(k2)
		if err != nil {
			return fmt.Errorf("cannot merge config named %s: %w", name, err)
		}
	}

	return nil
}

func (o otelConfigRenderer) mergeOwnTelemetryConfig(k *koanf.Koanf, set *protobufs.TelemetryConnectionSettings) error {
	if set == nil {
		return nil
	}

	ownMetricsCfg, err := composeOwnTelemetryConfig(set.DestinationEndpoint)
	if err != nil {
		return fmt.Errorf("compose own telemetry config: %w", err)
	}

	if err := k.Load(rawbytes.Provider(ownMetricsCfg), yaml.Parser(), koanf.WithMergeFunc(configMergeFunc)); err != nil {
		return fmt.Errorf("load own telemetry config: %w", err)
	}

	return nil
}

func (o otelConfigRenderer) mergeExtraConfig(k *koanf.Koanf, agentDescription *protobufs.AgentDescription) error {
	extraConfig, err := composeExtraConfig(o.healthcheckEndpoint, agentDescription)
	if err != nil {
		return fmt.Errorf("compose extra config: %w", err)
	}

	if err = k.Load(rawbytes.Provider(extraConfig), yaml.Parser(), koanf.WithMergeFunc(configMergeFunc)); err != nil {
		return fmt.Errorf("load extra config: %w", err)
	}

	return nil
}

func (o otelConfigRenderer) mergeOpAMPExtensionConfig(k *koanf.Koanf, instanceID string) error {
	opampConfig, err := composeOpAMPExtensionConfig(instanceID, o.supervisorPort, o.pidProvider.PID(), o.orphanPollInterval)
	if err != nil {
		return fmt.Errorf("compose opamp config: %w", err)
	}

	if err = k.Load(rawbytes.Provider(opampConfig), yaml.Parser(), koanf.WithMergeFunc(configMergeFunc)); err != nil {
		return fmt.Errorf("load opamp config: %w", err)
	}

	return nil
}
