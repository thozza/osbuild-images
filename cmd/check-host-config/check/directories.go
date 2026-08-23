package check

import (
	"fmt"
	"strconv"
	"syscall"

	"github.com/osbuild/image-builder/internal/buildconfig"
	"gopkg.in/yaml.v3"
)

type DirectoryCheckEntry struct {
	Path  string `yaml:"path"`
	Mode  string `yaml:"mode,omitempty"`
	User  any    `yaml:"user,omitempty"`
	Group any    `yaml:"group,omitempty"`
}

type DirectoriesCheckParams struct {
	Entries []DirectoryCheckEntry `yaml:"entries"`
}

func directoriesFromYAML(node *yaml.Node) (CheckParams, error) {
	var p DirectoriesCheckParams
	if err := node.Decode(&p); err != nil {
		return nil, err
	}
	if len(p.Entries) == 0 {
		return nil, fmt.Errorf("'entries' list must not be empty")
	}
	return p, nil
}

func init() {
	RegisterCheck(RegisteredCheck{
		Meta: &Metadata{
			Name: "directories",
		},
		Func:       directoriesCheck,
		FromBuildConfig: directoriesFromConfig,
		FromYAML:        directoriesFromYAML,
	})
}

func directoriesFromConfig(c *buildconfig.BuildConfig) (CheckParams, error) {
	if c == nil || c.Blueprint == nil || c.Blueprint.Customizations == nil {
		return nil, nil
	}
	dirs := c.Blueprint.Customizations.Directories
	if len(dirs) == 0 {
		return nil, nil
	}
	entries := make([]DirectoryCheckEntry, len(dirs))
	for i, d := range dirs {
		entries[i] = DirectoryCheckEntry{
			Path:  d.Path,
			Mode:  d.Mode,
			User:  d.User,
			Group: d.Group,
		}
	}
	return DirectoriesCheckParams{Entries: entries}, nil
}

func directoriesCheck(meta *Metadata, params CheckParams) error {
	p, ok := params.(DirectoriesCheckParams)
	if !ok {
		return Fail("invalid params type")
	}

	for _, dir := range p.Entries {
		if !ExistsDir(dir.Path) {
			return Fail("directory does not exist:", dir.Path)
		}

		info, err := Stat(dir.Path)
		if err != nil {
			return Fail("failed to get directory info:", dir.Path)
		}
		stat, ok := info.Sys().(*syscall.Stat_t)
		if !ok {
			return Fail("check only works on UNIX-like")
		}
		mode, uid, gid := info.Mode(), stat.Uid, stat.Gid
		if !info.IsDir() {
			return Fail("path is not a directory:", dir.Path)
		}

		if dir.Mode != "" {
			userMode, err := strconv.ParseUint(dir.Mode, 8, 32)
			if err != nil {
				return Fail("failed to parse directory mode:", dir.Path)
			}

			if int64(mode.Perm()) != int64(userMode) {
				return Fail("directory mode does not match:", dir.Path)
			}
		}

		if dir.User != nil {
			expectedUid, err := resolveUser(dir.User)
			if err != nil {
				return Fail("failed to resolve user:", dir.Path, err)
			}
			if uid != expectedUid {
				return Fail("directory user does not match:", dir.Path)
			}
		}

		if dir.Group != nil {
			expectedGid, err := resolveGroup(dir.Group)
			if err != nil {
				return Fail("failed to resolve group:", dir.Path, err)
			}
			if gid != expectedGid {
				return Fail("directory group does not match:", dir.Path)
			}
		}
	}

	return Pass()
}
