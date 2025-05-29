Images
======

This repository is, primarily, a Go library for generating osbuild manifests
([more details here](./docs/developer/code-manifest-generation.md)).
It also has some libraries for uploading artifacts to cloud platforms and Koji.

## Getting Started

This section provides a brief overview of how to use the `images` library and `cmd/` tools.

### Using the `images` Library

The `images` library allows you to programmatically define and generate OSBuild manifests in Go. These manifests describe how to build an operating system image.

Here's a simplified example of how you might use the library to create a manifest for a basic Fedora image:

```go
package main

import (
	"fmt"
	"github.com/osbuild/images/pkg/distro"
	"github.com/osbuild/images/pkg/distro/fedora"
	"github.com/osbuild/images/pkg/manifest"
	"github.com/osbuild/images/pkg/ostree"
	"github.com/osbuild/images/pkg/image"
)

func main() {
	// Get the Fedora 39 distribution
	distroRegistry := distro.NewRegistry()
	fedora39, err := distroRegistry.GetDistro("fedora-39")
	if err != nil {
		panic(err)
	}

	// Define the image type (e.g., qcow2)
	imgType, err := fedora39.GetImageType("qcow2")
	if err != nil {
		panic(err)
	}

	// Create a new manifest generator
	// Note: In a real scenario, you would typically define architectures,
	// image options, repositories, and other parameters.
	// This is a highly simplified example.
	manifestGenerator := imgType.Manifest(nil, distro.ImageOptions{}, nil, nil)

	// Generate the manifest
	// Note: The actual manifest generation might involve more steps
	// and configurations depending on the image type and options.
	_, err = manifestGenerator.Serialize(nil, nil, nil, nil)
	if err != nil {
		panic(err)
	}

	fmt.Println("Manifest generation (simplified example) complete.")
	// In a real application, you would save the manifest to a file
	// or use it with OSBuild.
}
```

**Note:** This example is highly simplified. Real-world usage involves more detailed configuration of image options, package sets, repositories, and system configurations. Refer to the [developer documentation](./docs/developer/code-manifest-generation.md) for more comprehensive examples and guidance.

### Using the `cmd/` Tools

The `cmd/` directory contains several command-line tools that are useful for development and testing purposes. These tools are not considered part of the stable library API but can be helpful for working with the `images` project.

Some of the available tools are:
*   **`cmd/build`**: Builds an image based on a manifest. (See full list below)
*   **`cmd/gen-manifests`**: Generates manifests for different distributions and image types.
*   **`cmd/list-images`**: Lists all supported combinations of distributions, architectures, and image types.
*   **`cmd/osbuild-upload-gcp`**: Uploads an image to Google Cloud.

Here's an example of how you might use `cmd/build` to build an image. First, you would need a manifest file (e.g., `manifest.json`). You can generate one using `cmd/gen-manifests` or by writing your own.

```bash
# Example: Build an image using a pre-generated manifest
# Ensure osbuild is installed and configured on your system.

# 1. (Optional) Generate a manifest if you don't have one:
#    go run ./cmd/gen-manifests --distro fedora-39 --image-type qcow2 --output manifest.json

# 2. Build the image using the manifest:
#    sudo go run ./cmd/build ./manifest.json --output-dir ./output --store ./osbuild-store
#
#    - ./manifest.json: Path to the OSBuild manifest.
#    - --output-dir ./output: Directory where the final image will be saved.
#    - --store ./osbuild-store: Directory to use as the OSBuild store.
```

**Important:**
- The `cmd/` tools often require OSBuild and its dependencies to be installed on your system.
- The specific flags and usage may vary. Use the `--help` flag with each command for detailed information (e.g., `go run ./cmd/build --help`).
- These tools are primarily for development and testing. For production image building, it's recommended to use established OSBuild workflows.

## Command-Line Tools (`cmd/`)

The `cmd/` directory contains a suite of command-line interface (CLI) tools. These binaries are implemented for development, testing, and utility purposes, and are not considered part of the stable `images` library API. They often provide convenient wrappers or functionalities related to image building, manifest generation, and cloud interactions.

