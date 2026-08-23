package check

import (
	"log"
	"strings"

	"github.com/osbuild/image-builder/internal/buildconfig"
	"gopkg.in/yaml.v3"
)

type KernelCheckParams struct {
	Name   string `yaml:"name,omitempty"`
	Append string `yaml:"append,omitempty"`
}

func kernelFromYAML(node *yaml.Node) (CheckParams, error) {
	var p KernelCheckParams
	if err := node.Decode(&p); err != nil {
		return nil, err
	}
	return p, nil
}

func init() {
	RegisterCheck(RegisteredCheck{
		Meta: &Metadata{
			Name:         "kernel",
			TempDisabled: "https://github.com/osbuild/image-builder/pull/2175",
		},
		Func:            kernelCheck,
		FromBuildConfig: kernelFromConfig,
		FromYAML:        kernelFromYAML,
	})
}

func kernelFromConfig(c *buildconfig.BuildConfig) (CheckParams, error) {
	if c == nil || c.Blueprint == nil || c.Blueprint.Customizations == nil {
		return nil, nil
	}
	k := c.Blueprint.Customizations.Kernel
	if k == nil {
		return nil, nil
	}
	return KernelCheckParams{Name: k.Name, Append: k.Append}, nil
}

func kernelCheck(meta *Metadata, params CheckParams) error {
	p, ok := params.(KernelCheckParams)
	if !ok {
		return Fail("invalid params type")
	}

	// Only query RPM for the kernel package provides. We do no test if the
	// specific kernel is actually booted as the testing in container is not
	// reliable.
	if p.Name != "" {
		_, _, _, err := ExecString("rpm", "-q", "--provides", p.Name)
		if err != nil {
			return Fail("kernel package not found:", p.Name, "error:", err)
		}

		log.Printf("Kernel name check passed: %s is installed\n", p.Name)
	}

	if len(p.Append) > 0 {
		cmdline, err := ReadFile("/proc/cmdline")
		if err != nil {
			return Fail("failed to read /proc/cmdline:", err)
		}

		if !strings.Contains(string(cmdline), p.Append) {
			return Fail("kernel options append does not match:", p.Append)
		}
	}

	return Pass()
}
