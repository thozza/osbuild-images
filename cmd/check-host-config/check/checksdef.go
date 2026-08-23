package check

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type ChecksDefinition struct {
	Version int          `yaml:"version"`
	Checks  []CheckEntry `yaml:"checks"`
}

type CheckEntry struct {
	Type string    `yaml:"type"`
	Node yaml.Node `yaml:"-"`
}

func (e *CheckEntry) UnmarshalYAML(node *yaml.Node) error {
	var raw struct {
		Type string `yaml:"type"`
	}
	if err := node.Decode(&raw); err != nil {
		return err
	}
	if raw.Type == "" {
		return fmt.Errorf("check entry missing required 'type' field")
	}
	e.Type = raw.Type
	e.Node = *node
	return nil
}

func ParseChecksFile(data []byte) (*ChecksDefinition, error) {
	var def ChecksDefinition
	if err := yaml.Unmarshal(data, &def); err != nil {
		return nil, fmt.Errorf("parsing checks file: %w", err)
	}
	if def.Version != 1 {
		return nil, fmt.Errorf("unsupported checks definition version: %d", def.Version)
	}
	if len(def.Checks) == 0 {
		return nil, fmt.Errorf("no checks defined")
	}
	return &def, nil
}
