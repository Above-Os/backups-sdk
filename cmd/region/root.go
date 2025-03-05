package region

import (
	"bytetrade.io/web3os/backups-sdk/cmd/options"
	"bytetrade.io/web3os/backups-sdk/pkg/storage"
	"github.com/spf13/cobra"
)

func NewCmdRegions() *cobra.Command {
	rootCmdRegions := &cobra.Command{
		Use:               "regions",
		Short:             "Olares Storage Support Regions",
		CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	}

	rootCmdRegions.AddCommand(newSpaceRegions())

	return rootCmdRegions
}

func newSpaceRegions() *cobra.Command {
	o := options.NewRegionSpaceOption()
	cmd := &cobra.Command{
		Use:   "space",
		Short: "Space Storage Regions",
		Run: func(cmd *cobra.Command, args []string) {
			var regionService = storage.NewRegionService(&storage.RegionOption{Space: o})
			regionService.Regions()
		},
	}
	o.AddFlags(cmd)
	return cmd
}
