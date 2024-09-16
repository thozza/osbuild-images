// Simple tool to dump a JSON object containing all package sets for all
// supported (tested) distros x arches x image types.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/osbuild/images/internal/cmdutil"
	"github.com/osbuild/images/pkg/blueprint"
	"github.com/osbuild/images/pkg/distro"
	"github.com/osbuild/images/pkg/distrofactory"
	"github.com/osbuild/images/pkg/ostree"
	"github.com/osbuild/images/pkg/reporegistry"
)

func main() {
	// selection args
	var arches, distros, imgTypes cmdutil.MultiValue
	flag.Var(&arches, "arches", "comma-separated list of architectures (globs supported)")
	flag.Var(&distros, "distros", "comma-separated list of distributions (globs supported)")
	flag.Var(&imgTypes, "types", "comma-separated list of image types (globs supported)")
	flag.Parse()

	testedRepoRegistry, err := reporegistry.NewTestedDefault()
	if err != nil {
		panic(fmt.Sprintf("failed to create repo registry with tested distros: %v", err))
	}

	df := distrofactory.NewDefault()

	distros, invalidDistros := distros.ResolveArgValues(testedRepoRegistry.ListDistros())
	if len(invalidDistros) > 0 {
		fmt.Fprintf(os.Stderr, "WARNING: invalid distro names: [%s]\n", strings.Join(invalidDistros, ","))
	}
	for _, distroName := range distros {
		distribution := df.GetDistro(distroName)
		if distribution == nil {
			fmt.Fprintf(os.Stderr, "WARNING: invalid distro name %q\n", distroName)
			continue
		}

		distroArches, invalidArches := arches.ResolveArgValues(distribution.ListArches())
		if len(invalidArches) > 0 {
			fmt.Fprintf(os.Stderr, "WARNING: invalid arch names [%s] for distro %q\n", strings.Join(invalidArches, ","), distroName)
		}
		for _, archName := range distroArches {
			arch, err := distribution.GetArch(archName)
			if err != nil {
				// resolveArgValues should prevent this
				panic(fmt.Sprintf("invalid arch name %q for distro %q: %s\n", archName, distroName, err.Error()))
			}

			daImgTypes, invalidImageTypes := imgTypes.ResolveArgValues(arch.ListImageTypes())
			if len(invalidImageTypes) > 0 {
				fmt.Fprintf(os.Stderr, "WARNING: invalid image type names [%s] for distro %q and arch %q\n", strings.Join(invalidImageTypes, ","), distroName, archName)
			}
			for _, imgTypeName := range daImgTypes {
				imgType, err := arch.GetImageType(imgTypeName)
				if err != nil {
					// resolveArgValues should prevent this
					panic(fmt.Sprintf("invalid image type %q for distro %q and arch %q: %s\n", imgTypeName, distroName, archName, err.Error()))
				}

				// set up bare minimum args for image type
				var customizations *blueprint.Customizations
				if imgType.Name() == "edge-simplified-installer" || imgType.Name() == "iot-simplified-installer" {
					customizations = &blueprint.Customizations{
						InstallationDevice: "/dev/null",
					}
				}
				bp := blueprint.Blueprint{
					Customizations: customizations,
				}
				options := distro.ImageOptions{
					OSTree: &ostree.ImageOptions{
						URL: "https://example.com", // required by some image types
					},
				}

				manifest, _, err := imgType.Manifest(&bp, options, nil, 0)
				if err != nil {
					panic(err)
				}

				encoder := json.NewEncoder(os.Stdout)
				encoder.SetIndent("", "  ")
				_ = encoder.Encode(manifest.GetPackageSetChains())
			}
		}
	}
}
