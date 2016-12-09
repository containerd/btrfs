// Created by cgo -godefs - DO NOT EDIT
// cgo -godefs ctypes_linux.go

package btrfs

const (
	firstFreeObjectID  uint64 = 0x100
	superMagic         int64  = 0x9123683e
	flagSubvolReadonly uint64 = 0x2
)

const (
	ioctlSubvolCreate = 0x5000940e
	ioctlSnapCreate   = 0x50009401
	ioctlSnapCreateV2 = 0x50009417
	ioctlSnapDestroy  = 0x5000940f
)

type (
	volArgs struct {
		Fd   uintptr
		Name [4088]byte
	}
	volArgsV2 struct {
		Fd      uintptr
		Transid uint64
		Flags   uint64
		Anon0   [32]byte
		Name    [4040]byte
	}
)
