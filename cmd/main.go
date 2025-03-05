package main

import (
	"fmt"
	"os"
	"path"

	"bytetrade.io/web3os/backups-sdk/cmd/backup"
	"bytetrade.io/web3os/backups-sdk/cmd/region"
	"bytetrade.io/web3os/backups-sdk/cmd/restore"
	"bytetrade.io/web3os/backups-sdk/cmd/snapshots"
	"bytetrade.io/web3os/backups-sdk/pkg/util"
	"bytetrade.io/web3os/backups-sdk/pkg/util/logger"
	"github.com/spf13/cobra"
)

const (
	defaultBaseDir = ".olares"
)

func main() {
	var homeDir = util.GetHomeDir()
	var jsonLogDir = path.Join(homeDir, defaultBaseDir, "logs")
	logger.InitLog(jsonLogDir, true)

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
