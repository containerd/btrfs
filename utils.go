package btrfs

import (
	"fmt"
	"os"
	"syscall"
	"unsafe"
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

type rawItem struct {
	TransID  uint64
	ObjectID uint64
	Type     uint32
	Offset   uint64
	Data     []byte
}

func treeSearchRaw(f *os.File, key btrfs_ioctl_search_key) (out []rawItem, _ error) {
	args := btrfs_ioctl_search_args{
		key: key,
	}
	if err := iocTreeSearch(f, &args); err != nil {
		return nil, err
	}
	out = make([]rawItem, 0, args.key.nr_items)
	buf := args.buf[:]
	for i := 0; i < int(args.key.nr_items); i++ {
		h := (*btrfs_ioctl_search_header)(unsafe.Pointer(&buf[0]))
		buf = buf[unsafe.Sizeof(btrfs_ioctl_search_header{}):]
		out = append(out, rawItem{
			TransID:  h.transid,
			ObjectID: h.objectid,
			Type:     h.typ,
			Offset:   h.offset,
			Data:     buf[:h.len], // TODO: reallocate?
		})
		buf = buf[h.len:]
	}
	return out, nil
}
