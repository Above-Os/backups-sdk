package stats

import (
	"github.com/spf13/cobra"
	"olares.com/backups-sdk/pkg/options"
	"olares.com/backups-sdk/pkg/storage"
)

func NewCmdStats() *cobra.Command {
	rootStatsCmds := &cobra.Command{
		Use:               "stats",
		Short:             "Scan the repository and show basic statistics",
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	}

	rootStatsCmds.AddCommand(NewCmdSpace())
	rootStatsCmds.AddCommand(NewCmdS3())
	rootStatsCmds.AddCommand(NewCmdCos())
	rootStatsCmds.AddCommand(NewCmdFs())

	return rootStatsCmds
}

func NewCmdSpace() *cobra.Command {
	o := options.NewSnapshotsSpaceOption()
	cmd := &cobra.Command{
		Use:   "space",
		Short: "Repository stats from Space",
		Run: func(cmd *cobra.Command, args []string) {
			var statsService = storage.NewStatsService(&storage.SnapshotsOption{Space: o})
			statsService.Stats()
		},
	}
	o.AddFlags(cmd)
	return cmd
}

func NewCmdS3() *cobra.Command {
	o := options.NewSnapshotsAwsOption()
	cmd := &cobra.Command{
		Use:   "s3",
		Short: "Repository stats from S3",
		Run: func(cmd *cobra.Command, args []string) {
			var statsService = storage.NewStatsService(&storage.SnapshotsOption{Aws: o})
			statsService.Stats()
		},
	}
	o.AddFlags(cmd)
	return cmd
}

func NewCmdCos() *cobra.Command {
	o := options.NewSnapshotsTencentCloudOption()
	cmd := &cobra.Command{
		Use:   "cos",
		Short: "Repository stats from Tencent COS",
		Run: func(cmd *cobra.Command, args []string) {
			var statsService = storage.NewStatsService(&storage.SnapshotsOption{TencentCloud: o})
			statsService.Stats()
		},
	}
	o.AddFlags(cmd)
	return cmd
}

func NewCmdFs() *cobra.Command {
	o := options.NewSnapshotsFilesystemOption()
	cmd := &cobra.Command{
		Use:   "fs",
		Short: "Repository stats from FileSystem",
		Run: func(cmd *cobra.Command, args []string) {
			var statsService = storage.NewStatsService(&storage.SnapshotsOption{Filesystem: o})
			statsService.Stats()

		},
	}
	o.AddFlags(cmd)
	return cmd
}
