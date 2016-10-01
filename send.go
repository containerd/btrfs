package btrfs

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func Send(w io.Writer, parent string, subvols ...string) error {
	if len(subvols) == 0 {
		return nil
	}
	// use first send subvol to determine mount_root
	subvol, err := filepath.Abs(subvols[0])
	if err != nil {
		return err
	}
	mountRoot, err := findMountRoot(subvol)
	if err == os.ErrNotExist {
		return fmt.Errorf("cannot find a mountpoint for %s", subvol)
	} else if err != nil {
		return err
	}
	var (
		cloneSrc []uint64
		parentID uint64
	)
	if parent != "" {
		parent, err = filepath.Abs(parent)
		if err != nil {
			return err
		}
		f, err := os.Open(parent)
		if err != nil {
			return fmt.Errorf("cannot open parent: %v", err)
		}
		id, err := getPathRootID(f)
		f.Close()
		if err != nil {
			return fmt.Errorf("cannot get parent root id: %v", err)
		}
		parentID = id
		cloneSrc = append(cloneSrc, id)
	}
	// check all subvolumes
	paths := make([]string, 0, len(subvols))
	for _, sub := range subvols {
		sub, err = filepath.Abs(sub)
		if err != nil {
			return err
		}
		paths = append(paths, sub)
		mount, err := findMountRoot(sub)
		if err != nil {
			return err
		} else if mount != mountRoot {
			return fmt.Errorf("all subvolumes must be from the same filesystem (%s is not)", sub)
		}
		ok, err := IsReadOnly(sub)
		if err != nil {
			return err
		} else if !ok {
			return fmt.Errorf("subvolume %s is not read-only", sub)
		}
	}
	//full := len(cloneSrc) == 0
	for i, sub := range paths {
		//if len(cloneSrc) > 1 {
		//	// TODO: find_good_parent
		//}
		//if !full { // TODO
		//	cloneSrc = append(cloneSrc, )
		//}
		fs, err := Open(sub, true)
		if err != nil {
			return err
		}
		var flags uint64
		if i != 0 { // not first
			flags |= _BTRFS_SEND_FLAG_OMIT_STREAM_HEADER
		}
		if i < len(paths)-1 { // not last
			flags |= _BTRFS_SEND_FLAG_OMIT_END_CMD
		}
		err = send(w, fs.f, parentID, cloneSrc, flags)
		fs.Close()
		if err != nil {
			return fmt.Errorf("error sending %s: %v", sub, err)
		}
	}
	return nil
}

func send(w io.Writer, subvol *os.File, parent uint64, sources []uint64, flags uint64) error {
	pr, pw, err := os.Pipe()
	if err != nil {
		return err
	}
	errc := make(chan error, 1)
	go func() {
		defer pr.Close()
		_, err := io.Copy(w, pr)
		errc <- err
	}()
	fd := pw.Fd()
	wait := func() error {
		pw.Close()
		return <-errc
	}
	args := &btrfs_ioctl_send_args{
		send_fd:     int64(fd),
		parent_root: parent,
		flags:       flags,
	}
	if len(sources) != 0 {
		args.clone_sources = &sources[0]
		args.clone_sources_count = uint64(len(sources))
	}
	if err := iocSend(subvol, args); err != nil {
		wait()
		return err
	}
	return wait()
}
