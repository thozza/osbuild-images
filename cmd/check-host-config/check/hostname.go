package check

import (
	"errors"
	"fmt"
	"strings"

	"github.com/osbuild/image-builder/internal/buildconfig"
	"gopkg.in/yaml.v3"
)

const hostnameFilePath = "/etc/hostname"

type HostnameCheckParams struct {
	Expected string `yaml:"expected"`
}

func hostnameFromYAML(node *yaml.Node) (CheckParams, error) {
	var p HostnameCheckParams
	if err := node.Decode(&p); err != nil {
		return nil, err
	}
	if p.Expected == "" {
		return nil, fmt.Errorf("'expected' must not be empty")
	}
	return p, nil
}

func init() {
	RegisterCheckWithParams(RegisteredCheck{
		Meta: &Metadata{
			Name:                   "hostname",
			RequiresBlueprint:      true,
			RequiresCustomizations: true,
		},
		ParamFunc:       hostnameCheck,
		FromBuildConfig: hostnameFromConfig,
		FromYAML:        hostnameFromYAML,
	})
}

func hostnameFromConfig(c *buildconfig.BuildConfig) (CheckParams, error) {
	if c == nil || c.Blueprint == nil || c.Blueprint.Customizations == nil {
		return nil, nil
	}
	expected := c.Blueprint.Customizations.Hostname
	if expected == nil || *expected == "" {
		return nil, nil
	}
	return HostnameCheckParams{Expected: *expected}, nil
}

var ErrHostname = errors.New("hostname")

func getHostname() (string, error) {
	if hostname, _, _, err := ExecString("hostnamectl", "hostname"); err == nil && hostname != "" {
		return hostname, nil
	}

	if hostname, _, _, err := ExecString("hostname"); err == nil && hostname != "" {
		return hostname, nil
	}

	data, err := ReadFile(hostnameFilePath)
	if err != nil {
		return "", fmt.Errorf("%w: could not read %q", ErrHostname, hostnameFilePath)
	}

	firstLine, _, _ := strings.Cut(string(data), "\n")
	hostname := strings.TrimSpace(firstLine)
	if hostname != "" {
		return hostname, nil
	}
	return "", fmt.Errorf("%w: could not get hostname: tried hostnamectl, hostname, and %s", ErrHostname, hostnameFilePath)
}

func hostnameCheck(meta *Metadata, params CheckParams) error {
	p, ok := params.(HostnameCheckParams)
	if !ok {
		return Fail("invalid params type")
	}

	hostname, err := getHostname()
	if err != nil {
		return err
	}

	// we only emit a warning here since the hostname gets reset by cloud-init and we're not
	// entirely sure how to deal with it yet on the service level
	if hostname != p.Expected {
		return Warning("hostname does not match, got", hostname, "expected", p.Expected)
	}

	return Pass()
}
