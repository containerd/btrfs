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
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/containerd/btrfs/v2"
)

func NewStreamReader(r io.Reader) (*StreamReader, error) {
	// read magic and version
	buf := make([]byte, len(sendStreamMagic)+4)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return nil, fmt.Errorf("cannot read magic: %v", err)
	} else if string(buf[:sendStreamMagicSize]) != sendStreamMagic {
		return nil, errors.New("unexpected stream header")
	}
	version := sendEndianess.Uint32(buf[sendStreamMagicSize:])
	if version != sendStreamVersion {
		return nil, fmt.Errorf("stream version %d not supported", version)
	}
	return &StreamReader{r: r}, nil
}

type StreamReader struct {
	r   io.Reader
	buf [cmdHeaderSize]byte
}

func (r *StreamReader) readCmdHeader() (h cmdHeader, err error) {
	_, err = io.ReadFull(r.r, r.buf[:cmdHeaderSize])
	if err == io.EOF {
		return
	} else if err != nil {
		err = fmt.Errorf("cannot read command header: %v", err)
		return
	}
	err = h.Unmarshal(r.buf[:cmdHeaderSize])
	// TODO: check CRC
	return
}

type SendTLV struct {
	Attr sendCmdAttr
	Val  interface{}
}

func (r *StreamReader) readTLV(rd io.Reader) (*SendTLV, error) {
	_, err := io.ReadFull(rd, r.buf[:tlvHeaderSize])
	if err == io.EOF {
		return nil, err
	} else if err != nil {
		return nil, fmt.Errorf("cannot read tlv header: %v", err)
	}
	var h tlvHeader
	if err = h.Unmarshal(r.buf[:tlvHeaderSize]); err != nil {
		return nil, err
	}
	typ := sendCmdAttr(h.Type)
	if sendCmdAttr(typ) > sendAttrMax { // || th.Len > _BTRFS_SEND_BUF_SIZE {
		return nil, fmt.Errorf("invalid tlv in cmd: %q", typ)
	}
	buf := make([]byte, h.Len)
	_, err = io.ReadFull(rd, buf)
	if err != nil {
		return nil, fmt.Errorf("cannot read tlv: %v", err)
	}
	var v interface{}
	switch typ {
	case sendAttrCtransid, sendAttrCloneCtransid,
		sendAttrUid, sendAttrGid, sendAttrMode,
		sendAttrIno, sendAttrFileOffset, sendAttrSize,
		sendAttrCloneOffset, sendAttrCloneLen:
		if len(buf) != 8 {
			return nil, fmt.Errorf("unexpected int64 size: %v", h.Len)
		}
		v = sendEndianess.Uint64(buf[:8])
	case sendAttrPath, sendAttrPathTo, sendAttrClonePath, sendAttrXattrName:
		v = string(buf)
	case sendAttrData, sendAttrXattrData:
		v = buf
	case sendAttrUuid, sendAttrCloneUuid:
		if h.Len != btrfs.UUIDSize {
			return nil, fmt.Errorf("unexpected UUID size: %v", h.Len)
		}
		var u btrfs.UUID
		copy(u[:], buf)
		v = u
	case sendAttrAtime, sendAttrMtime, sendAttrCtime, sendAttrOtime:
		if h.Len != 12 {
			return nil, fmt.Errorf("unexpected timestamp size: %v", h.Len)
		}
		v = time.Unix( // btrfs_timespec
			int64(sendEndianess.Uint64(buf[:8])),
			int64(sendEndianess.Uint32(buf[8:])),
		)
	default:
		return nil, fmt.Errorf("unsupported tlv type: %v (len: %v)", typ, h.Len)
	}
	return &SendTLV{Attr: typ, Val: v}, nil
}
func (r *StreamReader) ReadCommand() (_ Cmd, gerr error) {
	h, err := r.readCmdHeader()
	if err != nil {
		return nil, err
	}
	var tlvs []SendTLV
	rd := io.LimitReader(r.r, int64(h.Len))
	defer io.Copy(ioutil.Discard, rd)
	for {
		tlv, err := r.readTLV(rd)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("command %v: %v", h.Cmd, err)
		}
		tlvs = append(tlvs, *tlv)
	}
	var c Cmd
	switch h.Cmd {
	case sendCmdEnd:
		c = &StreamEnd{}
	case sendCmdSubvol:
		c = &SubvolCmd{}
	case sendCmdSnapshot:
		c = &SnapshotCmd{}
	case sendCmdChown:
		c = &ChownCmd{}
	case sendCmdChmod:
		c = &ChmodCmd{}
	case sendCmdUtimes:
		c = &UTimesCmd{}
	case sendCmdMkdir:
		c = &MkdirCmd{}
	case sendCmdRename:
		c = &RenameCmd{}
	case sendCmdMkfile:
		c = &MkfileCmd{}
	case sendCmdWrite:
		c = &WriteCmd{}
	case sendCmdTruncate:
		c = &TruncateCmd{}
	}
	if c == nil {
		return &UnknownSendCmd{Kind: h.Cmd, Params: tlvs}, nil
	}
	if err := c.decode(tlvs); err != nil {
		return nil, err
	}
	return c, nil
}

