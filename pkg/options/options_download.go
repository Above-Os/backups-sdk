package options

import "github.com/spf13/cobra"

type DownloadOption struct {
	DownloadCdnUrl string
}

func NewDownloadOption() *DownloadOption {
	return &DownloadOption{}
}

func (o *DownloadOption) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&o.DownloadCdnUrl, "download-cdn-url", "", "", "Download CDN Url")
}
