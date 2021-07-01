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

// Package mtab contains tools to work with /etc/mtab file.
package mtab

import (
	"bufio"
	"io"
	"os"
	"strings"
)

type MountPoint struct {
	Dev   string
	Mount string
	Type  string
	Opts  string
}

// Mounts returns a list of mount point from /etc/mtab.
func Mounts() ([]MountPoint, error) {
	file, err := os.Open("/etc/mtab")
	if err != nil {
		return nil, err
	}
	defer file.Close()
	r := bufio.NewReader(file)
	var out []MountPoint
	for {
		line, err := r.ReadString('\n')
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		fields := strings.Fields(line)
		out = append(out, MountPoint{
			Dev:   fields[0],
			Mount: fields[1],
			Type:  fields[2],
			Opts:  fields[3],
		})
	}
	return out, nil
}
