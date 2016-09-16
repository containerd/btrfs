package main

import (
	"fmt"
	"github.com/dennwc/btrfs"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	RootCmd.AddCommand(
		SubvolumeCmd,
		SendCmd,
		ReceiveCmd,
	)

	SubvolumeCmd.AddCommand(
		SubvolumeCreateCmd,
		SubvolumeDeleteCmd,
		SubvolumeListCmd,
	)
}

var RootCmd = &cobra.Command{
	Use:   "btrfs [--help] [--version] <group> [<group>...] <command> [<args>]",
	Short: "Use --help as an argument for information on a specific group or command.",
}

var SubvolumeCmd = &cobra.Command{
	Use: "subvolume <command> <args>",
}

var SubvolumeCreateCmd = &cobra.Command{
	Use:   "create [-i <qgroupid>] [<dest>/]<name>",
	Short: "Create a subvolume",
	Long:  `Create a subvolume <name> in <dest>.  If <dest> is not given subvolume <name> will be created in the current directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("subvolume not specified")
		} else if len(args) > 1 {
			return fmt.Errorf("only one subvolume name is allowed")
		}
		return btrfs.CreateSubVolume(args[0])
	},
}

var SubvolumeDeleteCmd = &cobra.Command{
	Use:   "delete [options] <subvolume> [<subvolume>...]",
	Short: "Delete subvolume(s)",
	Long: `Delete subvolumes from the filesystem. The corresponding directory
is removed instantly but the data blocks are removed later.
The deletion does not involve full commit by default due to
performance reasons (as a consequence, the subvolume may appear again
after a crash). Use one of the --commit options to wait until the
operation is safely stored on the media.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, arg := range args {
			if err := btrfs.DeleteSubVolume(arg); err != nil {
				return err
			}
		}
		return nil
	},
}

var SubvolumeListCmd = &cobra.Command{
	Use:   "list <mount>",
	Short: "List subvolumes",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("expected one destination argument")
		}
		list, err := btrfs.ListSubVolumes(args[0])
		if err == nil {
			for _, v := range list {
				fmt.Printf("%+v\n", v)
			}
		}
		return err
	},
}

var SendCmd = &cobra.Command{
	Use:   "send [-ve] [-p <parent>] [-c <clone-src>] [-f <outfile>] <subvol> [<subvol>...]",
	Short: "Send the subvolume(s) to stdout.",
	Long: `Sends the subvolume(s) specified by <subvol> to stdout.
<subvol> should be read-only here.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return btrfs.Send(os.Stdout, "", args...)
	},
}

var ReceiveCmd = &cobra.Command{
	Use:   "receive [-ve] [-f <infile>] [--max-errors <N>] <mount>",
	Short: "Receive subvolumes from stdin.",
	Long: `Receives one or more subvolumes that were previously
sent with btrfs send. The received subvolumes are stored
into <mount>.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("expected one destination argument")
		}
		return btrfs.Receive(os.Stdin, args[0])
	},
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