type errUnexpectedAttrType struct {
	Cmd CmdType
	Val SendTLV
}

func (e errUnexpectedAttrType) Error() string {
	return fmt.Sprintf("unexpected type for %q (in %q): %T",
		e.Val.Attr, e.Cmd, e.Val.Val)
}

type errUnexpectedAttr struct {
	Cmd CmdType
	Val SendTLV
}

func (e errUnexpectedAttr) Error() string {
	return fmt.Sprintf("unexpected attr %q for %q (%T)",
		e.Val.Attr, e.Cmd, e.Val.Val)
}

type Cmd interface {
	Type() CmdType
	decode(tlvs []SendTLV) error
}

type UnknownSendCmd struct {
	Kind   CmdType
	Params []SendTLV
}

func (c UnknownSendCmd) Type() CmdType {
	return c.Kind
}
func (c *UnknownSendCmd) decode(tlvs []SendTLV) error {
	c.Params = tlvs
	return nil
}

type StreamEnd struct{}

func (c StreamEnd) Type() CmdType {
	return sendCmdEnd
}
func (c *StreamEnd) decode(tlvs []SendTLV) error {
	if len(tlvs) != 0 {
		return fmt.Errorf("unexpected TLVs for stream end command: %#v", tlvs)
	}
	return nil
}

type SubvolCmd struct {
	Path     string
	UUID     btrfs.UUID
	CTransID uint64
}

func (c SubvolCmd) Type() CmdType {
	return sendCmdSubvol
}
func (c *SubvolCmd) decode(tlvs []SendTLV) error {
	for _, tlv := range tlvs {
		var ok bool
		switch tlv.Attr {
		case sendAttrPath:
			c.Path, ok = tlv.Val.(string)
		case sendAttrUuid:
			c.UUID, ok = tlv.Val.(btrfs.UUID)
		case sendAttrCtransid:
			c.CTransID, ok = tlv.Val.(uint64)
		default:
			return errUnexpectedAttr{Val: tlv, Cmd: c.Type()}
		}
		if !ok {
			return errUnexpectedAttrType{Val: tlv, Cmd: c.Type()}
		}
	}
	return nil
}

type SnapshotCmd struct {
	Path         string
	UUID         btrfs.UUID
	CTransID     uint64
	CloneUUID    btrfs.UUID
	CloneTransID uint64
}

func (c SnapshotCmd) Type() CmdType {
	return sendCmdSnapshot
}
func (c *SnapshotCmd) decode(tlvs []SendTLV) error {
	for _, tlv := range tlvs {
		var ok bool
		switch tlv.Attr {
		case sendAttrPath:
			c.Path, ok = tlv.Val.(string)
		case sendAttrUuid:
			c.UUID, ok = tlv.Val.(btrfs.UUID)
		case sendAttrCtransid:
			c.CTransID, ok = tlv.Val.(uint64)
		case sendAttrCloneUuid:
			c.CloneUUID, ok = tlv.Val.(btrfs.UUID)
		case sendAttrCloneCtransid:
			c.CloneTransID, ok = tlv.Val.(uint64)
		default:
			return errUnexpectedAttr{Val: tlv, Cmd: c.Type()}
		}
		if !ok {
			return errUnexpectedAttrType{Val: tlv, Cmd: c.Type()}
		}
	}
	return nil
}

type ChownCmd struct {
	Path     string
	UID, GID uint64
}

func (c ChownCmd) Type() CmdType {
	return sendCmdChown
}
func (c *ChownCmd) decode(tlvs []SendTLV) error {
	for _, tlv := range tlvs {
		var ok bool
		switch tlv.Attr {
		case sendAttrPath:
			c.Path, ok = tlv.Val.(string)
		case sendAttrUid:
			c.UID, ok = tlv.Val.(uint64)
		case sendAttrGid:
			c.GID, ok = tlv.Val.(uint64)
		default:
			return errUnexpectedAttr{Val: tlv, Cmd: c.Type()}
		}
		if !ok {
			return errUnexpectedAttrType{Val: tlv, Cmd: c.Type()}
		}
	}
	return nil
}

type ChmodCmd struct {
	Path string
	Mode uint64
}

