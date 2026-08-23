package check

import (
	"log"

	"github.com/osbuild/image-builder/internal/buildconfig"
)

func init() {
	RegisterCheck(RegisteredCheck{
		Meta: &Metadata{
			Name: "fw-srv-enabled",
		},
		Func:       firewallServicesEnabledCheck,
		FromBuildConfig: firewallServicesEnabledFromConfig,
		FromYAML:        serviceListFromYAML,
	})
}

func firewallServicesEnabledFromConfig(c *buildconfig.BuildConfig) (CheckParams, error) {
	if c == nil || c.Blueprint == nil || c.Blueprint.Customizations == nil {
		return nil, nil
	}
	fw := c.Blueprint.Customizations.Firewall
	if fw == nil || fw.Services == nil || len(fw.Services.Enabled) == 0 {
		return nil, nil
	}
	return ServiceListParams{Services: fw.Services.Enabled}, nil
}

func firewallServicesEnabledCheck(meta *Metadata, params CheckParams) error {
	p, ok := params.(ServiceListParams)
	if !ok {
		return Fail("invalid params type")
	}

	for _, service := range p.Services {
		log.Printf("Checking enabled firewall service: %s\n", service)
		// NOTE: sudo works here without password because we test this only on ami
		// initialised with cloud-init, which sets sudo NOPASSWD for the user
		state, _, _, err := ExecString("sudo", "firewall-cmd", "--query-service="+service)
		if err != nil {
			return Fail("firewall service is not enabled:", service, "error:", err)
		}
		if state != "yes" {
			return Fail("firewall service is not enabled:", service, "state:", state)
		}
		log.Printf("Firewall service was enabled service=%s state=%s\n", service, state)
	}

	return Pass()
}
