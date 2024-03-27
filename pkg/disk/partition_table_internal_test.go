package disk

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/osbuild/images/internal/common"
	"github.com/osbuild/images/pkg/blueprint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPartitionTableFeatures(t *testing.T) {
	type testCase struct {
		partitionType    string
		expectedFeatures partitionTableFeatures
	}
	testCases := []testCase{
		{"plain", partitionTableFeatures{XFS: true, FAT: true}},
		{"luks", partitionTableFeatures{XFS: true, FAT: true, LUKS: true}},
		{"luks+lvm", partitionTableFeatures{XFS: true, FAT: true, LUKS: true, LVM: true}},
		{"btrfs", partitionTableFeatures{XFS: true, FAT: true, Btrfs: true}},
	}

	for _, tc := range testCases {
		pt := testPartitionTables[tc.partitionType]
		assert.Equal(t, tc.expectedFeatures, pt.features())

	}
}

func TestPartitionTableDefaultPartitioningMode(t *testing.T) {
	type testCase struct {
		defaultMode   *PartitioningMode
		requestedMode PartitioningMode
		mountpoints   []blueprint.FilesystemCustomization
		isLVM         bool
	}

	testCases := []testCase{
		// default behavior
		{
			defaultMode:   nil,
			requestedMode: DefaultPartitioningMode,
			isLVM:         false,
		},
		{
			defaultMode:   common.ToPtr(DefaultPartitioningMode),
			requestedMode: DefaultPartitioningMode,
			isLVM:         false,
		},
		{
			defaultMode:   nil,
			requestedMode: AutoLVMPartitioningMode,
			isLVM:         false,
		},
		{
			defaultMode:   common.ToPtr(DefaultPartitioningMode),
			requestedMode: AutoLVMPartitioningMode,
			isLVM:         false,
		},
		{
			defaultMode:   nil,
			requestedMode: LVMPartitioningMode,
			isLVM:         true,
		},
		{
			defaultMode:   common.ToPtr(DefaultPartitioningMode),
			requestedMode: LVMPartitioningMode,
			isLVM:         true,
		},
		{
			defaultMode:   nil,
			requestedMode: RawPartitioningMode,
			isLVM:         false,
		},
		{
			defaultMode:   common.ToPtr(DefaultPartitioningMode),
			requestedMode: RawPartitioningMode,
			isLVM:         false,
		},
		// AutoLVM is the default if no default explicitly is set
		{
			defaultMode:   nil,
			requestedMode: DefaultPartitioningMode,
			mountpoints:   []blueprint.FilesystemCustomization{{Mountpoint: "/var", MinSize: 10 * GiB}},
			isLVM:         true,
		},
		{
			defaultMode:   common.ToPtr(DefaultPartitioningMode),
			requestedMode: DefaultPartitioningMode,
			mountpoints:   []blueprint.FilesystemCustomization{{Mountpoint: "/var", MinSize: 10 * GiB}},
			isLVM:         true,
		},
		// Explicit AutoLVM default mode
		{
			defaultMode:   common.ToPtr(AutoLVMPartitioningMode),
			requestedMode: DefaultPartitioningMode,
			mountpoints:   []blueprint.FilesystemCustomization{{Mountpoint: "/var", MinSize: 10 * GiB}},
			isLVM:         true,
		},
		// LVM default mode
		{
			defaultMode:   common.ToPtr(LVMPartitioningMode),
			requestedMode: DefaultPartitioningMode,
			mountpoints:   []blueprint.FilesystemCustomization{{Mountpoint: "/var", MinSize: 10 * GiB}},
			isLVM:         true,
		},
		// Raw default mode
		{
			defaultMode:   common.ToPtr(RawPartitioningMode),
			requestedMode: DefaultPartitioningMode,
			mountpoints:   []blueprint.FilesystemCustomization{{Mountpoint: "/var", MinSize: 10 * GiB}},
			isLVM:         false,
		},
		// overriding the default mode
		{
			defaultMode:   common.ToPtr(LVMPartitioningMode),
			requestedMode: RawPartitioningMode,
			mountpoints:   []blueprint.FilesystemCustomization{{Mountpoint: "/var", MinSize: 10 * GiB}},
			isLVM:         false,
		},
		{
			defaultMode:   common.ToPtr(AutoLVMPartitioningMode),
			requestedMode: RawPartitioningMode,
			mountpoints:   []blueprint.FilesystemCustomization{{Mountpoint: "/var", MinSize: 10 * GiB}},
			isLVM:         false,
		},
		{
			defaultMode:   common.ToPtr(RawPartitioningMode),
			requestedMode: LVMPartitioningMode,
			mountpoints:   []blueprint.FilesystemCustomization{{Mountpoint: "/var", MinSize: 10 * GiB}},
			isLVM:         true,
		},
		{
			defaultMode:   common.ToPtr(RawPartitioningMode),
			requestedMode: AutoLVMPartitioningMode,
			mountpoints:   []blueprint.FilesystemCustomization{{Mountpoint: "/var", MinSize: 10 * GiB}},
			isLVM:         true,
		},
	}

	for idx, tc := range testCases {
		t.Run(fmt.Sprintf("test case %d", idx), func(t *testing.T) {
			pt := testPartitionTables["plain"]
			pt.DefaultPartitioningMode = tc.defaultMode

			/* #nosec G404 */
			newPT, err := NewPartitionTable(&pt, tc.mountpoints, 20*GiB, tc.requestedMode, nil, rand.New(rand.NewSource(0)))
			require.NoError(t, err)

			ptFeatures := newPT.features()
			if tc.isLVM {
				require.True(t, ptFeatures.LVM)
			} else {
				require.False(t, ptFeatures.LVM)
			}
		})
	}
}
