package btrfs

//go:generate go run ./cmd/hgen.go -u -g -t BTRFS_ -p btrfs -o btrfs_tree_hc.go btrfs_tree.h
//go:generate gofmt -l -w btrfs_tree_hc.go
