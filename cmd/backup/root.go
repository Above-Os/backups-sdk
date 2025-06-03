package backup

import (
	"context"

	"github.com/spf13/cobra"
	"olares.com/backups-sdk/pkg/constants"
	"olares.com/backups-sdk/pkg/options"
	"olares.com/backups-sdk/pkg/storage"
	"olares.com/backups-sdk/pkg/utils"
)

var p = func(percentDone float64) {}
var dryRun = false

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
			var backupService = storage.NewBackupService(&storage.BackupOption{Ctx: context.WithValue(context.TODO(), constants.TraceId, utils.NewUUID()), Space: o, Operator: constants.StorageOperatorCli})
			backupService.Backup(dryRun, p)
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
			var backupService = storage.NewBackupService(&storage.BackupOption{Ctx: context.WithValue(context.TODO(), constants.TraceId, utils.NewUUID()), Aws: o, Operator: constants.StorageOperatorCli})
			backupService.Backup(dryRun, p)
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
			var backupService = storage.NewBackupService(&storage.BackupOption{Ctx: context.WithValue(context.TODO(), constants.TraceId, utils.NewUUID()), TencentCloud: o, Operator: constants.StorageOperatorCli})
			backupService.Backup(dryRun, p)
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
			var backupService = storage.NewBackupService(&storage.BackupOption{Ctx: context.WithValue(context.TODO(), constants.TraceId, utils.NewUUID()), Filesystem: o, Operator: constants.StorageOperatorCli})
			backupService.Backup(dryRun, p)
		},
	}
	o.AddFlags(cmd)
	return cmd
}
