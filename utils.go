package btrfs

import (
	"fmt"
	"os"
	"syscall"
)

func isBtrfs(path string) (bool, error) {
	var stfs syscall.Statfs_t
	if err := syscall.Statfs(path, &stfs); err != nil {
		return false, err
	}
	return stfs.Type == SuperMagic, nil
}

// openDir does the following checks before calling Open:
// 1: path is in a btrfs filesystem
// 2: path is a directory
func openDir(path string) (*os.File, error) {
	if ok, err := isBtrfs(path); err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrNotBtrfs{Path: path}
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	} else if st, err := file.Stat(); err != nil {
		file.Close()
		return nil, err
	} else if !st.IsDir() {
		file.Close()
		return nil, fmt.Errorf("not a directory: %s", path)
	}
	return file, nil
}
