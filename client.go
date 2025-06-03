package backupssdk

import (
	"errors"
	"runtime"

	"github.com/spf13/cobra"
	"olares.com/backups-sdk/cmd/backup"
	"olares.com/backups-sdk/cmd/download"
	"olares.com/backups-sdk/cmd/region"
	"olares.com/backups-sdk/cmd/restore"
	"olares.com/backups-sdk/cmd/snapshots"
	"olares.com/backups-sdk/pkg/logger"
	"olares.com/backups-sdk/pkg/storage"
)

func NewBackupCommands() *cobra.Command {
	cmds := &cobra.Command{
		Use:   "backups",
		Short: "Olares backup tool-kit",
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

func NewStatsService(option *storage.SnapshotsOption) *storage.StatsService {
	logger.SetLogger(option.Logger)

	return storage.NewStatsService(option)
}

func NewSnapshotsService(option *storage.SnapshotsOption) *storage.SnapshotsService {
	logger.SetLogger(option.Logger)

	return storage.NewSnapshotsService(option)
}
