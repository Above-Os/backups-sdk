package backup

import (
	"bytetrade.io/web3os/backups-sdk/pkg/constants"
	"bytetrade.io/web3os/backups-sdk/pkg/options"
	"bytetrade.io/web3os/backups-sdk/pkg/storage"
	"github.com/spf13/cobra"
)

func NewCmdBackup() *cobra.Command {
	rootBackupCmds := &cobra.Command{
		Use:               "backup",
		Short:             "Back up data to multiple storage targets: Space, S3, COS, and local",
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
		Short: "Backup data to the Space",
		Run: func(cmd *cobra.Command, args []string) {
			var backupService = storage.NewBackupService(&storage.BackupOption{Space: o, Operator: constants.StorageOperatorCli})
			backupService.Backup()
		},
	}
	o.AddFlags(cmd)
	return cmd
}

func NewCmdS3() *cobra.Command {
	o := options.NewBackupAwsOption()
	cmd := &cobra.Command{
		Use:   "s3",
		Short: "Backup data to Amazon S3 or S3-compatible storage",
		Run: func(cmd *cobra.Command, args []string) {
			var backupService = storage.NewBackupService(&storage.BackupOption{Aws: o, Operator: constants.StorageOperatorCli})
			backupService.Backup()
		},
	}
	o.AddFlags(cmd)
	return cmd
}

func NewCmdCos() *cobra.Command {
	o := options.NewBackupTencentCloudOption()
	cmd := &cobra.Command{
		Use:   "cos",
		Short: "Backup data to Tencent Cloud Object Storage (COS)",
		Run: func(cmd *cobra.Command, args []string) {
			var backupService = storage.NewBackupService(&storage.BackupOption{TencentCloud: o, Operator: constants.StorageOperatorCli})
			backupService.Backup()
		},
	}
	o.AddFlags(cmd)
	return cmd
}

func NewCmdFs() *cobra.Command {
	o := options.NewBackupFilesystemOption()
	cmd := &cobra.Command{
		Use:   "fs",
		Short: "Backup data to the local filesystem or disk",
		Run: func(cmd *cobra.Command, args []string) {
			var backupService = storage.NewBackupService(&storage.BackupOption{Filesystem: o, Operator: constants.StorageOperatorCli})
			backupService.Backup()
		},
	}
	o.AddFlags(cmd)
	return cmd
}
