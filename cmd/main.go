package main

import (
	"fmt"
	"os"
	"path"

	"bytetrade.io/web3os/backups-sdk/cmd/backup"
	"bytetrade.io/web3os/backups-sdk/cmd/region"
	"bytetrade.io/web3os/backups-sdk/cmd/restore"
	"bytetrade.io/web3os/backups-sdk/cmd/snapshots"
	"bytetrade.io/web3os/backups-sdk/pkg/constants"
	"bytetrade.io/web3os/backups-sdk/pkg/logger"
	"bytetrade.io/web3os/backups-sdk/pkg/utils"
	"github.com/spf13/cobra"
)

func main() {
	var homeDir = utils.GetHomeDir()
	var jsonLogDir = path.Join(homeDir, constants.DefaultBaseDir, constants.DefaultLogsDir)

	logger.InitLogger(jsonLogDir, true)

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

	if err := cmds.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
