## check-host-config

A command used to test a host environment against a build configuration or a
YAML checks definition file. Before validation starts, it waits until systemd
reports the system as fully booted.

The command is safe to run on development machines; it does not change
configuration or leave resources behind.

### Usage

**BuildConfig mode** (existing):

    check-host-config <config.json>

**YAML checks mode**:

    check-host-config --checks-file checks.yaml
    cat checks.yaml | check-host-config --checks-file -

**Validation only** (parse and validate without running checks):

    check-host-config --checks-file checks.yaml --validate

In YAML mode, `TempDisabled` and `RunOn` metadata are ignored - every listed
check runs unconditionally. The `version` field guards the top-level structure
only, not per-check parameter schemas.

See `check/testdata/example-checks.yaml` for a complete working example
covering all 16 check types. It can be validated with `--validate`.

### YAML checks file format

```yaml
version: 1
checks:
  - type: <check-name>
    <check-specific parameters>
```

Supported check types and their parameters:

| Check type | Parameters |
|---|---|
| `srv-enabled`, `srv-disabled`, `srv-masked`, `fw-srv-enabled`, `fw-srv-disabled` | `services: [list]` |
| `fw-ports` | `ports: ["8080:tcp"]` |
| `files` | `entries: [{path, mode, user, group, content}]` |
| `directories` | `entries: [{path, mode, user, group}]` |
| `filesystem` | `mountpoints: ["/var", "/home"]` |
| `hostname` | `expected: "myhost.example.com"` |
| `bootc-status` | (no parameters) |
| `users` | `users: ["root", "appuser"]` |
| `kernel` | `name: "kernel-rt"`, `append: "quiet"` |
| `modularity` | `modules: ["nodejs:18"]` |
| `cacerts` | `certificates: ["-----BEGIN..."]` |
| `oscap` | `profile_id: "xccdf_..."`, `datastream: "/path/to/ds.xml"` (optional) |

### Implementing new checks

Each check is registered with a `RegisterCheck` call in an `init()` function.
The check receives extracted parameters (not the raw BuildConfig) via `Func`.
Two extractors must be provided: `FromBuildConfig` for BuildConfig mode and
`FromYAML` for YAML mode.

```go
func init() {
    RegisterCheck(RegisteredCheck{
        Meta: &Metadata{
            Name:         "users",
            TempDisabled: "",
            RunOn:        []string{"centos", "!rhel"},
        },
        Func:            usersCheck,
        FromBuildConfig: usersFromConfig,
        FromYAML:        usersFromYAML,
    })
}
```

Metadata fields:

* **Name**: name of the check.
* **TempDisabled**: the check is temporarily disabled (skipped) when this is
  not an empty string (e.g. issue URL). Only evaluated in BuildConfig mode.
* **RunOn**: when set, run on specific distro IDs (or use bang to exclude).
  Only evaluated in BuildConfig mode.

`FromBuildConfig` returns `nil` params to skip the check (replaces the old
`RequiresBlueprint`/`RequiresCustomizations`/`RequiresBootc` flags).

Checks can return:

* pass - `Pass()` function (returns `nil`)
* warning - `Warning(reason)` function
* error - `Fail(reason)` function
* skip - `Skip(reason)` function

### Unit testing

For complex checks (e.g. OpenSCAP, DNF, RPM), unit tests help identify problems
early. Functions such as `Exec`, `ExecString`, `Exists`, `Grep`, and `ReadFile`
are available in the package and can be mocked using helpers in unit tests. Most
tests except OpenSCAP checks are clearer in tabular form; prefer that style. For
readability, the package provides helper types and functions in
`mock_helpers_test.go` that allow reusing the following testing pattern:

	tests := []struct {
		name         string
		config       *blueprint.KernelCustomization
		mockExec     map[string]ExecResult
		mockReadFile map[string]ReadFileResult
		wantErr      error
	}

Each test case has a name and configuration, and returns either nil or an
error. It can also define one or more mocks, represented as maps from input
(typically a string, or a struct for functions with multiple arguments) to
output (a result struct). The mock maps can be formatted like this:

    {
        name: "fail when append does not match",
        config: &blueprint.KernelCustomization{
            Append: "debug",
        },
        mockReadFile: map[string]ReadFileResult{
            "/proc/cmdline": {Data: []byte("root=UUID=1234-5678 ro")},
        },
        wantErr: check.ErrCheckFailed,
    },

The helper functions install these mocks:

    installMockExec(t, tt.mockExec)
    installMockReadFile(t, tt.mockReadFile)

### Smoke (end-to-end) tests

In addition to unit tests, `main_test.go` contains a set of "smoke" tests that
run in a Fedora container, which is set up so the tests can pass. To build the
container locally and run the tests, run `make host-check-test`.
