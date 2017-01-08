package btrfs

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

const nativeReceive = false

func Receive(r io.Reader, dstDir string) error {
	if !nativeReceive {
		buf := bytes.NewBuffer(nil)
		cmd := exec.Command("btrfs", "receive", dstDir)
		cmd.Stdin = r
		cmd.Stderr = buf
		if err := cmd.Run(); err != nil {
			if buf.Len() != 0 {
				return errors.New(buf.String())
			}
			return err
		}
		return nil
	}
	var err error
	dstDir, err = filepath.Abs(dstDir)
	if err != nil {
		return err
	}
	realMnt, err := findMountRoot(dstDir)
	if err != nil {
		return err
	}
	dir, err := os.OpenFile(dstDir, os.O_RDONLY|syscall.O_NOATIME, 0755)
	if err != nil {
		return err
	}
	mnt, err := os.OpenFile(realMnt, os.O_RDONLY|syscall.O_NOATIME, 0755)
	if err != nil {
		return err
	}
	// We want to resolve the path to the subvolume we're sitting in
	// so that we can adjust the paths of any subvols we want to receive in.
	subvolID, err := getFileRootID(mnt)
	if err != nil {
		return err
	}
	sr, err := newStreamReader(r)
	if err != nil {
		return err
	}
	_, _, _ = dir, subvolID, sr
	panic("not implemented")
}

type streamReader struct {
	r    io.Reader
	hbuf []byte
	buf  *bytes.Buffer
}
type sendCommandArgs struct {
	Type tlvType
	Data []byte
}
type sendCommand struct {
	Type sendCmd
	Args []sendCommandArgs
}

func (sr *streamReader) ReadCommand() (*sendCommand, error) {
	sr.buf.Reset()
	var h cmdHeader
	if sr.hbuf == nil {
		sr.hbuf = make([]byte, h.Size())
	}
	if _, err := io.ReadFull(sr.r, sr.hbuf); err != nil {
		return nil, err
	} else if err = h.Unmarshal(sr.hbuf); err != nil {
		return nil, err
	}
	if sr.buf == nil {
		sr.buf = bytes.NewBuffer(nil)
	}
	if _, err := io.CopyN(sr.buf, sr.r, int64(h.Len)); err != nil {
		return nil, err
	}
	tbl := crc32.MakeTable(0)
	crc := crc32.Checksum(sr.buf.Bytes(), tbl)
	if crc != h.Crc {
		return nil, fmt.Errorf("crc missmatch in command: %x vs %x", crc, h.Crc)
	}
	cmd := sendCommand{Type: sendCmd(h.Cmd)}
	var th tlvHeader
	data := sr.buf.Bytes()
	for {
		if n := len(data); n < th.Size() {
			if n != 0 {
				return nil, io.ErrUnexpectedEOF
			}
			break
		}
		if err := th.Unmarshal(data); err != nil {
			return nil, err
		}
		data = data[th.Size():]
		if th.Type > _BTRFS_SEND_A_MAX { // || th.Len > _BTRFS_SEND_BUF_SIZE {
			return nil, fmt.Errorf("invalid tlv in cmd: %+v", th)
		}
		b := make([]byte, th.Len)
		copy(b, data)
		cmd.Args = append(cmd.Args, sendCommandArgs{Type: th.Type, Data: b})
	}
	return &cmd, nil
}

func newStreamReader(r io.Reader) (*streamReader, error) {
	buf := make([]byte, sendStreamMagicSize+4)
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return nil, err
	} else if bytes.Compare(buf[:sendStreamMagicSize], []byte(_BTRFS_SEND_STREAM_MAGIC)) != 0 {
		return nil, errors.New("unexpected stream header")
	}
	version := binary.LittleEndian.Uint32(buf[sendStreamMagicSize:])
	if version > _BTRFS_SEND_STREAM_VERSION {
		return nil, fmt.Errorf("stream version %d not supported", version)
	}
	return &streamReader{r: r}, nil
}