Below is a list of the available tools and their functions:

*   **`cmd/boot-aws`**: Manages AWS EC2 instances by handling image uploads, instance creation, and teardown. It supports setting up instances with specified SSH keys and user data, and can execute commands on these instances.
*   **`cmd/build`**: Builds an image based on a provided configuration file, distribution, and image type. It handles manifest generation and invokes `osbuild`.
*   **`cmd/gen-manifests`**: Generates OSBuild manifests for various distributions, architectures, and image types, based on a configuration map. It supports parallel generation and can resolve package sets, container images, and ostree commits.
*   **`cmd/list-images`**: Lists all supported combinations of distributions, architectures, and image types. Output can be filtered and formatted as JSON.
*   **`cmd/osbuild-composer-image-definitions`**: Generates a JSON map of available image type definitions for each distribution and architecture, typically for consumption by osbuild-composer.
*   **`cmd/osbuild-mock-openid-provider`**: Implements a mock OpenID provider with `/certs` and `/token` endpoints for testing OAuth2 authentication flows.
*   **`cmd/osbuild-package-sets`**: Dumps a JSON object containing all package sets for a specified distribution, architecture, and image type.
*   **`cmd/osbuild-playground`**: A development tool for experimenting with image building, allowing users to specify an image type, distribution, architecture, and custom options.
*   **`cmd/osbuild-upload-azure`**: Uploads an image file to Azure Blob Storage as a page blob and can tag the blob.
*   **`cmd/osbuild-upload-container`**: Uploads an OCI-archive (tarball) to a container registry.
*   **`cmd/osbuild-upload-gcp`**: Uploads an image to Google Cloud Storage, imports it into Compute Engine, and can share it with specified accounts.
*   **`cmd/osbuild-upload-generic-s3`**: Uploads an image file to an S3-compatible object storage service.
*   **`cmd/osbuild-upload-oci`**: Uploads an image file to Oracle Cloud Infrastructure (OCI) object storage and can create an image in the OCI compute service.

### OSBuild Tool Kit (`cmd/otk/`)

The `cmd/otk/` subdirectory hosts a collection of specialized tools, likely part of an "OSBuild Tool Kit". These tools are designed to perform granular tasks in the image building process, often by consuming and producing JSON data structures that can be piped together. They are particularly focused on disk partitioning, filesystem setup, and component resolution.

*   **`cmd/otk/osbuild-gen-partition-table`**: Generates a detailed partition table specification (as JSON) based on user-defined properties, partitions, and modifications.
*   **`cmd/otk/osbuild-make-fstab-stage`**: Consumes a partition table specification (from `osbuild-gen-partition-table`) and produces an OSBuild `fstab` stage.
*   **`cmd/otk/osbuild-make-grub2-inst-stage`**: Consumes a partition table specification and platform information to produce an OSBuild GRUB2 installation stage.
*   **`cmd/otk/osbuild-make-ostree-source`**: Takes resolved ostree commit information and generates an OSBuild ostree source definition.
*   **`cmd/otk/osbuild-make-partition-mounts-devices`**: Consumes a partition table specification and generates the corresponding OSBuild mount and device definitions.
*   **`cmd/otk/osbuild-make-partition-stages`**: Consumes a partition table specification and produces the OSBuild stages necessary to create and format those partitions.
*   **`cmd/otk/osbuild-resolve-containers`**: Resolves a list of container image references for a given architecture, providing their digests and image IDs.
*   **`cmd/otk/osbuild-resolve-ostree-commit`**: Resolves an ostree commit for a given URL and ref, returning its checksum and other metadata.

## Supported Image Types

The `images` library and its associated tools can generate a wide variety of disk images, installers, and container images for different platforms and use cases. The exact list of supported image types can vary by distribution and architecture.

Some of the common and representative image types include:

*   **Disk Images:**
    *   `qcow2`: QEMU Copy-On-Write image, widely used in virtualization.
    *   `vhd`: Virtual Hard Disk, often used with Microsoft Azure.
    *   `vmdk`: Virtual Machine Disk, used with VMware virtualization products.
    *   `ami`: Amazon Machine Image, for use with AWS EC2.
    *   `gce`: Google Compute Engine image, for use on Google Cloud Platform.
