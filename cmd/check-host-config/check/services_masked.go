package check

import (
	"strings"

	"github.com/osbuild/image-builder/internal/buildconfig"
)

func init() {
	RegisterCheck(RegisteredCheck{
		Meta: &Metadata{
			Name: "srv-masked",
		},
		Func:       servicesMaskedCheck,
		FromBuildConfig: servicesMaskedFromConfig,
		FromYAML:        serviceListFromYAML,
	})
}

func servicesMaskedFromConfig(c *buildconfig.BuildConfig) (CheckParams, error) {
	if c == nil || c.Blueprint == nil || c.Blueprint.Customizations == nil {
		return nil, nil
	}
	services := c.Blueprint.Customizations.Services
	if services == nil || len(services.Masked) == 0 {
		return nil, nil
	}
	return ServiceListParams{Services: services.Masked}, nil
}

func servicesMaskedCheck(meta *Metadata, params CheckParams) error {
	p, ok := params.(ServiceListParams)
	if !ok {
		return Fail("invalid params type")
	}

	stdout, _, _, err := ExecString("systemctl", "list-unit-files", "--state=masked")
	if err != nil {
		return Fail("failed to list masked services:", err)
	}

	for _, service := range p.Services {
		// Prevent false positives by appending suffix if it is not present
		if !strings.Contains(service, ".") {
			service = service + ".service"
		}

		if !strings.Contains(stdout, service) {
			return Fail("service is not masked:", service)
		}
	}

	return Pass()
}
