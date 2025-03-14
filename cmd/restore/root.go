package restore

import (
	"bytetrade.io/web3os/backups-sdk/pkg/options"
	"bytetrade.io/web3os/backups-sdk/pkg/storage"
	"github.com/spf13/cobra"
)

func NewCmdRestore() *cobra.Command {
	rootBackupCmds := &cobra.Command{
		Use:               "restore",
		Short:             "Restore data from multiple storage targets: Space, S3, COS, and local",
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	}

	rootBackupCmds.AddCommand(NewCmdSpace())
	rootBackupCmds.AddCommand(NewCmdS3())
	rootBackupCmds.AddCommand(NewCmdCos())
	rootBackupCmds.AddCommand(NewCmdFs())

	return rootBackupCmds
}

func NewCmdSpace() *cobra.Command {
	o := options.NewRestoreSpaceOption()
	cmd := &cobra.Command{
		Use:   "space",
		Short: "Restore data from Space",
		Run: func(cmd *cobra.Command, args []string) {
			var restoreService = storage.NewRestoreService(&storage.RestoreOption{Space: o})
			restoreService.Restore()
		},
	}
	o.AddFlags(cmd)
	return cmd
}

func NewCmdS3() *cobra.Command {
	o := options.NewRestoreS3Option()
	cmd := &cobra.Command{
		Use:   "s3",
		Short: "Restore data from Amazon S3 or S3-compatible storage",
		Run: func(cmd *cobra.Command, args []string) {
			var restoreService = storage.NewRestoreService(&storage.RestoreOption{S3: o})
			restoreService.Restore()
		},
	}
	o.AddFlags(cmd)
	return cmd
}

func NewCmdCos() *cobra.Command {
	o := options.NewRestoreCosOption()
	cmd := &cobra.Command{
		Use:   "cos",
		Short: "Restore data from Tencent Cloud Object Storage (COS)",
		Run: func(cmd *cobra.Command, args []string) {
			var restoreService = storage.NewRestoreService(&storage.RestoreOption{Cos: o})
			restoreService.Restore()
		},
	}
	o.AddFlags(cmd)
	return cmd
}

func NewCmdFs() *cobra.Command {
	o := options.NewRestoreFilesystemOption()
	cmd := &cobra.Command{
		Use:   "fs",
		Short: "Restore data from the local filesystem or disk",
		Run: func(cmd *cobra.Command, args []string) {
			var restoreService = storage.NewRestoreService(&storage.RestoreOption{Filesystem: o})
			restoreService.Restore()
		},
	}
	o.AddFlags(cmd)
	return cmd
}
