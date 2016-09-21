package btrfs

import (
	"encoding/binary"
	"io"
)

const (
	_BTRFS_SEND_STREAM_MAGIC   = "btrfs-stream"
	sendStreamMagicSize        = len(_BTRFS_SEND_STREAM_MAGIC)
	_BTRFS_SEND_STREAM_VERSION = 1
)

const (
	_BTRFS_SEND_BUF_SIZE  = 64 * 1024
	_BTRFS_SEND_READ_SIZE = 48 * 1024
)

type tlvType uint16

const (
	tlvU8 = tlvType(iota)
	tlvU16
	tlvU32
	tlvU64
	tlvBinary
	tlvString
	tlvUUID
	tlvTimespec
)

type streamHeader struct {
	Magic   [len(_BTRFS_SEND_STREAM_MAGIC)]byte
	Version uint32
}

type cmdHeader struct {
	Len uint32 // len excluding the header
	Cmd uint16
	Crc uint32 // crc including the header with zero crc field
}

func (h *cmdHeader) Size() int { return 10 }
func (h *cmdHeader) Unmarshal(p []byte) error {
	if len(p) < h.Size() {
		return io.ErrUnexpectedEOF
	}
	h.Len = binary.LittleEndian.Uint32(p[0:])
	h.Cmd = binary.LittleEndian.Uint16(p[4:])
	h.Crc = binary.LittleEndian.Uint32(p[6:])
	return nil
}

type tlvHeader struct {
	Type tlvType
	Len  uint16 // len excluding the header
}

func (h *tlvHeader) Size() int { return 4 }
func (h *tlvHeader) Unmarshal(p []byte) error {
	if len(p) < h.Size() {
		return io.ErrUnexpectedEOF
	}
	h.Type = tlvType(binary.LittleEndian.Uint16(p[0:]))
	h.Len = binary.LittleEndian.Uint16(p[2:])
	return nil
}

type sendCmd uint16

const (
	_BTRFS_SEND_C_UNSPEC = sendCmd(iota)

	_BTRFS_SEND_C_SUBVOL
	_BTRFS_SEND_C_SNAPSHOT

	_BTRFS_SEND_C_MKFILE
	_BTRFS_SEND_C_MKDIR
	_BTRFS_SEND_C_MKNOD
	_BTRFS_SEND_C_MKFIFO
	_BTRFS_SEND_C_MKSOCK
	_BTRFS_SEND_C_SYMLINK

	_BTRFS_SEND_C_RENAME
	_BTRFS_SEND_C_LINK
	_BTRFS_SEND_C_UNLINK
	_BTRFS_SEND_C_RMDIR

	_BTRFS_SEND_C_SET_XATTR
	_BTRFS_SEND_C_REMOVE_XATTR

	_BTRFS_SEND_C_WRITE
	_BTRFS_SEND_C_CLONE

	_BTRFS_SEND_C_TRUNCATE
	_BTRFS_SEND_C_CHMOD
	_BTRFS_SEND_C_CHOWN
	_BTRFS_SEND_C_UTIMES

	_BTRFS_SEND_C_END
	_BTRFS_SEND_C_UPDATE_EXTENT
	__BTRFS_SEND_C_MAX
)

const _BTRFS_SEND_C_MAX = __BTRFS_SEND_C_MAX - 1

type sendCmdAttr uint16

const (
	_BTRFS_SEND_A_UNSPEC = iota

	_BTRFS_SEND_A_UUID
	_BTRFS_SEND_A_CTRANSID

	_BTRFS_SEND_A_INO
	_BTRFS_SEND_A_SIZE
	_BTRFS_SEND_A_MODE
	_BTRFS_SEND_A_UID
	_BTRFS_SEND_A_GID
	_BTRFS_SEND_A_RDEV
	_BTRFS_SEND_A_CTIME
	_BTRFS_SEND_A_MTIME
	_BTRFS_SEND_A_ATIME
	_BTRFS_SEND_A_OTIME

	_BTRFS_SEND_A_XATTR_NAME
	_BTRFS_SEND_A_XATTR_DATA

	_BTRFS_SEND_A_PATH
	_BTRFS_SEND_A_PATH_TO
	_BTRFS_SEND_A_PATH_LINK

	_BTRFS_SEND_A_FILE_OFFSET
	_BTRFS_SEND_A_DATA

	_BTRFS_SEND_A_CLONE_UUID
	_BTRFS_SEND_A_CLONE_CTRANSID
	_BTRFS_SEND_A_CLONE_PATH
	_BTRFS_SEND_A_CLONE_OFFSET
	_BTRFS_SEND_A_CLONE_LEN

	__BTRFS_SEND_A_MAX
)
const _BTRFS_SEND_A_MAX = __BTRFS_SEND_A_MAX - 1
