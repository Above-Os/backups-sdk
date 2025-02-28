package backup

import (
	"bytetrade.io/web3os/backups-sdk/cmd/options"
	"bytetrade.io/web3os/backups-sdk/pkg/backup"
	"github.com/spf13/cobra"
)

func NewCmdBackup() *cobra.Command {
	rootBackupCmds := &cobra.Command{
		Use:               "backup",
		Short:             "Olares Backup Tool",
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	}

	rootBackupCmds.AddCommand(NewCmdSpace())
	rootBackupCmds.AddCommand(NewCmdS3())
	rootBackupCmds.AddCommand(NewCmdCos())
	rootBackupCmds.AddCommand(NewCmdFs())

	return rootBackupCmds
}

func NewCmdSpace() *cobra.Command {
	o := options.NewBackupSpaceOption()
	cmd := &cobra.Command{
		Use:   "space",
		Short: "Backup files to Space",
		Run: func(cmd *cobra.Command, args []string) {
			var backupService = backup.NewBackupService(o.BaseDir)
			backupService.BackupToSpace(o)
		},
	}
	o.AddFlags(cmd)
	return cmd
}

func NewCmdS3() *cobra.Command {
	o := options.NewBackupS3Option()
	cmd := &cobra.Command{
		Use:   "s3",
		Short: "Backup files to S3",
		Run: func(cmd *cobra.Command, args []string) {
			var backupService = backup.NewBackupService(o.BaseDir)
			backupService.BackupToS3(o)
		},
	}
	o.AddFlags(cmd)
	return cmd
}

func NewCmdCos() *cobra.Command {
	o := options.NewBackupCosOption()
	cmd := &cobra.Command{
		Use:   "cos",
		Short: "Backup files to Tencent COS",
		Run: func(cmd *cobra.Command, args []string) {
			var backupService = backup.NewBackupService(o.BaseDir)
			backupService.BackupToCos(o)
		},
	}
	o.AddFlags(cmd)
	return cmd
}

func NewCmdFs() *cobra.Command {
	o := options.NewBackupFilesystemOption()
	cmd := &cobra.Command{
		Use:   "fs",
		Short: "Backup files to FileSystem",
		Run: func(cmd *cobra.Command, args []string) {
			var backupService = backup.NewBackupService(o.BaseDir)
			backupService.BackupToFilesystem(o)
		},
	}
	o.AddFlags(cmd)
	return cmd
}
