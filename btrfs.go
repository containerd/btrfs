package btrfs

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	"github.com/pkg/errors"
)

// IsSubvolume returns nil if the path is a valid subvolume. An error is
// returned if the path does not exist or the path is not a valid subvolume.
func IsSubvolume(path string) error {
	fi, err := os.Lstat(path)
	if err != nil {
		return err
	}

	if err := isFileInfoSubvol(fi); err != nil {
		return err
	}

	var statfs syscall.Statfs_t
	if err := syscall.Statfs(path, &statfs); err != nil {
		return err
	}

	return isStatfsSubvol(&statfs)
}

// SubvolCreate creates a subvolume at the provided path.
func SubvolCreate(path string) error {
	dir, name := filepath.Split(path)

	if err := IsSubvolume(dir); err != nil {
		return errors.Wrapf(err, "%v is not a subvolume", dir)
	}

	fp, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer fp.Close()

	var args volArgs
	args.Fd = fp.Fd()
	copy(args.Name[:], []byte(name))

	if err := ioctl(fp.Fd(), ioctlSubvolCreate, uintptr(unsafe.Pointer(&args))); err != nil {
		return errors.Wrap(err, "btrfs subvolume create failed")
	}

	return nil
}

// SubvolSnapshot creates a snapshot in dst from src. If readonly is true, the
// snapshot will be readonly.
func SubvolSnapshot(dst, src string, readonly bool) error {
	dstdir, dstname := filepath.Split(dst)

	dstfp, err := openSubvolDir(dstdir)
	if err != nil {
		return errors.Wrapf(err, "opening snapshot desination subvolume failed")
	}
	defer dstfp.Close()

	srcfp, err := openSubvolDir(src)
	if err != nil {
		return errors.Wrapf(err, "opening snapshot source subvolume failed")
	}

	// dstdir is the ioctl arg, wile srcdir gets set on the args
	var args volArgsV2
	copy(args.Name[:], dstname)
	args.Fd = srcfp.Fd()

	if readonly {
		args.Flags |= flagSubvolReadonly
	}

	if err := ioctl(dstfp.Fd(), ioctlSnapCreateV2, uintptr(unsafe.Pointer(&args))); err != nil {
		return errors.Wrapf(err, "snapshot create failed")
	}

	return nil
}

func SubvolDelete(path string) error {
	fmt.Println("delete", path)
	dir, name := filepath.Split(path)
	fp, err := openSubvolDir(dir)
	if err != nil {
		return errors.Wrapf(err, "failed opening %v", path)
	}
	defer fp.Close()

	// remove child subvolumes
	if err := filepath.Walk(path, func(p string, fi os.FileInfo, err error) error {
		if err != nil {
			if os.IsNotExist(err) || p == path {
				return nil
			}

			return errors.Wrapf(err, "failed walking subvolume %v", p)
		}

		if !fi.IsDir() {
			return nil // just ignore it!
		}

		if p == path {
			return nil
		}

		if err := isFileInfoSubvol(fi); err != nil {
			return err
		}

		if err := SubvolDelete(p); err != nil {
			return err
		}

		return filepath.SkipDir // children get walked by call above.
	}); err != nil {
		return err
	}

	var args volArgs
	copy(args.Name[:], name)

	if err := ioctl(fp.Fd(), ioctlSnapDestroy, uintptr(unsafe.Pointer(&args))); err != nil {
		return errors.Wrapf(err, "failed removing subvolume %v", path)
	}

	return nil
}

func openSubvolDir(path string) (*os.File, error) {
	if err := IsSubvolume(path); err != nil {
		return nil, errors.Wrapf(err, "%v must be a subvolume", path)
	}

	fp, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "opening %v as subvolume failed", path)
	}

	return fp, nil
}

func isStatfsSubvol(statfs *syscall.Statfs_t) error {
	if statfs.Type != superMagic {
		return errors.Errorf("not a btrfs filesystem")
	}

	return nil
}

func isFileInfoSubvol(fi os.FileInfo) error {
	if !fi.IsDir() {
		errors.Errorf("must be a directory")
	}

	stat := fi.Sys().(*syscall.Stat_t)

	if stat.Ino != firstFreeObjectID {
		return errors.Errorf("incorrect inode type")
	}

	return nil
}
