package container

import "github.com/osbuild/images/pkg/customizations/fsnode"

// NetworkBackend is the type of network backend used by Podman.
type NetworkBackend string

const (
	NetworkBackendCNI     NetworkBackend = "cni"
	NetworkBackendNetavag NetworkBackend = "netavark"
)

func GenDefaultNetworkBackendFile(backend *NetworkBackend) (*fsnode.File, error) {
	if backend == nil {
		return nil, nil
	}

	file, err := fsnode.NewFile("/var/lib/containers/storage/defaultNetworkBackend", nil, nil, nil, []byte(*backend))
	if err != nil {
		return nil, err
	}
	return file, nil
}