func (c ChmodCmd) Type() CmdType {
	return sendCmdChmod
}
func (c *ChmodCmd) decode(tlvs []SendTLV) error {
	for _, tlv := range tlvs {
		var ok bool
		switch tlv.Attr {
		case sendAttrPath:
			c.Path, ok = tlv.Val.(string)
		case sendAttrMode:
			c.Mode, ok = tlv.Val.(uint64)
		default:
			return errUnexpectedAttr{Val: tlv, Cmd: c.Type()}
		}
		if !ok {
			return errUnexpectedAttrType{Val: tlv, Cmd: c.Type()}
		}
	}
	return nil
}

type UTimesCmd struct {
	Path                string
	ATime, MTime, CTime time.Time
}

func (c UTimesCmd) Type() CmdType {
	return sendCmdUtimes
}
func (c *UTimesCmd) decode(tlvs []SendTLV) error {
	for _, tlv := range tlvs {
		var ok bool
		switch tlv.Attr {
		case sendAttrPath:
			c.Path, ok = tlv.Val.(string)
		case sendAttrAtime:
			c.ATime, ok = tlv.Val.(time.Time)
		case sendAttrMtime:
			c.MTime, ok = tlv.Val.(time.Time)
		case sendAttrCtime:
			c.CTime, ok = tlv.Val.(time.Time)
		default:
			return errUnexpectedAttr{Val: tlv, Cmd: c.Type()}
		}
		if !ok {
			return errUnexpectedAttrType{Val: tlv, Cmd: c.Type()}
		}
	}
	return nil
}

type MkdirCmd struct {
	Path string
	Ino  uint64
}

func (c MkdirCmd) Type() CmdType {
	return sendCmdMkdir
}
func (c *MkdirCmd) decode(tlvs []SendTLV) error {
	for _, tlv := range tlvs {
		var ok bool
		switch tlv.Attr {
		case sendAttrPath:
			c.Path, ok = tlv.Val.(string)
		case sendAttrIno:
			c.Ino, ok = tlv.Val.(uint64)
		default:
			return errUnexpectedAttr{Val: tlv, Cmd: c.Type()}
		}
		if !ok {
			return errUnexpectedAttrType{Val: tlv, Cmd: c.Type()}
		}
	}
	return nil
}

type RenameCmd struct {
	From, To string
}

func (c RenameCmd) Type() CmdType {
	return sendCmdRename
}
func (c *RenameCmd) decode(tlvs []SendTLV) error {
	for _, tlv := range tlvs {
		var ok bool
		switch tlv.Attr {
		case sendAttrPath:
			c.From, ok = tlv.Val.(string)
		case sendAttrPathTo:
			c.To, ok = tlv.Val.(string)
		default:
			return errUnexpectedAttr{Val: tlv, Cmd: c.Type()}
		}
		if !ok {
			return errUnexpectedAttrType{Val: tlv, Cmd: c.Type()}
		}
	}
	return nil
}

type MkfileCmd struct {
	Path string
	Ino  uint64
}

func (c MkfileCmd) Type() CmdType {
	return sendCmdMkfile
}
func (c *MkfileCmd) decode(tlvs []SendTLV) error {
	for _, tlv := range tlvs {
		var ok bool
		switch tlv.Attr {
		case sendAttrPath:
			c.Path, ok = tlv.Val.(string)
		case sendAttrIno:
			c.Ino, ok = tlv.Val.(uint64)
		default:
			return errUnexpectedAttr{Val: tlv, Cmd: c.Type()}
		}
		if !ok {
			return errUnexpectedAttrType{Val: tlv, Cmd: c.Type()}
		}
	}
	return nil
}

type WriteCmd struct {
	Path string
	Off  uint64
	Data []byte
}

func (c WriteCmd) Type() CmdType {
	return sendCmdWrite
}
func (c *WriteCmd) decode(tlvs []SendTLV) error {
	for _, tlv := range tlvs {
		var ok bool
		switch tlv.Attr {
		case sendAttrPath:
			c.Path, ok = tlv.Val.(string)
		case sendAttrFileOffset:
			c.Off, ok = tlv.Val.(uint64)
		case sendAttrData:
			c.Data, ok = tlv.Val.([]byte)
		default:
			return errUnexpectedAttr{Val: tlv, Cmd: c.Type()}
		}
		if !ok {
			return errUnexpectedAttrType{Val: tlv, Cmd: c.Type()}
		}
	}
	return nil
}

type TruncateCmd struct {
	Path string
	Size uint64
}

func (c TruncateCmd) Type() CmdType {
	return sendCmdTruncate
}
func (c *TruncateCmd) decode(tlvs []SendTLV) error {
	for _, tlv := range tlvs {
		var ok bool
		switch tlv.Attr {
		case sendAttrPath:
			c.Path, ok = tlv.Val.(string)
		case sendAttrSize:
			c.Size, ok = tlv.Val.(uint64)
		default:
			return errUnexpectedAttr{Val: tlv, Cmd: c.Type()}
		}
		if !ok {
			return errUnexpectedAttrType{Val: tlv, Cmd: c.Type()}
		}
	}
	return nil
}
