package check

import (
	"fmt"
	"os"
	"strconv"
	"syscall"

	"github.com/osbuild/image-builder/internal/buildconfig"
	"gopkg.in/yaml.v3"
)

type FileCheckEntry struct {
	Path    string `yaml:"path"`
	Mode    string `yaml:"mode,omitempty"`
	User    any    `yaml:"user,omitempty"`
	Group   any    `yaml:"group,omitempty"`
	Content string `yaml:"content,omitempty"`
}

type FilesCheckParams struct {
	Entries []FileCheckEntry `yaml:"entries"`
}

func filesFromYAML(node *yaml.Node) (CheckParams, error) {
	var p FilesCheckParams
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
			Name: "files",
		},
		Func:       filesCheck,
		FromBuildConfig: filesFromConfig,
		FromYAML:        filesFromYAML,
	})
}

func filesFromConfig(c *buildconfig.BuildConfig) (CheckParams, error) {
	if c == nil || c.Blueprint == nil || c.Blueprint.Customizations == nil {
		return nil, nil
	}
	files := c.Blueprint.Customizations.Files
	if len(files) == 0 {
		return nil, nil
	}
	entries := make([]FileCheckEntry, len(files))
	for i, f := range files {
		entries[i] = FileCheckEntry{
			Path:    f.Path,
			Mode:    f.Mode,
			User:    f.User,
			Group:   f.Group,
			Content: f.Data,
		}
	}
	return FilesCheckParams{Entries: entries}, nil
}

func filesCheck(meta *Metadata, params CheckParams) error {
	p, ok := params.(FilesCheckParams)
	if !ok {
		return Fail("invalid params type")
	}

	for _, file := range p.Entries {
		if !Exists(file.Path) {
			return Fail("file does not exist:", file.Path)
		}

		info, err := Stat(file.Path)
		if err != nil {
			return Fail("failed to get file info:", file.Path)
		}
		stat, ok := info.Sys().(*syscall.Stat_t)
		if !ok {
			return Fail("check only works on UNIX-like")
		}
		mode, uid, gid := info.Mode(), stat.Uid, stat.Gid

		if file.Mode != "" {
			userMode, err := strconv.ParseUint(file.Mode, 8, 32)
			if err != nil {
				return Fail("failed to parse file mode:", file.Path)
			}

			if mode.Perm() != os.FileMode(userMode) {
				return Fail("file mode does not match:", file.Path)
			}
		}

		if file.User != nil {
			expectedUid, err := resolveUser(file.User)
			if err != nil {
				return Fail("failed to resolve user:", file.Path, err)
			}
			if uid != expectedUid {
				return Fail("file user does not match:", file.Path)
			}
		}

		if file.Group != nil {
			expectedGid, err := resolveGroup(file.Group)
			if err != nil {
				return Fail("failed to resolve group:", file.Path, err)
			}
			if gid != expectedGid {
				return Fail("file group does not match:", file.Path)
			}
		}

		if len(file.Content) > 0 {
			content, err := ReadFile(file.Path)
			if err != nil {
				return Fail("failed to read file:", file.Path)
			}

			if string(content) != file.Content {
				return Fail("file content does not match:", file.Path)
			}
		}
	}

	return Pass()
}
