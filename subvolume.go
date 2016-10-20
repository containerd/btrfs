package btrfs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func checkSubVolumeName(name string) bool {
	return name != "" && name[0] != 0 && !strings.ContainsRune(name, '/') &&
		name != "." && name != ".."
}

func IsSubVolume(path string) (bool, error) {
	var st syscall.Stat_t
	if err := syscall.Stat(path, &st); err != nil {
		return false, &os.PathError{Op: "stat", Path: path, Err: err}
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
	defer fdst.Close()
	// TODO: make SnapshotSubVolume a method on FS to use existing fd
	f, err := openDir(subvol)
	if err != nil {
		return fmt.Errorf("cannot open dest dir: %v", err)
	}
	defer f.Close()
	args := btrfs_ioctl_vol_args_v2{
		fd: int64(f.Fd()),
	}
	if ro {
		args.flags |= SubvolReadOnly
	}
	// TODO
	//if inherit != nil {
	//	args.flags |= subvolQGroupInherit
	//	args.size = qgroup_inherit_size(inherit)
	//	args.qgroup_inherit = inherit
	//}
	copy(args.name[:], newName)
	if err := iocSnapCreateV2(fdst, &args); err != nil {
		return fmt.Errorf("ioc failed: %v", err)
	}
	return nil
}

type Subvolume struct {
	ObjectID     uint64
	TransID      uint64
	Name         string
	RefTree      uint64
	DirID        uint64
	Gen          uint64
	OGen         uint64
	Flags        uint64
	UUID         UUID
	ParentUUID   UUID
	ReceivedUUID UUID
	OTime        time.Time
	CTime        time.Time
}

func listSubVolumes(f *os.File) (map[uint64]Subvolume, error) {
	sk := btrfs_ioctl_search_key{
		// search in the tree of tree roots
		tree_id: 1,

		// Set the min and max to backref keys. The search will
		// only send back this type of key now.
		min_type: rootItemKey,
		max_type: rootBackrefKey,

		min_objectid: firstFreeObjectid,

		// Set all the other params to the max, we'll take any objectid
		// and any trans.
		max_objectid: lastFreeObjectid,
		max_offset:   maxUint64,
		max_transid:  maxUint64,

		nr_items: 4096, // just a big number, doesn't matter much
	}
	m := make(map[uint64]Subvolume)
	for {
		out, err := treeSearchRaw(f, sk)
		if err != nil {
			return nil, err
		} else if len(out) == 0 {
			break
		}
		for _, obj := range out {
			switch obj.Type {
			case rootBackrefKey:
				ref := asRootRef(obj.Data)
				o := m[obj.ObjectID]
				o.TransID = obj.TransID
				o.ObjectID = obj.ObjectID
				o.RefTree = obj.Offset
				o.DirID = ref.DirID
				o.Name = ref.Name
				m[obj.ObjectID] = o
			case rootItemKey:
				o := m[obj.ObjectID]
				o.TransID = obj.TransID
				o.ObjectID = obj.ObjectID
				// TODO: decode whole object?
				o.Gen = asUint64(obj.Data[160:]) // size of btrfs_inode_item
				o.Flags = asUint64(obj.Data[160+6*8:])
				const sz = 439
				const toff = sz - 8*8 - 4*12
				o.CTime = asTime(obj.Data[toff+0*12:])
				o.OTime = asTime(obj.Data[toff+1*12:])
				o.OGen = asUint64(obj.Data[toff-3*8:])
				const uoff = toff - 4*8 - 3*UUIDSize
				copy(o.UUID[:], obj.Data[uoff+0*UUIDSize:])
				copy(o.ParentUUID[:], obj.Data[uoff+1*UUIDSize:])
				copy(o.ReceivedUUID[:], obj.Data[uoff+2*UUIDSize:])
				m[obj.ObjectID] = o
			}
		}
		// record the mins in key so we can make sure the
		// next search doesn't repeat this root
		last := out[len(out)-1]
		sk.min_objectid = last.ObjectID
		sk.min_type = last.Type
		sk.min_offset = last.Offset + 1
		if sk.min_offset == 0 { // overflow
			sk.min_type++
		} else {
			continue
		}
		if sk.min_type > rootBackrefKey {
			sk.min_type = rootItemKey
			sk.min_objectid++
		} else {
			continue
		}
		if sk.min_objectid > sk.max_objectid {
			break
		}
	}
	return m, nil
}
