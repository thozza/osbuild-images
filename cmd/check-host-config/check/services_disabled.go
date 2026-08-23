package check

import (
	"log"

	"github.com/osbuild/image-builder/internal/buildconfig"
)

func init() {
	RegisterCheck(RegisteredCheck{
		Meta: &Metadata{
			Name: "srv-disabled",
		},
		Func:            servicesDisabledCheck,
		FromBuildConfig: servicesDisabledFromConfig,
		FromYAML:        serviceListFromYAML,
	})
}

func servicesDisabledFromConfig(c *buildconfig.BuildConfig) (CheckParams, error) {
	if c == nil || c.Blueprint == nil || c.Blueprint.Customizations == nil {
		return nil, nil
	}
	services := c.Blueprint.Customizations.Services
	if services == nil || len(services.Disabled) == 0 {
		return nil, nil
	}
	return ServiceListParams{Services: services.Disabled}, nil
}

func servicesDisabledCheck(meta *Metadata, params CheckParams) error {
	p, ok := params.(ServiceListParams)
	if !ok {
		return Fail("invalid params type")
	}

	for _, service := range p.Services {
		log.Printf("Checking disabled service: %s\n", service)
		// systemctl is-enabled returns non-zero exit code for disabled services,
		// but still outputs "disabled", so we check the output regardless of error
		state, _, _, err := ExecString("systemctl", "is-enabled", service)
		if state == "" && err != nil {
			return Fail("service is not disabled:", service, "error:", err)
		}

		if state != "disabled" {
			return Fail("service is not disabled:", service, "state:", state)
		}
		log.Printf("Service was disabled service=%s state=%s\n", service, state)
	}

	return Pass()
}
