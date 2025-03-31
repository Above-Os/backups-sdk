package backupssdk

import (
	"path"

	"bytetrade.io/web3os/backups-sdk/cmd/backup"
	"bytetrade.io/web3os/backups-sdk/cmd/download"
	"bytetrade.io/web3os/backups-sdk/cmd/region"
	"bytetrade.io/web3os/backups-sdk/cmd/restore"
	"bytetrade.io/web3os/backups-sdk/cmd/snapshots"
	"bytetrade.io/web3os/backups-sdk/pkg/constants"
	"bytetrade.io/web3os/backups-sdk/pkg/logger"
	"bytetrade.io/web3os/backups-sdk/pkg/storage"
	"bytetrade.io/web3os/backups-sdk/pkg/utils"
	"github.com/spf13/cobra"
)

func NewBackupCommands() *cobra.Command {
	var homeDir = utils.GetHomeDir()
	var jsonLogDir = path.Join(homeDir, constants.DefaultBaseDir, constants.DefaultLogsDir)

	logger.InitLogger(jsonLogDir, true)

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
	cmds.AddCommand(region.NewCmdRegions())

	return cmds
}

func NewBackupService(option *storage.BackupOption) *storage.BackupService {
	logger.SetLogger(option.Logger)

	return storage.NewBackupService(option)
}

func NewRestoreService(option *storage.RestoreOption) *storage.RestoreService {
	logger.SetLogger(option.Logger)

	return storage.NewRestoreService(option)
}

func NewRegionService(option *storage.RegionOption) *storage.RegionService {
	logger.SetLogger(option.Logger)

	return storage.NewRegionService(option)
}
