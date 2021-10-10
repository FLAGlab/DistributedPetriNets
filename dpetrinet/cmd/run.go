package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/FLAGlab/DistributedPetriNets/petrinet"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a  distributed petrinet node using a json file",
	Long:  `Run a distributed petrinet node usgin the path defined as first arg`,
	Run: func(cmd *cobra.Command, args []string) {
		pn := petrinet.PetriNet{}
		file, _ := os.ReadFile(args[0])
		err := json.Unmarshal([]byte(file), &pn)
		if err != nil {
			fmt.Println(err)
		}
		pn.Init()
	},
}
