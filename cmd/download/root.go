package download

import (
	"bytetrade.io/web3os/backups-sdk/cmd/options"
	"bytetrade.io/web3os/backups-sdk/pkg/files"
	"github.com/spf13/cobra"
)

// download restic
func NewCmdDownload() *cobra.Command {
	o := options.NewDownloadOption()
	rootDownloadCmd := &cobra.Command{
		Use:   "download",
		Short: "Download restic",
		Run: func(cmd *cobra.Command, args []string) {
			files.Download(o.DownloadCdnUrl)
		},
	}

	o.AddFlags(rootDownloadCmd)
	return rootDownloadCmd
}
