package btrfs

//go:generate sh -c "go tool cgo -godefs ctypes_linux.go > types_linux.go"
//go:generate perl -i -pne s|Fd(\s+)int64|Fd\1uintptr|g types_linux.go
//go:generate perl -i -pne s|int8|byte|g types_linux.go
//go:generate gofmt -s -w types_linux.go
