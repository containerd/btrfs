package btrfs

import "os"

func getPathRootID(file *os.File) (uint64, error) {
	args := btrfs_ioctl_ino_lookup_args{
		objectid: firstFreeObjectid,
	}
	if err := iocInoLookup(file, &args); err != nil {
		return 0, err
	}
	return args.treeid, nil
}