*   **Installers:**
    *   Various Anaconda-based installers.
    *   OSTree-based installers and simplified installers.
*   **Containers & Archives:**
    *   `container`: OCI/Docker compatible container images.
    *   `tar`: Tar archives, often used for filesystem images or initial ramdisks.
*   **Specialized Images:**
    *   `bootc-disk`: Bootable disk images using `bootc`.

This is not an exhaustive list. For a more comprehensive overview of available image types, their specific options, and how they are defined, please refer to the [Image Types Documentation](./docs/image-types/).

## Supported Distributions

The `images` library is designed to generate images for a variety of Linux distributions. Core support, with distribution-specific logic and package sets, is built-in for:

*   **Fedora**
*   **Red Hat Enterprise Linux (RHEL)** and its derivatives. Specific versions like RHEL 7, 8, 9, and 10 have dedicated definitions.

Support for other distributions, such as **CentOS Stream**, **AlmaLinux**, **Rocky Linux**, and others, can often be achieved by providing custom repository configurations.

For a comprehensive list of distributions against which this project is tested, and for examples of repository configurations, please consult the files within the [test/data/repositories/](./test/data/repositories/) directory. This directory contains repository definitions used in automated testing and can serve as a guide for targeting additional distributions or specific versions.

## Project Overview

This section provides high-level information about the OSBuild Images project.

### Project Links

