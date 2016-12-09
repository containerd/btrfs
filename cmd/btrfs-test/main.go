package main

import (
	"flag"
	"log"
	"os"

	"github.com/stevvooe/go-btrfs"
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
	default:
		log.Fatal("unknown command", os.Args[1])
	}
}
