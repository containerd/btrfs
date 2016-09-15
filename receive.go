package btrfs

import (
	"bytes"
	"errors"
	"io"
	"os/exec"
)

func Receive(r io.Reader, mount string) error {
	// TODO: write a native implementation?
	//tf, err := ioutil.TempFile("","btrfs_snap")
	//if err != nil {
	//	return err
	//}
	//defer func(){
	//	name := tf.Name()
	//	tf.Close()
	//	os.Remove(name)
	//}()
	buf := bytes.NewBuffer(nil)
	cmd := exec.Command("btrfs", "receive", mount)
	cmd.Stdin = r
	cmd.Stderr = buf
	if err := cmd.Run(); err != nil {
		if buf.Len() != 0 {
			return errors.New(buf.String())
		}
		return err
	}
	return nil
}
