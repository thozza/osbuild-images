package check

import (
	"fmt"
	"log"

	"github.com/osbuild/image-builder/internal/buildconfig"
	"gopkg.in/yaml.v3"
)

type ServiceListParams struct {
	Services []string `yaml:"services"`
}

func serviceListFromYAML(node *yaml.Node) (CheckParams, error) {
	var p ServiceListParams
	if err := node.Decode(&p); err != nil {
		return nil, err
	}
	if len(p.Services) == 0 {
		return nil, fmt.Errorf("'services' list must not be empty")
	}
	return p, nil
}

func init() {
	RegisterCheckWithParams(RegisteredCheck{
		Meta: &Metadata{
			Name:                   "srv-enabled",
			RequiresBlueprint:      true,
			RequiresCustomizations: true,
		},
		ParamFunc:       servicesEnabledCheck,
		FromBuildConfig: servicesEnabledFromConfig,
		FromYAML:        serviceListFromYAML,
	})
}

func servicesEnabledFromConfig(c *buildconfig.BuildConfig) (CheckParams, error) {
	if c == nil || c.Blueprint == nil || c.Blueprint.Customizations == nil {
		return nil, nil
	}
	services := c.Blueprint.Customizations.Services
	if services == nil || len(services.Enabled) == 0 {
		return nil, nil
	}
	return ServiceListParams{Services: services.Enabled}, nil
}

func servicesEnabledCheck(meta *Metadata, params CheckParams) error {
	p, ok := params.(ServiceListParams)
	if !ok {
		return Fail("invalid params type")
	}

	for _, service := range p.Services {
		log.Printf("Checking enabled service: %s\n", service)
		state, _, _, err := ExecString("systemctl", "is-enabled", service)
		if err != nil {
			return Fail("service is not enabled:", service, "error:", err)
		}
		if state != "enabled" {
			return Fail("service is not enabled:", service, "state:", state)
		}
		log.Printf("Service was enabled service=%s state=%s\n", service, state)
	}

	return Pass()
}
