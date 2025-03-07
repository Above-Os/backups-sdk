package options

import "github.com/spf13/cobra"

var _ Option = &SpaceRegionOptions{}

type SpaceRegionOptions struct {
	OlaresDid      string `json:"olares_did"`
	AccessToken    string `json:"access_token"`
	CloudApiMirror string `json:"cloud_api_mirror"`
}

func NewRegionSpaceOption() *SpaceRegionOptions {
	return &SpaceRegionOptions{}
}

func (s *SpaceRegionOptions) AddFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&s.OlaresDid, "olares-did", "", "", "Olares DID")
	cmd.Flags().StringVarP(&s.AccessToken, "access-token", "", "", "Space Access Token")
	cmd.Flags().StringVarP(&s.CloudApiMirror, "cloud-api-mirror", "", "", "Cloud API mirror")
}
