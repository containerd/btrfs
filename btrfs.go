package btrfs

import (
	"github.com/dennwc/btrfs/ioctl"
	"os"
)

//go:generate go run cmd/hgen.go -p btrfs -o btrfs_tree_hc.go btrfs_tree.h
//go:generate go fmt -w btrfs_tree_hc.go

const SuperMagic = 0x9123683E

func Open(path string) (*FS, error) {
	if ok, err := IsSubVolume(path); err != nil {
		return nil, err
	} else if !ok {
		return nil, ErrNotBtrfs{Path: path}
	}
	dir, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return &FS{f: dir}, nil
}

type FS struct {
	f *os.File
}

func (f *FS) Close() error {
	return f.f.Close()
}

type Info struct {
	MaxID          uint64
	NumDevices     uint64
	FSID           FSID
	NodeSize       uint32
	SectorSize     uint32
	CloneAlignment uint32
}

func (f *FS) Info() (out Info, err error) {
	var arg btrfs_ioctl_fs_info_args
	if err = ioctl.Do(f.f, _BTRFS_IOC_FS_INFO, &arg); err != nil {
		return
	}
	out = Info{
		MaxID:          arg.max_id,
		NumDevices:     arg.num_devices,
		FSID:           arg.fsid,
		NodeSize:       arg.nodesize,
		SectorSize:     arg.sectorsize,
		CloneAlignment: arg.clone_alignment,
	}
	return
}

type DevStats struct {
	WriteErrs uint64
	ReadErrs  uint64
	FlushErrs uint64
	// Checksum error, bytenr error or contents is illegal: this is an
	// indication that the block was damaged during read or write, or written to
	// wrong location or read from wrong location.
	CorruptionErrs uint64
	// An indication that blocks have not been written.
	GenerationErrs uint64
	Unknown        []uint64
}

func (f *FS) GetDevStats(id uint64) (out DevStats, err error) {
	var arg btrfs_ioctl_get_dev_stats
	arg.devid = id
	//arg.nr_items = _BTRFS_DEV_STAT_VALUES_MAX
	arg.flags = 0
	if err = ioctl.Do(f.f, _BTRFS_IOC_GET_DEV_STATS, &arg); err != nil {
		return
	}
	i := 0
	out.WriteErrs = arg.values[i]
	i++
	out.ReadErrs = arg.values[i]
	i++
	out.FlushErrs = arg.values[i]
	i++
	out.CorruptionErrs = arg.values[i]
	i++
	out.GenerationErrs = arg.values[i]
	i++
	if int(arg.nr_items) > i {
		out.Unknown = arg.values[i:arg.nr_items]
	}
	return
}

type FSFeatureFlags struct {
	Compatible   FeatureFlags
	CompatibleRO FeatureFlags
	Incompatible IncompatFeatures
}

func (f *FS) GetFeatures() (out FSFeatureFlags, err error) {
	var arg btrfs_ioctl_feature_flags
	if err = ioctl.Do(f.f, _BTRFS_IOC_GET_FEATURES, &arg); err != nil {
		return
	}
	out = FSFeatureFlags{
		Compatible:   arg.compat_flags,
		CompatibleRO: arg.compat_ro_flags,
		Incompatible: arg.incompat_flags,
	}
	return
}

func (f *FS) GetSupportedFeatures() (out FSFeatureFlags, err error) {
	var arg [3]btrfs_ioctl_feature_flags
	if err = ioctl.Do(f.f, _BTRFS_IOC_GET_SUPPORTED_FEATURES, &arg); err != nil {
		return
	}
	out = FSFeatureFlags{
		Compatible:   arg[0].compat_flags,
		CompatibleRO: arg[0].compat_ro_flags,
		Incompatible: arg[0].incompat_flags,
	}
	//for i, a := range arg {
	//	out[i] = FSFeatureFlags{
	//		Compatible:   a.compat_flags,
	//		CompatibleRO: a.compat_ro_flags,
	//		Incompatible: a.incompat_flags,
	//	}
	//}
	return
}

func (f *FS) Sync() (err error) {
	if err = ioctl.Do(f.f, _BTRFS_IOC_START_SYNC, nil); err != nil {
		return
	}
	return ioctl.Do(f.f, _BTRFS_IOC_WAIT_SYNC, nil)
}
