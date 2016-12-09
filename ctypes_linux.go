// +build ignore

package btrfs

// plucks the go definitions for btrfs ioctl calls

/*
#include <stddef.h>
#include <linux/magic.h>
#include <btrfs/ioctl.h>
#include <btrfs/ctree.h>
*/
import "C"

const (
	firstFreeObjectID  uint64 = C.BTRFS_FIRST_FREE_OBJECTID
	superMagic         int64  = C.BTRFS_SUPER_MAGIC
	flagSubvolReadonly uint64 = C.BTRFS_SUBVOL_RDONLY
)

// ioctl requests
const (
	ioctlSubvolCreate = C.BTRFS_IOC_SUBVOL_CREATE
	ioctlSnapCreate   = C.BTRFS_IOC_SNAP_CREATE
	ioctlSnapCreateV2 = C.BTRFS_IOC_SNAP_CREATE_V2
	ioctlSnapDestroy  = C.BTRFS_IOC_SNAP_DESTROY
)

type (
	volArgs   C.struct_btrfs_ioctl_vol_args
	volArgsV2 C.struct_btrfs_ioctl_vol_args_v2
)
