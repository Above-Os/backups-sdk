package snapshots

import (
	"github.com/spf13/cobra"
	"olares.com/backups-sdk/pkg/options"
	"olares.com/backups-sdk/pkg/storage"
)

func NewCmdSnapshots() *cobra.Command {
	rootSnapshotsCmds := &cobra.Command{
		Use:               "snapshots",
		Short:             "Manage and view backup snapshots",
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	}

	rootSnapshotsCmds.AddCommand(NewCmdSpace())
	rootSnapshotsCmds.AddCommand(NewCmdS3())
	rootSnapshotsCmds.AddCommand(NewCmdCos())
	rootSnapshotsCmds.AddCommand(NewCmdFs())

	return rootSnapshotsCmds
}

func NewCmdSpace() *cobra.Command {
	o := options.NewSnapshotsSpaceOption()
	cmd := &cobra.Command{
		Use:   "space",
		Short: "Backup snapshots from Space",
		Run: func(cmd *cobra.Command, args []string) {
			var snapshotsService = storage.NewSnapshotsService(&storage.SnapshotsOption{Space: o})
			snapshotsService.Snapshots()
		},
	}
	o.AddFlags(cmd)
	return cmd
}

func NewCmdS3() *cobra.Command {
	o := options.NewSnapshotsAwsOption()
	cmd := &cobra.Command{
		Use:   "s3",
		Short: "Backup snapshots from S3",
		Run: func(cmd *cobra.Command, args []string) {
			var snapshotsService = storage.NewSnapshotsService(&storage.SnapshotsOption{Aws: o})
			snapshotsService.Snapshots()
		},
	}
	o.AddFlags(cmd)
	return cmd
}

func NewCmdCos() *cobra.Command {
	o := options.NewSnapshotsTencentCloudOption()
	cmd := &cobra.Command{
		Use:   "cos",
		Short: "Backup snapshots from Tencent COS",
		Run: func(cmd *cobra.Command, args []string) {
			var snapshotsService = storage.NewSnapshotsService(&storage.SnapshotsOption{TencentCloud: o})
			snapshotsService.Snapshots()
		},
	}
	o.AddFlags(cmd)
	return cmd
}

func NewCmdFs() *cobra.Command {
	o := options.NewSnapshotsFilesystemOption()
	cmd := &cobra.Command{
		Use:   "fs",
		Short: "Backup snapshots from FileSystem",
		Run: func(cmd *cobra.Command, args []string) {
			var snapshotsService = storage.NewSnapshotsService(&storage.SnapshotsOption{Filesystem: o})
			snapshotsService.Snapshots()

		},
	}
	o.AddFlags(cmd)
	return cmd
}