*   **Website**: <https://www.osbuild.org>
*   **Bug Tracker**: <https://github.com/osbuild/images/issues>
*   **Discussions**: <https://github.com/orgs/osbuild/discussions>
*   **Matrix (chat)**: [Image Builder channel on Fedora Chat](https://matrix.to/#/#image-builder:fedoraproject.org?web-instance[element.io]=chat.fedoraproject.org)
*   **Changelog**: <https://github.com/osbuild/images/releases>

### Core Principles

1.  The image definitions API is internal and can therefore be broken. The blueprint API is the stable API.
2.  Nonsensical manifests should not compile (at the Golang level).
3.  OSBuild units (stages, sources, inputs, mounts, devices) should be directly mapped into Go objects.
4.  Image definitions don’t test distributions that are end-of-life. Respective code-paths should be dropped.
5.  Image definitions need to support the oldest supported target distribution.

## Contributing

We welcome contributions to the OSBuild Images project!

*   **Main Developer Guide**: Please refer to the [central developer guide](https://www.osbuild.org/docs/developer-guide/index) to learn about our general workflow, code style, and more.
*   **Local Development**: See also the [local developer documentation](./docs/developer) for useful information about working with this specific project.

### YAML-based Image Definitions

More and more parts of the library are converted to use YAML to define core
parts of an image. See this [example](./pkg/distro/packagesets/fedora/)
directory. For local development that just changes the YAML-based
definitions, the library can be forced to use alternative YAML directories.

For example, if there is a `./my-yaml/fedora/package-sets.yaml`, it can be
used via:
```console
$ IMAGE_BUILDER_EXPERIMENTAL=yamldir=./my-yaml image-builder build minimal-raw --distro fedora-42
```
**Warning:** This is an experimental and unsupported feature that should never be used in production and may change at any time.

We plan to eventually stabilize this as a switch for the `image-builder` tool, but for now, this environment option is required.

## Build Requirements

To build and fully test the `images` library and its associated tools, you'll need several dependencies. These are categorized below. The `Containerfile` in the repository provides a reference environment with most of these dependencies installed.

**1. Go Environment:**

*   **Go Compiler:** A recent version of Go is required. You can find installation instructions at [golang.org/doc/install](https://golang.org/doc/install).
    *   *Fedora/RHEL-based:* `sudo dnf install golang`
    *   *Ubuntu/Debian-based:* `sudo apt install golang`
    *   *macOS (using Homebrew):* `brew install go`

**2. Core Build Dependencies (for the Go library):**

These are generally required to compile the core library.

*   **gpgme-devel (GPGME Development Libraries):** Needed for GPG signature handling.
    *   *Fedora/RHEL-based:* `sudo dnf install gpgme-devel`
    *   *Ubuntu/Debian-based:* `sudo apt install libgpgme-dev`
    *   *macOS (using Homebrew):* `brew install gpgme` (Note: Homebrew usually installs headers automatically)

**3. Optional Build Dependencies (for specific Go packages or build tags):**

These are needed if you are working with or testing specific Go packages within the `images` repository.

*   **btrfs-progs-devel (BTRFS Development Libraries):** Required for building `pkg/container` and related tests that use the BTRFS graph driver.
    *   *Fedora/RHEL-based:* `sudo dnf install btrfs-progs-devel`
    *   *Ubuntu/Debian-based:* `sudo apt install libbtrfs-dev`
    *   *Note:* Can often be skipped if not using BTRFS functionalities, by using build tags like `exclude_graphdriver_btrfs`.
*   **device-mapper-devel (Device Mapper Development Libraries):** Required for building `pkg/container` and related tests that use the Device Mapper graph driver.
    *   *Fedora/RHEL-based:* `sudo dnf install device-mapper-devel`
    *   *Ubuntu/Debian-based:* `sudo apt install libdevmapper-dev`
    *   *Note:* Can often be skipped if not using Device Mapper functionalities, by using build tags like `exclude_graphdriver_devicemapper`.
*   **krb5-devel (Kerberos Development Libraries):** Required for building `pkg/upload/koji` and its associated tests.
    *   *Fedora/RHEL-based:* `sudo dnf install krb5-devel`
    *   *Ubuntu/Debian-based:* `sudo apt install libkrb5-dev`
    *   *macOS (using Homebrew):* `brew install krb5`

**4. Runtime Dependencies (for tests and `cmd/` tools):**

These are not strictly build dependencies for the library itself, but are required to run tests or the command-line tools in `cmd/`.

*   **osbuild:** The core OSBuild tool. Essential for running integration tests and using `cmd/build`.
    *   *Installation instructions:* Follow the guide at [osbuild.org/docs/getting-started/installation/](https://www.osbuild.org/docs/getting-started/installation/). Installation often involves adding a custom repository.
    *   *Fedora:* `sudo dnf install osbuild osbuild-selinux` (and other subpackages as needed)
*   **osbuild-depsolve-dnf:** Used by `pkg/dnfjson` (which is used in some tests) and by tools like `cmd/gen-manifests` and `cmd/build` for package resolution.
    *   *Fedora/RHEL-based:* `sudo dnf install osbuild-depsolve-dnf`
    *   *Note:* Typically installed as a dependency of `osbuild` on Fedora.
*   **Python 3:** Some test scripts or utility scripts might rely on Python 3.
    *   *Fedora/RHEL-based:* `sudo dnf install python3`
    *   *Ubuntu/Debian-based:* `sudo apt install python3`
    *   *macOS (using Homebrew):* `brew install python`

**Managing Go Module Dependencies:**

Go module dependencies are defined in the `go.mod` file and are managed automatically by the Go toolchain. You can download them using:
```console
go mod download
```

**Note on Build Tags:**

As mentioned, some dependencies like `btrfs-progs-devel` and `device-mapper-devel` are only needed for specific parts of the codebase (primarily `pkg/container`). If you are not working on these parts, you might be able to exclude them using Go build tags (e.g., `go build -tags exclude_graphdriver_btrfs`). Refer to specific package documentation or build scripts for more details.
The `Containerfile` in the repository is a good reference for a complete build and test environment, especially for Fedora.

## Repository Access

*   **Web**: <https://github.com/osbuild/images>
*   **HTTPS**: `https://github.com/osbuild/images.git`
*   **SSH**: `git@github.com:osbuild/images.git`

## Pull Request Gating

Each pull request against `images` starts a series of automated tests. These tests run via GitHub Actions and GitLab CI. Each push to a pull request will automatically launch these tests.

## License

*   **Apache-2.0**
*   See the `LICENSE` file for details.
