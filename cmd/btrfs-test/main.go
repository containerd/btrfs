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

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/containerd/btrfs/v2"
)

var (
	readonly bool
)

func init() {
	flag.BoolVar(&readonly, "readonly", false, "readonly snapshot")
}

func main() {
	flag.Parse()

	switch os.Args[1] {
	case "create":
		if err := btrfs.SubvolCreate(os.Args[2]); err != nil {
			log.Fatalln(err)
		}
	case "snapshot":
		if err := btrfs.SubvolSnapshot(os.Args[3], os.Args[2], readonly); err != nil {
			log.Fatalln(err)
		}
	case "delete":
		if err := btrfs.SubvolDelete(os.Args[2]); err != nil {
			log.Fatalln(err)
		}
	case "list":
		infos, err := btrfs.SubvolList(os.Args[2])
		if err != nil {
			log.Fatalln(err)
		}
		tw := tabwriter.NewWriter(os.Stdout, 0, 8, 4, '\t', 0)

		fmt.Fprintf(tw, "ID\tParent\tTopLevel\tGen\tOGen\tUUID\tParentUUID\tPath\n")

		for _, subvol := range infos {
			fmt.Fprintf(tw, "%d\t%d\t%d\t%d\t%d\t%s\t%s\t%s\n",
				subvol.ID, subvol.ParentID, subvol.TopLevelID,
				subvol.Generation, subvol.OriginalGeneration, subvol.UUID, subvol.ParentUUID,
				subvol.Path)

		}

		tw.Flush()
	case "show":
		info, err := btrfs.SubvolInfo(os.Args[2])
		if err != nil {
			log.Fatalln(err)
		}

		fmt.Printf("%#v\n", info)
	default:
		log.Fatal("unknown command", os.Args[1])
	}
}
