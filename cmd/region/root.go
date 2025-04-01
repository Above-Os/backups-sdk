package region

import (
	"fmt"
	"os"

	"bytetrade.io/web3os/backups-sdk/pkg/options"
	"bytetrade.io/web3os/backups-sdk/pkg/storage"
	"bytetrade.io/web3os/backups-sdk/pkg/utils"
	"github.com/spf13/cobra"
)

func NewCmdRegions() *cobra.Command {
	rootCmdRegions := &cobra.Command{
		Use:               "region",
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
			data, err := regionService.Regions()
			if err != nil {
				panic(fmt.Errorf("Get space regions error: %v\n", err))
			}
			fmt.Println(utils.ToJSON(data))
			os.Exit(0)
		},
	}
	o.AddFlags(cmd)
	return cmd
}
