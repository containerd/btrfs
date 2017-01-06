package btrfs

import (
	"github.com/dennwc/btrfs/test"
	"io"
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

func TestCloneFile(t *testing.T) {
	dir, closer := btrfstest.New(t, sizeDef)
	defer closer()

	f1, err := os.Create(filepath.Join(dir, "1.dat"))
	if err != nil {
		t.Fatal(err)
	}
	defer f1.Close()

	const data = "btrfs_test"
	_, err = f1.WriteString(data)
	if err != nil {
		t.Fatal(err)
	}

	f2, err := os.Create(filepath.Join(dir, "2.dat"))
	if err != nil {
		t.Fatal(err)
	}
	defer f2.Close()

	err = CloneFile(f2, f1)
	if err != nil {
		t.Fatal(err)
	}

	buf := make([]byte, len(data))
	n, err := f2.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	buf = buf[:n]
	if string(buf) != data {
		t.Fatalf("wrong data returned: %q", string(buf))
	}
}
