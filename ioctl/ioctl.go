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
