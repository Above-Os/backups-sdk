package download

import (
	"github.com/spf13/cobra"
	"olares.com/backups-sdk/pkg/file"
	"olares.com/backups-sdk/pkg/options"
)

func NewCmdDownload() *cobra.Command {
	o := options.NewDownloadOption()
	cmds := &cobra.Command{
		Use:   "download",
		Short: "Download the backup dependency tool Restic",
		Run: func(cmd *cobra.Command, args []string) {
			if err := file.Download(o.DownloadCdnUrl); err != nil {
				panic(err)
			}
		},
	}
	o.AddFlags(cmds)
	return cmds
}
