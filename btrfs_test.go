package btrfs

import (
	"github.com/dennwc/btrfs/test"
	"os"
	"path/filepath"
	"testing"
)

const sizeDef = 256 * 1024 * 1024

func TestOpen(t *testing.T) {
	dir, closer := btrfstest.New(t, sizeDef)
	defer closer()
	fs, err := Open(dir, true)
	if err != nil {
		t.Fatal(err)
	}
	if err = fs.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestIsSubvolume(t *testing.T) {
	dir, closer := btrfstest.New(t, sizeDef)
	defer closer()

	isSubvol := func(path string, expect bool) {
		ok, err := IsSubVolume(path)
		if err != nil {
			t.Errorf("failed to check subvolume %v: %v", path, err)
			return
		} else if ok != expect {
			t.Errorf("unexpected result for %v", path)
		}
	}
	mkdir := func(path string) {
		path = filepath.Join(dir, path)
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatalf("cannot create dir %v: %v", path, err)
		}
		isSubvol(path, false)
	}

	mksub := func(path string) {
		path = filepath.Join(dir, path)
		if err := CreateSubVolume(path); err != nil {
			t.Fatalf("cannot create subvolume %v: %v", path, err)
		}
		isSubvol(path, true)
	}

	mksub("v1")

	mkdir("v1/d2")
	mksub("v1/v2")

	mkdir("v1/d2/d3")
	mksub("v1/d2/v3")

	mkdir("v1/v2/d3")
	mksub("v1/v2/v3")

	mkdir("d1")

	mkdir("d1/d2")
	mksub("d1/v2")

	mkdir("d1/d2/d3")
	mksub("d1/d2/v3")

	mkdir("d1/v2/d3")
	mksub("d1/v2/v3")
}

func TestCompression(t *testing.T) {
	dir, closer := btrfstest.New(t, sizeDef)
	defer closer()
	fs, err := Open(dir, true)
	if err != nil {
		t.Fatal(err)
	}
	defer fs.Close()
	if err := fs.CreateSubVolume("sub"); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "sub")

	if err := SetCompression(path, LZO); err != nil {
		t.Fatal(err)
	}
	if c, err := GetCompression(path); err != nil {
		t.Fatal(err)
	} else if c != LZO {
		t.Fatalf("unexpected compression returned: %q", string(c))
	}
}
