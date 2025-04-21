package main

import (
	"errors"
	"fmt"
	"os"
	"runtime"

	"bytetrade.io/web3os/backups-sdk/cmd/backup"
	"bytetrade.io/web3os/backups-sdk/cmd/download"
	"bytetrade.io/web3os/backups-sdk/cmd/region"
	"bytetrade.io/web3os/backups-sdk/cmd/restore"
	"bytetrade.io/web3os/backups-sdk/cmd/snapshots"
	"bytetrade.io/web3os/backups-sdk/cmd/stats"
	"bytetrade.io/web3os/backups-sdk/pkg/logger"
	"github.com/spf13/cobra"
)

func main() {
	if runtime.GOOS == "windows" {
		panic(errors.New("Windows system is not currently supported. Please switch to WSL (Windows Subsystem for Linux)."))
	}

	cmds := &cobra.Command{
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			logger.InitLogger(true)
		},
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
			DisableNoDescFlag: true,
			HiddenDefaultCmd:  true,
		},
	}
	cmds.AddCommand(backup.NewCmdBackup())
	cmds.AddCommand(restore.NewCmdRestore())
	cmds.AddCommand(snapshots.NewCmdSnapshots())
	cmds.AddCommand(region.NewCmdRegions())
	cmds.AddCommand(download.NewCmdDownload())
	cmds.AddCommand(stats.NewCmdStats())

	if err := cmds.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
