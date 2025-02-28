package backupssdk

import (
	"fmt"
	"os"

	"bytetrade.io/web3os/backups-sdk/cmd/backup"
	"bytetrade.io/web3os/backups-sdk/cmd/restore"
	"bytetrade.io/web3os/backups-sdk/cmd/snapshots"
	"bytetrade.io/web3os/backups-sdk/pkg/util"
	"github.com/spf13/cobra"
)

func init() {
	_, err := util.GetCommand("restic")
	if err != nil {
		// todo
		fmt.Println("restic not found")
		os.Exit(1)
	}
}

func NewBackupCommands() *cobra.Command {
	cmds := &cobra.Command{
		Use:   "backups",
		Short: "Olares backup tool-kit",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
			DisableNoDescFlag: true,
			HiddenDefaultCmd:  true,
		},
	}
	cmds.AddCommand(backup.NewCmdBackup())
	cmds.AddCommand(restore.NewCmdRestore())
	cmds.AddCommand(snapshots.NewCmdSnapshots())

	return cmds
}
