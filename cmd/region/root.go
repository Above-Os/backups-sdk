package region

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"olares.com/backups-sdk/pkg/options"
	"olares.com/backups-sdk/pkg/storage"
	"olares.com/backups-sdk/pkg/utils"
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
