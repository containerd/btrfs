package btrfs

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
)

func Send(w io.Writer, parent string, subvols ...string) error {
	if len(subvols) == 0 {
		return nil
	}
	// TODO: write a native implementation?
	args := []string{
		"send",
	}
	if parent != "" {
		args = append(args, "-p", parent)
	}
	tf, err := ioutil.TempFile("", "btrfs_snap")
	if err != nil {
		return err
	}
	defer func() {
		name := tf.Name()
		tf.Close()
		os.Remove(name)
	}()
	args = append(args, "-f", tf.Name())
	buf := bytes.NewBuffer(nil)
	args = append(args, subvols...)
	cmd := exec.Command("btrfs", args...)
	cmd.Stderr = buf
	if err = cmd.Run(); err != nil {
		if buf.Len() != 0 {
			return errors.New(buf.String())
		}
		return err
	}
	tf.Seek(0, 0)
	_, err = io.Copy(w, tf)
	return err
}
