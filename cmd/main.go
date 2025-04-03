package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"runtime"

	"bytetrade.io/web3os/backups-sdk/cmd/backup"
	"bytetrade.io/web3os/backups-sdk/cmd/download"
	"bytetrade.io/web3os/backups-sdk/cmd/region"
	"bytetrade.io/web3os/backups-sdk/cmd/restore"
	"bytetrade.io/web3os/backups-sdk/cmd/snapshots"
	"bytetrade.io/web3os/backups-sdk/cmd/stats"
	"bytetrade.io/web3os/backups-sdk/pkg/constants"
	"bytetrade.io/web3os/backups-sdk/pkg/logger"
	"bytetrade.io/web3os/backups-sdk/pkg/utils"
	"github.com/spf13/cobra"
)

func main() {
	var homeDir = utils.GetHomeDir()
	var jsonLogDir = path.Join(homeDir, constants.DefaultBaseDir, constants.DefaultLogsDir)

	logger.InitLogger(jsonLogDir, true)

	if runtime.GOOS == "windows" {
		panic(errors.New("Windows system is not currently supported. Please switch to WSL (Windows Subsystem for Linux)."))
	}

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
	cmds.AddCommand(region.NewCmdRegions())
	cmds.AddCommand(download.NewCmdDownload())
	cmds.AddCommand(stats.NewCmdStats())

	if err := cmds.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
