/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package send

import (
	"encoding/binary"
	"io"
	"strconv"
)

var sendEndianess = binary.LittleEndian

const (
	sendStreamMagic     = "btrfs-stream\x00"
	sendStreamMagicSize = len(sendStreamMagic)
	sendStreamVersion   = 1
)

const (
	sendBufSize  = 64 * 1024
	sendReadSize = 48 * 1024
)

const cmdHeaderSize = 10

type cmdHeader struct {
	Len uint32 // len excluding the header
	Cmd CmdType
	Crc uint32 // crc including the header with zero crc field
}

func (h *cmdHeader) Size() int { return cmdHeaderSize }
func (h *cmdHeader) Unmarshal(p []byte) error {
	if len(p) < cmdHeaderSize {
		return io.ErrUnexpectedEOF
	}
	h.Len = sendEndianess.Uint32(p[0:])
	h.Cmd = CmdType(sendEndianess.Uint16(p[4:]))
	h.Crc = sendEndianess.Uint32(p[6:])
	return nil
}

const tlvHeaderSize = 4

type tlvHeader struct {
	Type uint16
	Len  uint16 // len excluding the header
}

func (h *tlvHeader) Size() int { return tlvHeaderSize }
func (h *tlvHeader) Unmarshal(p []byte) error {
	if len(p) < tlvHeaderSize {
		return io.ErrUnexpectedEOF
	}
	h.Type = sendEndianess.Uint16(p[0:])
	h.Len = sendEndianess.Uint16(p[2:])
	return nil
}

type CmdType uint16

func (c CmdType) String() string {
	var name string
	if int(c) < len(cmdTypeNames) {
		name = cmdTypeNames[int(c)]
	}
	if name != "" {
		return name
	}
	return strconv.FormatInt(int64(c), 16)
}

var cmdTypeNames = []string{
	"<zero>",

	"subvol",
	"snapshot",

	"mkfile",
	"mkdir",
	"mknod",
	"mkfifo",
	"mksock",
	"symlink",

	"rename",
	"link",
	"unlink",
	"rmdir",

	"set_xattr",
	"remove_xattr",

	"write",
	"clone",

	"truncate",
	"chmod",
	"chown",
	"utimes",

	"end",
	"update_extent",
	"<max>",
}

const (
	sendCmdUnspec = CmdType(iota)

	sendCmdSubvol
	sendCmdSnapshot

	sendCmdMkfile
	sendCmdMkdir
	sendCmdMknod
	sendCmdMkfifo
	sendCmdMksock
	sendCmdSymlink

	sendCmdRename
	sendCmdLink
	sendCmdUnlink
	sendCmdRmdir

	sendCmdSetXattr
	sendCmdRemoveXattr

	sendCmdWrite
	sendCmdClone

	sendCmdTruncate
	sendCmdChmod
	sendCmdChown
	sendCmdUtimes

	sendCmdEnd
	sendCmdUpdateExtent
	_sendCmdMax
)

const sendCmdMax = _sendCmdMax - 1

type sendCmdAttr uint16

func (c sendCmdAttr) String() string {
	var name string
	if int(c) < len(sendAttrNames) {
		name = sendAttrNames[int(c)]
	}
	if name != "" {
		return name
	}
	return strconv.FormatInt(int64(c), 16)
}

const (
	sendAttrUnspec = sendCmdAttr(iota)

	sendAttrUuid
	sendAttrCtransid

	sendAttrIno
	sendAttrSize
	sendAttrMode
	sendAttrUid
	sendAttrGid
	sendAttrRdev
	sendAttrCtime
	sendAttrMtime
	sendAttrAtime
	sendAttrOtime

	sendAttrXattrName
	sendAttrXattrData

	sendAttrPath
	sendAttrPathTo
	sendAttrPathLink

	sendAttrFileOffset
	sendAttrData

	sendAttrCloneUuid
	sendAttrCloneCtransid
	sendAttrClonePath
	sendAttrCloneOffset
	sendAttrCloneLen

	_sendAttrMax
)
const sendAttrMax = _sendAttrMax - 1

var sendAttrNames = []string{
	"<zero>",

	"uuid",
	"ctransid",

	"ino",
	"size",
	"mode",
	"uid",
	"gid",
	"rdev",
	"ctime",
	"mtime",
	"atime",
	"otime",

	"xattrname",
	"xattrdata",

	"path",
	"pathto",
	"pathlink",

	"fileoffset",
	"data",

	"cloneuuid",
	"clonectransid",
	"clonepath",
	"cloneoffset",
	"clonelen",

	"<max>",
}
