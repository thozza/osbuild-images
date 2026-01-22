package manifestmock

import (
	"crypto/sha256"
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/osbuild/images/pkg/container"
	"github.com/osbuild/images/pkg/depsolvednf"
	"github.com/osbuild/images/pkg/ostree"
	"github.com/osbuild/images/pkg/rpmmd"
)

func ResolveContainers(containerSources map[string][]container.SourceSpec) map[string][]container.Spec {
	containerSpecs := make(map[string][]container.Spec, len(containerSources))
	for plName, sourceSpecs := range containerSources {
		specs := make([]container.Spec, len(sourceSpecs))
		for idx, src := range sourceSpecs {
			digest := fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(src.Name+src.Source+"digest")))
			id := fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(src.Name+src.Source+"imageid")))
			listDigest := fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(src.Name+src.Source+"list-digest")))
			name := src.Name
			if name == "" {
				name = src.Source
			}
			spec := container.Spec{
				Source:     src.Source,
				Digest:     digest,
				TLSVerify:  src.TLSVerify,
				ImageID:    id,
				LocalName:  name,
				ListDigest: listDigest,
			}
			specs[idx] = spec
		}
		containerSpecs[plName] = specs
	}
	return containerSpecs
}

func ResolveCommits(commitSources map[string][]ostree.SourceSpec) map[string][]ostree.CommitSpec {
	commits := make(map[string][]ostree.CommitSpec, len(commitSources))
	for name, commitSources := range commitSources {
		commitSpecs := make([]ostree.CommitSpec, len(commitSources))
		for idx, commitSource := range commitSources {
			commitSpecs[idx] = mockOSTreeResolve(commitSource)
		}
		commits[name] = commitSpecs
	}
	return commits
}

