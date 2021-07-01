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

package ioctl

import (
	"github.com/dennwc/ioctl"
	"os"
)

const (
	None  = ioctl.None
	Write = ioctl.Write
	Read  = ioctl.Read
)

// IOC
//
// Deprecated: use github/dennwc/ioctl
func IOC(dir, typ, nr, size uintptr) uintptr {
	return ioctl.IOC(dir, typ, nr, size)
}

// IO
//
// Deprecated: use github/dennwc/ioctl
func IO(typ, nr uintptr) uintptr {
	return ioctl.IO(typ, nr)
}

// IOC
//
// Deprecated: use github/dennwc/ioctl
func IOR(typ, nr, size uintptr) uintptr {
	return ioctl.IOR(typ, nr, size)
}

// IOW
//
// Deprecated: use github/dennwc/ioctl
func IOW(typ, nr, size uintptr) uintptr {
	return ioctl.IOW(typ, nr, size)
}

// IOWR
//
// Deprecated: use github/dennwc/ioctl
func IOWR(typ, nr, size uintptr) uintptr {
	return ioctl.IOWR(typ, nr, size)
}

// Ioctl
//
// Deprecated: use github/dennwc/ioctl
func Ioctl(f *os.File, ioc uintptr, addr uintptr) error {
	return ioctl.Ioctl(f, ioc, addr)
}

// Do
//
// Deprecated: use github/dennwc/ioctl
func Do(f *os.File, ioc uintptr, arg interface{}) error {
	return ioctl.Do(f, ioc, arg)
}
