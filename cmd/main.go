package main

import (
	"os"

	"bytetrade.io/web3os/backups-sdk/cmd/backup"
	"bytetrade.io/web3os/backups-sdk/cmd/download"
	"bytetrade.io/web3os/backups-sdk/cmd/restore"
	"bytetrade.io/web3os/backups-sdk/cmd/snapshots"
	"github.com/spf13/cobra"
)

func init() {
	// _, err := util.GetCommand("restic")
	// if err != nil {
	// 	// todo
	// 	fmt.Println("restic not found, execute with ' download' to download restic and install on Linux")
	// 	os.Exit(1)
	// }
}

func main() {
	cmds := &cobra.Command{
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
			DisableNoDescFlag: true,
			HiddenDefaultCmd:  true,
		},
	}
	cmds.AddCommand(backup.NewCmdBackup())
	cmds.AddCommand(restore.NewCmdRestore())
	cmds.AddCommand(snapshots.NewCmdSnapshots())
	cmds.AddCommand(download.NewCmdDownload())

	if err := cmds.Execute(); err != nil {
		// fmt.Println(err)
		os.Exit(1)
	}
}
