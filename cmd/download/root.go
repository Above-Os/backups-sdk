package download

import (
	"bytetrade.io/web3os/backups-sdk/pkg/file"
	"bytetrade.io/web3os/backups-sdk/pkg/options"
	"github.com/spf13/cobra"
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
