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
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	btrfstest "github.com/containerd/btrfs/v2/test"
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

func TestSubvolumes(t *testing.T) {
	dir, closer := btrfstest.New(t, sizeDef)
	defer closer()
	fs, err := Open(dir, false)
	if err != nil {
		t.Fatal(err)
	}
	defer fs.Close()

	mksub := func(in string, path string) {
		if in != "" {
			path = filepath.Join(dir, in, path)
		} else {
			path = filepath.Join(dir, path)
		}
		if err := CreateSubVolume(path); err != nil {
			t.Fatalf("cannot create subvolume %v: %v", path, err)
		}
	}
	delsub := func(path string) {
		path = filepath.Join(dir, path)
		if err := DeleteSubVolume(path); err != nil {
			t.Fatalf("cannot delete subvolume %v: %v", path, err)
		}
	}
	expect := func(exp []string) {
		subs, err := fs.ListSubvolumes(nil)
		if err != nil {
			t.Fatal(err)
		}
		var got []string
		for _, s := range subs {
			if s.UUID.IsZero() {
				t.Fatalf("zero uuid in %+v", s)
			}
			if s.Path != "" {
				got = append(got, s.Path)
			}
		}
		sort.Strings(got)
		sort.Strings(exp)
		if !reflect.DeepEqual(got, exp) {
			t.Fatalf("list failed:\ngot: %v\nvs\nexp: %v", got, exp)
		}
	}

	names := []string{"foo", "bar", "baz"}
	for _, name := range names {
		mksub("", name)
	}
	for _, name := range names {
		mksub(names[0], name)
	}
	expect([]string{
		"foo", "bar", "baz",
		"foo/foo", "foo/bar", "foo/baz",
	})
	delsub("foo/bar")
	expect([]string{
		"foo", "bar", "baz",
		"foo/foo", "foo/baz",
	})

	path := filepath.Join(names[0], names[2])
	mksub(path, "new")
	path = filepath.Join(path, "new")

	id, err := getPathRootID(filepath.Join(dir, path))
	if err != nil {
		t.Fatal(err)
	}
	info, err := subvolSearchByRootID(fs.f, id, "")
	if err != nil {
		t.Fatal(err)
	} else if info.Path != path {
		t.Fatalf("wrong path returned: %v vs %v", info.Path, path)
	}
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

func TestResize(t *testing.T) {
	dir, err := ioutil.TempDir("", "btrfs_data_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	fname := filepath.Join(dir, "data")
	if err = btrfstest.Mkfs(fname, sizeDef); err != nil {
		t.Fatal(err)
	}
	mnt := filepath.Join(dir, "mnt")
	if err = os.MkdirAll(mnt, 0755); err != nil {
		t.Fatal(err)
	}
	if err = btrfstest.Mount(mnt, fname); err != nil {
		t.Fatal(err)
	}
	defer btrfstest.Unmount(mnt)

	fs, err := Open(mnt, false)
	if err != nil {
		t.Fatal(err)
	}
	st, err := fs.Usage()
	fs.Close()
	if err != nil {
		t.Fatal(err)
	}

	if err = btrfstest.Unmount(mnt); err != nil {
		t.Fatal(err)
	}
	var newSize int64 = sizeDef
	newSize = int64(float64(newSize) * 1.1)
	if err = os.Truncate(fname, newSize); err != nil {
		t.Fatal(err)
	}
	if err = btrfstest.Mount(mnt, fname); err != nil {
		t.Fatal(err)
	}

	fs, err = Open(mnt, false)
	if err != nil {
		t.Fatal(err)
	}
	defer fs.Close()

	if err = fs.ResizeToMax(); err != nil {
		t.Fatal(err)
	}

	st2, err := fs.Usage()
	if err != nil {
		t.Fatal(err)
	} else if st.Total >= st2.Total {
		t.Fatal("to resized:", st.Total, st2.Total)
	}
}
