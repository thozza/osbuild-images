package check

import (
	"log"

	"github.com/osbuild/image-builder/internal/buildconfig"
)

func init() {
	RegisterCheckWithParams(RegisteredCheck{
		Meta: &Metadata{
			Name: "fw-srv-disabled",
		},
		ParamFunc:       firewallServicesDisabledCheck,
		FromBuildConfig: firewallServicesDisabledFromConfig,
		FromYAML:        serviceListFromYAML,
	})
}

func firewallServicesDisabledFromConfig(c *buildconfig.BuildConfig) (CheckParams, error) {
	if c == nil || c.Blueprint == nil || c.Blueprint.Customizations == nil {
		return nil, nil
	}
	fw := c.Blueprint.Customizations.Firewall
	if fw == nil || fw.Services == nil || len(fw.Services.Disabled) == 0 {
		return nil, nil
	}
	return ServiceListParams{Services: fw.Services.Disabled}, nil
}

func firewallServicesDisabledCheck(meta *Metadata, params CheckParams) error {
	p, ok := params.(ServiceListParams)
	if !ok {
		return Fail("invalid params type")
	}

	for _, service := range p.Services {
		log.Printf("Checking disabled firewall service: %s\n", service)
		// NOTE: sudo works here without password because we test this only on ami
		// initialised with cloud-init, which sets sudo NOPASSWD for the user
		state, _, code, err := ExecString("sudo", "firewall-cmd", "--query-service="+service)
		if err != nil && code != 1 { // 1 is the exit code for "service not found"
			return Fail("problem checking firewall service:", service, "error:", err)
		}
		if state == "yes" {
			return Fail("firewall service is not disabled:", service, "state:", state)
		}
		log.Printf("Firewall service was disabled service=%s state=%s\n", service, state)
	}

	return Pass()
}
