package check

import (
	"fmt"
	"log"
	"strings"

	"github.com/osbuild/image-builder/internal/buildconfig"
	"gopkg.in/yaml.v3"
)

type ModularityCheckParams struct {
	Modules []string `yaml:"modules"`
}

func modularityFromYAML(node *yaml.Node) (CheckParams, error) {
	var p ModularityCheckParams
	if err := node.Decode(&p); err != nil {
		return nil, err
	}
	if len(p.Modules) == 0 {
		return nil, fmt.Errorf("'modules' list must not be empty")
	}
	return p, nil
}

func init() {
	RegisterCheckWithParams(RegisteredCheck{
		Meta: &Metadata{
			Name:  "modularity",
			RunOn: []string{"centos-9"},
		},
		ParamFunc:       modularityCheck,
		FromBuildConfig: modularityFromConfig,
		FromYAML:        modularityFromYAML,
	})
}

func modularityFromConfig(c *buildconfig.BuildConfig) (CheckParams, error) {
	if c == nil || c.Blueprint == nil {
		return nil, nil
	}
	var modules []string
	for _, mod := range c.Blueprint.EnabledModules {
		modules = append(modules, mod.Name+":"+mod.Stream)
	}
	for _, pkg := range c.Blueprint.Packages {
		if strings.HasPrefix(pkg.Name, "@") && strings.Contains(pkg.Name, ":") {
			moduleName := strings.TrimPrefix(pkg.Name, "@")
			modules = append(modules, moduleName)
		}
	}
	if len(modules) == 0 {
		return nil, nil
	}
	return ModularityCheckParams{Modules: modules}, nil
}

func modularityCheck(meta *Metadata, params CheckParams) error {
	p, ok := params.(ModularityCheckParams)
	if !ok {
		return Fail("invalid params type")
	}

	// Verify modules that are enabled on a system, if any. Modules can either be enabled separately
	// or they can be installed through packages directly. We test both cases here.
	//
	// Caveat is that when a module is enabled yet _no_ packages are installed from it this breaks.
	// Let's not do that in the test?

	log.Println("Checking enabled modules")

	// Get list of enabled modules from dnf (use -y for non-interactive, -q to suppress download progress output)
	stdout, _, _, err := Exec("dnf", "-y", "-q", "module", "list", "--enabled")
	if err != nil {
		return Fail("failed to list enabled modules:", err)
	}

	// Parse dnf output: detect table rows dynamically (lines with at least 3 columns)
	lines := strings.Split(string(stdout), "\n")
	enabledModules := make(map[string]bool)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		moduleKey := fields[0] + ":" + fields[1]
		enabledModules[moduleKey] = true
	}
	if len(enabledModules) == 0 {
		return Fail("dnf module list returned nothing")
	}

	for _, expected := range p.Modules {
		if !enabledModules[expected] {
			return Fail("module was not enabled:", expected)
		}
		log.Printf("Expected module %q was enabled\n", expected)
	}

	return Pass()
}
