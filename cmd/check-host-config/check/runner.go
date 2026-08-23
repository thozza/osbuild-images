package check

import (
	"fmt"

	"github.com/osbuild/image-builder/internal/buildconfig"
)

type RunMode int

const (
	ModeBuildConfig RunMode = iota
	ModeYAML
)

type CheckRun struct {
	Check  RegisteredCheck
	Params CheckParams
	Err    error
}

func PrepareFromBuildConfig(config *buildconfig.BuildConfig) []CheckRun {
	var runs []CheckRun
	for _, c := range GetAllChecks() {
		params, err := c.FromBuildConfig(config)
		if err != nil {
			runs = append(runs, CheckRun{Check: c, Err: err})
			continue
		}
		if params == nil {
			continue
		}
		runs = append(runs, CheckRun{Check: c, Params: params})
	}
	return runs
}

func PrepareFromYAML(def *ChecksDefinition) ([]CheckRun, error) {
	var runs []CheckRun
	for i, entry := range def.Checks {
		c, ok := FindCheckByName(entry.Type)
		if !ok {
			return nil, fmt.Errorf("checks[%d]: unknown check type %q", i, entry.Type)
		}
		if c.FromYAML == nil {
			return nil, fmt.Errorf("checks[%d]: check %q does not support YAML mode", i, entry.Type)
		}
		params, err := c.FromYAML(&entry.Node)
		if err != nil {
			return nil, fmt.Errorf("checks[%d] (%s): %w", i, entry.Type, err)
		}
		runs = append(runs, CheckRun{Check: c, Params: params})
	}
	return runs, nil
}
