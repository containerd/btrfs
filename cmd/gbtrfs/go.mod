module github.com/containerd/btrfs/v2/cmd/gbtrfs

go 1.16

replace github.com/containerd/btrfs/v2 => ../..

require (
	github.com/containerd/btrfs/v2 v2.0.0-00010101000000-000000000000
	github.com/spf13/cobra v1.1.3
)
