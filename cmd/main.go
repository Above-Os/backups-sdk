package main

import (
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"olares.com/backups-sdk/cmd/backup"
	"olares.com/backups-sdk/cmd/download"
	"olares.com/backups-sdk/cmd/region"
	"olares.com/backups-sdk/cmd/restore"
	"olares.com/backups-sdk/cmd/snapshots"
	"olares.com/backups-sdk/cmd/stats"
	"olares.com/backups-sdk/pkg/logger"
)

func main() {
	cmds := &cobra.Command{
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if runtime.GOOS == "windows" {
				panic(errors.New("Windows system is not currently supported. Please switch to WSL (Windows Subsystem for Linux)."))
			}

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
