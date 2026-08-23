package check

import (
	"github.com/osbuild/image-builder/internal/buildconfig"
	"gopkg.in/yaml.v3"
)

// Metadata provides information about a check. It is used to manage the execution
// of the check and to provide context in logs and reports.
type Metadata struct {
	Name         string   // Name of the check (used for lookup and logging)
	TempDisabled string   // Set to non-empty string with URL to issue tracker to disable the check temporarily
	RunOn        []string // List of OS IDs to run the check on (prefix with `!` to exclude)
}

// CheckParams is the intermediate type that decouples check logic from BuildConfig.
type CheckParams any

// CheckFunc is a check function that receives extracted params instead of BuildConfig.
type CheckFunc func(meta *Metadata, params CheckParams) error

// ParamsFromConfig extracts CheckParams from a BuildConfig.
type ParamsFromConfig func(*buildconfig.BuildConfig) (CheckParams, error)

// ParamsFromYAML parses CheckParams from a YAML node.
type ParamsFromYAML func(*yaml.Node) (CheckParams, error)

// RegisteredCheck represents a registered check with its metadata and function.
type RegisteredCheck struct {
	Meta            *Metadata
	Func            CheckFunc
	FromBuildConfig ParamsFromConfig
	FromYAML        ParamsFromYAML
}

// Result represents the outcome of a check execution.
type Result struct {
	Meta  *Metadata // Metadata of the check
	Error error     // Error, warning, skip or nil if the check passed
}

// Make the Result type sortable by error type: passed, skipped, warning, failed
type SortedResults []Result

func (sr SortedResults) Len() int {
	return len(sr)
}

func (sr SortedResults) Swap(i, j int) {
	sr[i], sr[j] = sr[j], sr[i]
}

func (sr SortedResults) Less(i, j int) bool {
	getRank := func(err error) int {
		switch {
		case err == nil:
			return 0 // passed
		case IsSkip(err):
			return 1 // skipped
		case IsWarning(err):
			return 2 // warning
		case IsFail(err):
			return 3 // failed
		default:
			return 4 // unknown errors last
		}
	}
	return getRank(sr[i].Error) < getRank(sr[j].Error)
}

// RegisterCheck registers a check implementation.
func RegisterCheck(rc RegisteredCheck) {
	checkRegistry = append(checkRegistry, rc)
}

var checkRegistry []RegisteredCheck

// GetAllChecks returns all registered checks.
func GetAllChecks() []RegisteredCheck {
	return checkRegistry
}

// FindCheckByName finds a registered check by its name.
func FindCheckByName(name string) (RegisteredCheck, bool) {
	for _, chk := range checkRegistry {
		if chk.Meta.Name == name {
			return chk, true
		}
	}
	return RegisteredCheck{}, false
}

func MustFindCheckByName(name string) RegisteredCheck {
	chk, found := FindCheckByName(name)
	if !found {
		panic("check not found: " + name)
	}
	return chk
}
