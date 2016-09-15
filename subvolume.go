package btrfs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

func checkSubVolumeName(name string) bool {
	return name != "" && name[0] != 0 && !strings.ContainsRune(name, '/') &&
		name != "." && name != ".."
}

func IsSubVolume(path string) (bool, error) {
	var st syscall.Stat_t
	if err := syscall.Stat(path, &st); err != nil {
		return false, err
	}
	if st.Ino != firstFreeObjectid ||
		st.Mode&syscall.S_IFMT != syscall.S_IFDIR {
		return false, nil
	}
	return isBtrfs(path)
}

func CreateSubVolume(path string) error {
	var inherit *btrfs_qgroup_inherit // TODO

	cpath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	newName := filepath.Base(cpath)
	dstDir := filepath.Dir(cpath)
	if !checkSubVolumeName(newName) {
		return fmt.Errorf("invalid subvolume name: %s", newName)
	} else if len(newName) >= volNameMax {
		return fmt.Errorf("subvolume name too long: %s", newName)
	}
	dst, err := openDir(dstDir)
	if err != nil {
		return err
	}
	defer dst.Close()
	if inherit != nil {
		panic("not implemented") // TODO
		args := btrfs_ioctl_vol_args_v2{
			flags: subvolQGroupInherit,
			btrfs_ioctl_vol_args_v2_u1: btrfs_ioctl_vol_args_v2_u1{
				//size: 	qgroup_inherit_size(inherit),
				qgroup_inherit: inherit,
			},
		}
		copy(args.name[:], newName)
		return iocSubvolCreateV2(dst, &args)
	}
	var args btrfs_ioctl_vol_args
	copy(args.name[:], newName)
	return iocSubvolCreate(dst, &args)
}

func DeleteSubVolume(path string) error {
	if ok, err := IsSubVolume(path); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("not a subvolume: %s", path)
	}
	cpath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	dname := filepath.Dir(cpath)
	vname := filepath.Base(cpath)

	dir, err := openDir(dname)
	if err != nil {
		return err
	}
	defer dir.Close()
	var args btrfs_ioctl_vol_args
	copy(args.name[:], vname)
	return iocSnapDestroy(dir, &args)
}

func SnapshotSubVolume(subvol, dst string, ro bool) error {
	if ok, err := IsSubVolume(subvol); err != nil {
		return err
	} else if !ok {
		return fmt.Errorf("not a subvolume: %s", subvol)
	}
	exists := false
	if st, err := os.Stat(dst); err != nil && !os.IsNotExist(err) {
		return err
	} else if err == nil {
		if !st.IsDir() {
			return fmt.Errorf("'%s' exists and it is not a directory", dst)
		}
		exists = true
	}
	var (
		newName string
		dstDir  string
	)
	if exists {
		newName = filepath.Base(subvol)
		dstDir = dst
	} else {
		newName = filepath.Base(dst)
		dstDir = filepath.Dir(dst)
	}
	if !checkSubVolumeName(newName) {
		return fmt.Errorf("invalid snapshot name '%s'", newName)
	} else if len(newName) >= volNameMax {
		return fmt.Errorf("snapshot name too long '%s'", newName)
	}
	fdst, err := openDir(dstDir)
	if err != nil {
		return err
	}
	// TODO: make SnapshotSubVolume a method on FS to use existing fd
	f, err := openDir(subvol)
	if err != nil {
		return err
	}
	args := btrfs_ioctl_vol_args_v2{
		fd: int64(f.Fd()),
	}
	if ro {
		args.flags |= subvolReadOnly
	}
	// TODO
	//if inherit != nil {
	//	args.flags |= subvolQGroupInherit
	//	args.size = qgroup_inherit_size(inherit)
	//	args.qgroup_inherit = inherit
	//}
	copy(args.name[:], newName)
	return iocSnapCreateV2(fdst, &args)
}
