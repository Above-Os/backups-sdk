package restore

import (
	"bytetrade.io/web3os/backups-sdk/cmd/options"
	"bytetrade.io/web3os/backups-sdk/pkg/storage"
	"github.com/spf13/cobra"
)

func NewCmdRestore() *cobra.Command {
	rootBackupCmds := &cobra.Command{
		Use:               "restore",
		Short:             "Olares Restore Tool",
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
		Short: "Restore files from Space",
		Run: func(cmd *cobra.Command, args []string) {
			var restoreService = storage.NewRestoreService(&storage.RestoreOption{Basedir: o.BaseDir, Space: o})
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
		Short: "Restore files from S3",
		Run: func(cmd *cobra.Command, args []string) {
			var restoreService = storage.NewRestoreService(&storage.RestoreOption{Basedir: o.BaseDir, S3: o})
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
		Short: "Restore files from Tencent COS",
		Run: func(cmd *cobra.Command, args []string) {
			var restoreService = storage.NewRestoreService(&storage.RestoreOption{Basedir: o.BaseDir, Cos: o})
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
		Short: "Restore files from FileSystem",
		Run: func(cmd *cobra.Command, args []string) {
			var restoreService = storage.NewRestoreService(&storage.RestoreOption{Basedir: o.BaseDir, Filesystem: o})
			restoreService.Restore()
		},
	}
	o.AddFlags(cmd)
	return cmd
}
