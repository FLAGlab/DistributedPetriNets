package cmd

import (
	"github.com/FLAGlab/DistributedPetriNets/reachability"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(serverCmd)
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Run the server that calculate the reachability graph",
	Long:  `Run the server that calculate the complete reachability graph`,
	Run: func(cmd *cobra.Command, args []string) {
		reachability.Run()
	},
}
