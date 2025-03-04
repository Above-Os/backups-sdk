package options

import (
	"github.com/spf13/cobra"
)

type Option interface {
	AddFlags(cmd *cobra.Command)
}
