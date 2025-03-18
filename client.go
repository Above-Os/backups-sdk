package backupssdk

import (
	"bytetrade.io/web3os/backups-sdk/cmd/backup"
	"bytetrade.io/web3os/backups-sdk/cmd/download"
	"bytetrade.io/web3os/backups-sdk/cmd/restore"
	"bytetrade.io/web3os/backups-sdk/cmd/snapshots"
	"bytetrade.io/web3os/backups-sdk/pkg/logger"
	"bytetrade.io/web3os/backups-sdk/pkg/storage"
	"github.com/spf13/cobra"
)

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
	cmds.AddCommand(download.NewCmdDownload())

	return cmds
}

func NewBackupService(option *storage.BackupOption) *storage.BackupService {
	logger.SetLogger(option.Logger)

	return storage.NewBackupService(option)
}
