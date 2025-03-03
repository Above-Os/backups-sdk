package options

import "github.com/spf13/cobra"

type DownloadOption struct {
	DownloadCdnUrl string
}

func NewDownloadOption() *DownloadOption {
	return &DownloadOption{}
}

func (d *DownloadOption) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&d.DownloadCdnUrl, "download-cdn-url", "", "Set the CDN accelerated download address in the format https://example.cdn.com. If not set, the default download address will be used")
}