func Depsolve(packageSets map[string][]rpmmd.PackageSet, archName string) map[string]depsolvednf.DepsolveResult {
	depsolvedSets := make(map[string]depsolvednf.DepsolveResult)

	// Iterate over each pipeline's package set chain
	for name, pkgSetChain := range packageSets {
		// transactions has one entry per PackageSet in the chain.
		// Each transaction contains packages that should be installed together
		// in a single RPM stage. Transactions must be installed in order.
		transactions := make(depsolvednf.TransactionList, 0, len(pkgSetChain))

		// Collect all unique repos used across all transactions (by ID)
		allReposByID := make(map[string]rpmmd.RepoConfig)

		// Track seen checksums to avoid duplicates across transactions
		seenChksumsInc := make(map[string]bool)
		seenChksumsExc := make(map[string]bool)

		// Each PackageSet in the chain represents a single transaction.
		for txIdx, pkgSet := range pkgSetChain {
			// transactionPackages holds packages for the current transaction.
			transactionPackages := make(rpmmd.PackageList, 0)

			// Get repos for this transaction from pkgSet.Repositories
			txRepos := pkgSet.Repositories
			// If empty, use a placeholder for rootDir-based depsolving scenarios
			// (where repos come from inside a container)
			// TODO: figure out a nicer way to handle this.
			if len(txRepos) == 0 {
				txRepos = []rpmmd.RepoConfig{{
					Id:       "rootdir-repos",
					Name:     "rootdir-repos",
					BaseURLs: []string{"file:///rootdir-based-depsolve"},
				}}
			}

			// Collect unique repos for DepsolveResult.Repos
			for _, repo := range txRepos {
				allReposByID[repo.Id] = repo
			}

			include := pkgSet.Include
			slices.Sort(include)
			for idx, pkgName := range include {
				checksum := fmt.Sprintf("%x", sha256.Sum256([]byte(pkgName)))
				// generate predictable but non-empty
				// release/version numbers
				ver := strconv.Itoa(int(pkgName[0]) % 9)
				rel := strconv.Itoa(int(pkgName[1]) % 9)
				// Round-robin repo assignment
				pkgRepo := &txRepos[idx%len(txRepos)]
				spec := rpmmd.Package{
					Name:            pkgName,
					Epoch:           0,
					Version:         ver,
					Release:         rel + ".fk1",
					Arch:            archName,
					RemoteLocations: []string{fmt.Sprintf("https://example.com/repo/packages/%s", pkgName)},
					Checksum:        rpmmd.Checksum{Type: "sha256", Value: checksum},
					Repo:            pkgRepo,
				}
				if seenChksumsInc[spec.Checksum.String()] {
					continue
				}
				seenChksumsInc[spec.Checksum.String()] = true

				transactionPackages = append(transactionPackages, spec)
			}

			exclude := pkgSet.Exclude
			slices.Sort(exclude)
			for i, excludeName := range exclude {
				pkgName := fmt.Sprintf("exclude:%s", excludeName)
				checksum := fmt.Sprintf("%x", sha256.Sum256([]byte(pkgName)))
				// Round-robin repo assignment
				pkgRepo := &txRepos[i%len(txRepos)]
				spec := rpmmd.Package{
					Name:            pkgName,
					Epoch:           0,
					Version:         "0",
					Release:         "0",
					Arch:            "noarch",
					RemoteLocations: []string{fmt.Sprintf("https://example.com/repo/packages/%s", pkgName)},
					Checksum:        rpmmd.Checksum{Type: "sha256", Value: checksum},
					Repo:            pkgRepo,
				}
				if seenChksumsExc[spec.Checksum.String()] {
					continue
				}
				seenChksumsExc[spec.Checksum.String()] = true

				transactionPackages = append(transactionPackages, spec)
			}

			// generate pseudo package for the config of this transaction
			var setRepoNames []string
			for _, setRepo := range pkgSet.Repositories {
				setRepoNames = append(setRepoNames, setRepo.Name)
			}
			configPackageName := fmt.Sprintf("%s:transaction-%d-repos:%s", name, txIdx, strings.Join(setRepoNames, "+"))
			if pkgSet.InstallWeakDeps {
				configPackageName += "-weak"
			}
			depsolveConfigPackage := rpmmd.Package{
				Name:            configPackageName,
				Epoch:           0,
				Version:         "",
				Release:         "",
				Arch:            "noarch",
				RemoteLocations: []string{fmt.Sprintf("https://example.com/repo/packages/%s", configPackageName)},
				Checksum:        rpmmd.Checksum{Type: "sha256", Value: fmt.Sprintf("%x", sha256.Sum256([]byte(configPackageName)))},
				Secrets:         "",
				CheckGPG:        false,
				IgnoreSSL:       false,
				Location:        "",
				RepoID:          "",
				Repo:            &txRepos[0],
			}
			transactionPackages = append(transactionPackages, depsolveConfigPackage)

			// Generate pseudo packages for all repos used in this transaction
			for idx := range txRepos {
				repo := &txRepos[idx]
				// the test repos have the form:
				//   https://rpmrepo..../el9/cs9-x86_64-rt-20240915
				// drop the date as it's not needed for this level of mocks
				baseURL := repo.BaseURLs[0]
				if idx := strings.LastIndex(baseURL, "-"); idx > 0 {
					baseURL = baseURL[:idx]
				}
				url, err := url.Parse(baseURL)
				if err != nil {
					panic(err)
				}
				url.Host = "example.com"
				url.Path = fmt.Sprintf("passed-arch:%s/passed-repo:%s", archName, url.Path)
				repoPkg := rpmmd.Package{
					Name:            url.String(),
					RemoteLocations: []string{url.String()},
					Checksum:        rpmmd.Checksum{Type: "sha256", Value: fmt.Sprintf("%x", sha256.Sum256([]byte(url.String())))},
					Repo:            repo,
				}
				transactionPackages = append(transactionPackages, repoPkg)
			}

			// Add this transaction to the list
			transactions = append(transactions, transactionPackages)
		}

		// Build sorted repos list for deterministic output
		allRepos := make([]rpmmd.RepoConfig, 0, len(allReposByID))
		for _, repo := range allReposByID {
			allRepos = append(allRepos, repo)
		}
		slices.SortFunc(allRepos, func(a, b rpmmd.RepoConfig) int {
			return strings.Compare(a.Id, b.Id)
		})

		depsolvedSets[name] = depsolvednf.DepsolveResult{
			Packages:     transactions.AllPackages(),
			Transactions: transactions,
			Repos:        allRepos,
		}
	}

	return depsolvedSets
}

var OSTreeResolve = mockOSTreeResolve

func mockOSTreeResolve(commitSource ostree.SourceSpec) ostree.CommitSpec {
	checksum := fmt.Sprintf("%x", sha256.Sum256([]byte(commitSource.URL+commitSource.Ref)))
	spec := ostree.CommitSpec{
		Ref:      commitSource.Ref,
		URL:      commitSource.URL,
		Checksum: checksum,
	}
	if commitSource.RHSM {
		spec.Secrets = "org.osbuild.rhsm.consumer"
	}
	return spec
}
