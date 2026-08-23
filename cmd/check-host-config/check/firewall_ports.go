package check

import (
	"fmt"
	"log"
	"strings"

	"github.com/osbuild/image-builder/internal/buildconfig"
	"gopkg.in/yaml.v3"
)

type PortListParams struct {
	Ports []string `yaml:"ports"`
}

func portListFromYAML(node *yaml.Node) (CheckParams, error) {
	var p PortListParams
	if err := node.Decode(&p); err != nil {
		return nil, err
	}
	if len(p.Ports) == 0 {
		return nil, fmt.Errorf("'ports' list must not be empty")
	}
	return p, nil
}

func init() {
	RegisterCheckWithParams(RegisteredCheck{
		Meta: &Metadata{
			Name: "fw-ports",
		},
		ParamFunc:       firewallPortsCheck,
		FromBuildConfig: firewallPortsFromConfig,
		FromYAML:        portListFromYAML,
	})
}

func firewallPortsFromConfig(c *buildconfig.BuildConfig) (CheckParams, error) {
	if c == nil || c.Blueprint == nil || c.Blueprint.Customizations == nil {
		return nil, nil
	}
	fw := c.Blueprint.Customizations.Firewall
	if fw == nil || len(fw.Ports) == 0 {
		return nil, nil
	}
	return PortListParams{Ports: fw.Ports}, nil
}

func firewallPortsCheck(meta *Metadata, params CheckParams) error {
	p, ok := params.(PortListParams)
	if !ok {
		return Fail("invalid params type")
	}

	for _, port := range p.Ports {
		// firewall-cmd --query-port uses / as the port/protocol separator, but
		// in the blueprint we use :.
		portQuery := strings.ReplaceAll(port, ":", "/")
		log.Printf("Checking enabled firewall port: %s\n", portQuery)
		// NOTE: sudo works here without password because we test this only on ami
		// initialised with cloud-init, which sets sudo NOPASSWD for the user
		state, _, _, err := ExecString("sudo", "firewall-cmd", "--query-port="+portQuery)
		if err != nil {
			return Fail("firewall port is not enabled:", port, "error:", err)
		}
		if state != "yes" {
			return Fail("firewall port is not enabled:", port, "state:", state)
		}
		log.Printf("Firewall port was enabled port=%s state=%s\n", portQuery, state)
	}

	return Pass()
}
