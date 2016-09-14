package btrfs

import (
	"syscall"
)

func IsSubVolume(path string) (bool, error) {
	var st syscall.Stat_t
	if err := syscall.Stat(path, &st); err != nil {
		return false, err
	}
	if st.Ino != BTRFS_FIRST_FREE_OBJECTID ||
		st.Mode&syscall.S_IFMT != syscall.S_IFDIR {
		return false, nil
	}
	var stfs syscall.Statfs_t
	if err := syscall.Statfs(path, &stfs); err != nil {
		return false, err
	}
	return stfs.Type == SuperMagic, nil
}
