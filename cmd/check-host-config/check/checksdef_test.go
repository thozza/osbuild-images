package check_test

import (
	"testing"

	check "github.com/osbuild/image-builder/cmd/check-host-config/check"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseChecksFile(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantErr    string
		wantChecks int
	}{
		{
			name: "valid file with one check",
			input: `
version: 1
checks:
  - type: srv-enabled
    services:
      - sshd
`,
			wantChecks: 1,
		},
		{
			name: "missing version field",
			input: `
checks:
  - type: srv-enabled
`,
			wantErr: "unsupported checks definition version: 0",
		},
		{
			name: "unsupported version",
			input: `
version: 2
checks:
  - type: srv-enabled
`,
			wantErr: "unsupported checks definition version: 2",
		},
		{
			name: "empty checks list",
			input: `
version: 1
checks: []
`,
			wantErr: "no checks defined",
		},
		{
			name: "entry missing type field",
			input: `
version: 1
checks:
  - services:
      - sshd
`,
			wantErr: "check entry missing required 'type' field",
		},
		{
			name: "multiple entries parsed in order with node preserved",
			input: `
version: 1
checks:
  - type: srv-enabled
    services:
      - sshd
  - type: hostname
    expected: myhost
  - type: bootc-status
`,
			wantChecks: 3,
		},
		{
			name:    "invalid YAML syntax",
			input:   `{{{not yaml`,
			wantErr: "parsing checks file:",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			def, err := check.ParseChecksFile([]byte(tt.input))
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, def)
			assert.Len(t, def.Checks, tt.wantChecks)
		})
	}
}

func TestParseChecksFile_EntryOrder(t *testing.T) {
	input := `
version: 1
checks:
  - type: srv-enabled
  - type: hostname
  - type: bootc-status
`
	def, err := check.ParseChecksFile([]byte(input))
	require.NoError(t, err)
	require.Len(t, def.Checks, 3)
	assert.Equal(t, "srv-enabled", def.Checks[0].Type)
	assert.Equal(t, "hostname", def.Checks[1].Type)
	assert.Equal(t, "bootc-status", def.Checks[2].Type)
}

func TestParseChecksFile_NodePreserved(t *testing.T) {
	input := `
version: 1
checks:
  - type: srv-enabled
    services:
      - sshd
      - chronyd
`
	def, err := check.ParseChecksFile([]byte(input))
	require.NoError(t, err)
	require.Len(t, def.Checks, 1)

	// The Node should contain the full mapping including 'services'
	var decoded struct {
		Type     string   `yaml:"type"`
		Services []string `yaml:"services"`
	}
	err = def.Checks[0].Node.Decode(&decoded)
	require.NoError(t, err)
	assert.Equal(t, "srv-enabled", decoded.Type)
	assert.Equal(t, []string{"sshd", "chronyd"}, decoded.Services)
}
