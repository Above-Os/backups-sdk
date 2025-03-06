package snapshots

import (
	"bytetrade.io/web3os/backups-sdk/pkg/options"
	"bytetrade.io/web3os/backups-sdk/pkg/storage"
	"github.com/spf13/cobra"
)

func NewCmdSnapshots() *cobra.Command {
	rootSnapshotsCmds := &cobra.Command{
		Use:               "snapshots",
		Short:             "Olares Backup Tool",
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
	o := options.NewSnapshotsS3Option()
	cmd := &cobra.Command{
		Use:   "s3",
		Short: "Backup snapshots from S3",
		Run: func(cmd *cobra.Command, args []string) {
			var snapshotsService = storage.NewSnapshotsService(&storage.SnapshotsOption{S3: o})
			snapshotsService.Snapshots()
		},
	}
	o.AddFlags(cmd)
	return cmd
}

func NewCmdCos() *cobra.Command {
	o := options.NewSnapshotsCosOption()
	cmd := &cobra.Command{
		Use:   "cos",
		Short: "Backup snapshots from Tencent COS",
		Run: func(cmd *cobra.Command, args []string) {
			var snapshotsService = storage.NewSnapshotsService(&storage.SnapshotsOption{Cos: o})
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
