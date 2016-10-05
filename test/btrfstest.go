package btrfstest

import (
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func run(name string, args ...string) error {
	buf := bytes.NewBuffer(nil)
	cmd := exec.Command(name, args...)
	cmd.Stdout = buf
	cmd.Stderr = buf
	err := cmd.Run()
	if err == nil {
		return nil
	} else if buf.Len() == 0 {
		return err
	}
	return errors.New("error: " + strings.TrimSpace(string(buf.Bytes())))
}

func Mkfs(file string, size int64) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	if err = f.Truncate(size); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = run("mkfs.btrfs", file); err != nil {
		os.Remove(file)
		return err
	}
	return err
}

func Mount(mount string, file string) error {
	if err := run("mount", file, mount); err != nil {
		return err
	}
	return nil
}

func New(t testing.TB, size int64) (string, func()) {
	f, err := ioutil.TempFile("", "btrfs_vol")
	if err != nil {
		t.Fatal(err)
	}
	name := f.Name()
	f.Close()
	rm := func() {
		os.Remove(name)
	}
	if err = Mkfs(name, size); err != nil {
		rm()
	}
	mount, err := ioutil.TempDir("", "btrfs_mount")
	if err != nil {
		rm()
		t.Fatal(err)
	}
	if err = Mount(mount, name); err != nil {
		rm()
		os.RemoveAll(mount)
		if txt := err.Error(); strings.Contains(txt, "permission denied") ||
			strings.Contains(txt, "only root") {
			t.Skip(err)
		} else {
			t.Fatal(err)
		}
	}
	done := false
	return mount, func() {
		if done {
			return
		}
		for i := 0; i < 5; i++ {
			if err := run("umount", mount); err == nil {
				break
			} else {
				log.Println("umount failed:", err)
				if strings.Contains(err.Error(), "busy") {
					time.Sleep(time.Second)
				} else {
					break
				}
			}
		}
		if err := os.Remove(mount); err != nil {
			log.Println("cleanup failed:", err)
		}
		rm()
		done = true
	}
}
