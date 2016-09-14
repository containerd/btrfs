package btrfs

// This header contains the structure definitions and constants used
// by file system objects that can be retrieved using
// the BTRFS_IOC_SEARCH_TREE ioctl.  That means basically anything that
// is needed to describe a leaf node's key or item contents.

const (
	// Holds pointers to all of the tree roots
	BTRFS_ROOT_TREE_OBJECTID = 1

	// Stores information about which extents are in use, and reference counts
	BTRFS_EXTENT_TREE_OBJECTID = 2

	// Chunk tree stores translations from logical -> physical block numbering
	// the super block points to the chunk tree.
	BTRFS_CHUNK_TREE_OBJECTID = 3

	// All files have objectids in this range.
	BTRFS_FIRST_FREE_OBJECTID       = 256
	BTRFS_LAST_FREE_OBJECTID        = 0xffffff00 // -256
	BTRFS_FIRST_CHUNK_TREE_OBJECTID = 256
)
