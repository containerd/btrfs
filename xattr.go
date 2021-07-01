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

package btrfs

import (
	"bytes"
	"os"
	"syscall"
)

const (
	xattrPrefix      = "btrfs."
	xattrCompression = xattrPrefix + "compression"
)

type Compression string

const (
	CompressionNone = Compression("")
	LZO             = Compression("lzo")
	ZLIB            = Compression("zlib")
)

func SetCompression(path string, v Compression) error {
	var value []byte
	if v != CompressionNone {
		var err error
		value, err = syscall.ByteSliceFromString(string(v))
		if err != nil {
			return err
		}
	}
	err := syscall.Setxattr(path, xattrCompression, value, 0)
	if err != nil {
		return &os.PathError{Op: "setxattr", Path: path, Err: err}
	}
	return nil
}

func GetCompression(path string) (Compression, error) {
	var buf []byte
	for {
		sz, err := syscall.Getxattr(path, xattrCompression, nil)
		if err == syscall.ENODATA || sz == 0 {
			return CompressionNone, nil
		} else if err != nil {
			return CompressionNone, &os.PathError{Op: "getxattr", Path: path, Err: err}
		}
		if cap(buf) < sz {
			buf = make([]byte, sz)
		} else {
			buf = buf[:sz]
		}
		sz, err = syscall.Getxattr(path, xattrCompression, buf)
		if err == syscall.ENODATA {
			return CompressionNone, nil
		} else if err == syscall.ERANGE {
			// xattr changed by someone else, and is larger than our current buffer
			continue
		} else if err != nil {
			return CompressionNone, &os.PathError{Op: "getxattr", Path: path, Err: err}
		}
		buf = buf[:sz]
		break
	}
	buf = bytes.TrimSuffix(buf, []byte{0})
	return Compression(buf), nil
}
