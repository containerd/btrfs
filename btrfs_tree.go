package btrfs

import (
	"time"
	"unsafe"
)

const (
	_BTRFS_BLOCK_GROUP_TYPE_MASK = (blockGroupData |
		blockGroupSystem |
		blockGroupMetadata)
	_BTRFS_BLOCK_GROUP_PROFILE_MASK = (blockGroupRaid0 |
		blockGroupRaid1 |
		blockGroupRaid5 |
		blockGroupRaid6 |
		blockGroupDup |
		blockGroupRaid10)
)

type rootRef struct {
	DirID    uint64
	Sequence uint64
	Name     string
}

func (rootRef) btrfsSize() int { return 18 }

func asUint64(p []byte) uint64 {
	return *(*uint64)(unsafe.Pointer(&p[0]))
}

func asUint32(p []byte) uint32 {
	return *(*uint32)(unsafe.Pointer(&p[0]))
}

func asUint16(p []byte) uint16 {
	return *(*uint16)(unsafe.Pointer(&p[0]))
}

func asTime(p []byte) time.Time {
	sec, nsec := asUint64(p[0:]), asUint32(p[8:])
	return time.Unix(int64(sec), int64(nsec))
}

func asRootRef(p []byte) rootRef {
	const sz = 18
	// assuming that it is highly unsafe to have sizeof(struct) > len(data)
	// (*btrfs_root_ref)(unsafe.Pointer(&p[0])) and sizeof(btrfs_root_ref) == 24
	ref := rootRef{
		DirID:    asUint64(p[0:]),
		Sequence: asUint64(p[8:]),
	}
	if n := asUint16(p[16:]); n > 0 {
		ref.Name = string(p[sz : sz+n : sz+n])
	}
	return ref
}
