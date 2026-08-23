package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/osbuild/image-builder/cmd/check-host-config/check"
	"github.com/osbuild/image-builder/internal/buildconfig"
)

// waitForSystem waits until the system is reported by systemd as "running" or the timeout is reached.
func waitForSystem(timeout time.Duration) error {
	if timeout <= 0 {
		return nil
	}

	if err := runningWait(timeout, 15*time.Second); err != nil {
		fmt.Fprintf(os.Stderr, "Error while waiting for system to be running: %v\n", err)
		if errors.Is(err, ErrTimeout) {
			if activatingUnits := listBadUnits(); len(activatingUnits) > 0 {
				fmt.Fprintf(os.Stderr, "Units still activating: %s\n", strings.Join(activatingUnits, " "))
				for _, unit := range activatingUnits {
					fmt.Fprintf(os.Stderr, "Unit %s journal:\n", unit)
					printUnitJournal(unit)
				}
			}
		}
		return err
	}
	return nil
}

// shouldRunOn returns whether a check should run on the current system.
func shouldRunOn(osRelease *check.OSRelease, runOn []string) bool {
	if len(runOn) == 0 || osRelease == nil {
		return true
	}

	currentID := strings.ToLower(strings.TrimSpace(osRelease.ID + "-" + osRelease.VersionID))
	var inclusions []string
	for _, entry := range runOn {
		entry = strings.TrimSpace(entry)
		if after, ok := strings.CutPrefix(entry, "!"); ok {
			if strings.ToLower(after) == currentID {
				return false
			}
		} else {
			inclusions = append(inclusions, strings.ToLower(entry))
		}
	}

	return len(inclusions) == 0 || slices.Contains(inclusions, currentID)
}

// runChecks runs all checks sequentially and processes their results.
func runChecks(runs []check.CheckRun, mode check.RunMode, osRelease *check.OSRelease, quiet bool) bool {
	defer log.SetPrefix("")
	if quiet {
		log.SetOutput(io.Discard)
		defer log.SetOutput(os.Stdout)
	}

	var results check.SortedResults
	for _, run := range runs {
		meta := run.Check.Meta
		var err error
		log.SetPrefix(meta.Name + ": ")

		if mode == check.ModeBuildConfig {
			switch {
			case !shouldRunOn(osRelease, meta.RunOn):
				err = check.Skip(osRelease.ID + "-" + osRelease.VersionID + " excluded via RunOn: " + strings.Join(meta.RunOn, ", "))
				results = append(results, check.Result{Meta: meta, Error: err})
				if err != nil {
					log.Println(err)
				}
				continue
			case meta.TempDisabled != "":
				err = check.Skip("temporarily disabled: " + meta.TempDisabled)
				results = append(results, check.Result{Meta: meta, Error: err})
				if err != nil {
					log.Println(err)
				}
				continue
			}
		}

		switch {
		case run.Err != nil:
			err = run.Err
		case run.Params == nil:
			err = check.Skip("no relevant configuration")
		default:
			err = run.Check.Func(meta, run.Params)
		}

		results = append(results, check.Result{Meta: meta, Error: err})
		if err != nil {
			log.Println(err)
		}
	}

	log.SetOutput(os.Stdout)
	sort.Sort(results)
	var seenError bool
	for _, res := range results {
		err := res.Error
		icon := check.IconFor(err)

		switch err {
		case nil:
			fmt.Printf("%s %s: passed\n", icon, res.Meta.Name)
		default:
			if !check.IsSkip(err) && !check.IsWarning(err) {
				seenError = true
			}
			fmt.Printf("%s %s: %s\n", icon, res.Meta.Name, err)
		}
	}

	return !seenError
}

func readInput(path string) ([]byte, error) {
	if path == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(path)
}

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(0)

	checksFile := flag.String("checks-file", "", "path to YAML checks definition file (- for stdin)")
	validate := flag.Bool("validate", false, "validate checks file and exit without running checks")
	waitTimeout := flag.Duration("wait-timeout", 15*time.Minute, "timeout for waiting for system to be running (0 to skip)")
	quiet := flag.Bool("quiet", false, "less logging output")
	flag.Parse()

	hasChecksFile := *checksFile != ""
	hasConfigArg := flag.Arg(0) != ""

	if *validate && !hasChecksFile {
		log.Fatalf("--validate requires --checks-file")
	}
	if hasChecksFile == hasConfigArg {
		log.Fatalf("Provide exactly one of: --checks-file <file> or <config.json>")
	}

	if hasChecksFile {
		data, err := readInput(*checksFile)
		if err != nil {
			log.Fatalf("Failed to read checks file: %v", err)
		}
		def, err := check.ParseChecksFile(data)
		if err != nil {
			log.Fatalf("Invalid checks file: %v", err)
		}
		runs, err := check.PrepareFromYAML(def)
		if err != nil {
			log.Fatalf("Failed to prepare checks: %v", err)
		}
		if *validate {
			fmt.Printf("Checks file valid: %d checks defined\n", len(runs))
			return
		}
		if err := waitForSystem(*waitTimeout); err != nil {
			log.Fatalf("Problem during waiting for system to be running: %v", err)
		}
		if !runChecks(runs, check.ModeYAML, nil, *quiet) {
			log.Fatalf("Checks from %q failed, return code 1", *checksFile)
		}
	} else {
		config, err := buildconfig.New(flag.Arg(0), nil)
		if err != nil {
			log.Fatalf("Failed to load build config: %v", err)
		}
		if err := waitForSystem(*waitTimeout); err != nil {
			log.Fatalf("Problem during waiting for system to be running: %v", err)
		}
		osRelease, _ := check.ParseOSRelease("")
		if osRelease == nil {
			log.Println("Could not parse /etc/os-release, RunOn filtering disabled")
		}
		runs := check.PrepareFromBuildConfig(config)
		if !runChecks(runs, check.ModeBuildConfig, osRelease, *quiet) {
			log.Fatalf("Host check with config %q failed, return code 1", flag.Arg(0))
		}
	}
}
