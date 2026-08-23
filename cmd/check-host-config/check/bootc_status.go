package check

import (
	"github.com/osbuild/image-builder/internal/buildconfig"
	"gopkg.in/yaml.v3"
)

type BootcStatusParams struct{}

func bootcStatusFromYAML(node *yaml.Node) (CheckParams, error) {
	return BootcStatusParams{}, nil
}

func init() {
	RegisterCheckWithParams(RegisteredCheck{
		Meta: &Metadata{
			Name:          "bootc-status",
			RequiresBootc: true,
		},
		ParamFunc:       bootcStatusCheck,
		FromBuildConfig: bootcStatusFromConfig,
		FromYAML:        bootcStatusFromYAML,
	})
}

func bootcStatusFromConfig(c *buildconfig.BuildConfig) (CheckParams, error) {
	if c == nil || c.Options.Bootc == nil {
		return nil, nil
	}
	return BootcStatusParams{}, nil
}

func bootcStatusCheck(meta *Metadata, params CheckParams) error {
	stdout, stderr, _, err := ExecString("sudo", "bootc", "status")
	if err != nil {
		return Fail("bootc status failed:", err, "\nstdout:", stdout, "\nstderr:", stderr)
	}
	return Pass()
}
