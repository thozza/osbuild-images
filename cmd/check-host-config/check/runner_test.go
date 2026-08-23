package check_test

import (
	_ "embed"
	"testing"

	check "github.com/osbuild/image-builder/cmd/check-host-config/check"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed testdata/example-checks.yaml
var exampleChecksYAML []byte

func TestExampleChecksFileCoversAllChecks(t *testing.T) {
	def, err := check.ParseChecksFile(exampleChecksYAML)
	require.NoError(t, err)

	yamlTypes := make(map[string]bool)
	for _, entry := range def.Checks {
		yamlTypes[entry.Type] = true
	}

	for _, c := range check.GetAllChecks() {
		assert.True(t, yamlTypes[c.Meta.Name],
			"check %q is registered but missing from testdata/example-checks.yaml", c.Meta.Name)
	}
}

func TestExampleChecksFileParsesAndPrepares(t *testing.T) {
	def, err := check.ParseChecksFile(exampleChecksYAML)
	require.NoError(t, err)

	runs, err := check.PrepareFromYAML(def)
	require.NoError(t, err)
	assert.Len(t, runs, len(def.Checks))
}

func TestPrepareFromYAML_UnknownCheckType(t *testing.T) {
	input := []byte(`
version: 1
checks:
  - type: nonexistent-check
`)
	def, err := check.ParseChecksFile(input)
	require.NoError(t, err)

	_, err = check.PrepareFromYAML(def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown check type")
}

func TestPrepareFromYAML_InvalidParams(t *testing.T) {
	input := []byte(`
version: 1
checks:
  - type: srv-enabled
    services: []
`)
	def, err := check.ParseChecksFile(input)
	require.NoError(t, err)

	_, err = check.PrepareFromYAML(def)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must not be empty")
}

func TestPrepareFromYAML_PreservesOrder(t *testing.T) {
	input := []byte(`
version: 1
checks:
  - type: hostname
    expected: myhost
  - type: bootc-status
  - type: users
    users:
      - root
`)
	def, err := check.ParseChecksFile(input)
	require.NoError(t, err)

	runs, err := check.PrepareFromYAML(def)
	require.NoError(t, err)
	require.Len(t, runs, 3)
	assert.Equal(t, "hostname", runs[0].Check.Meta.Name)
	assert.Equal(t, "bootc-status", runs[1].Check.Meta.Name)
	assert.Equal(t, "users", runs[2].Check.Meta.Name)
}
