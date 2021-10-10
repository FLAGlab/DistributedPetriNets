package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/FLAGlab/DistributedPetriNets/petrinet"
)

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

func main() {
	pn := petrinet.PetriNet{}
	file, _ := os.ReadFile("p1.json")
	err := json.Unmarshal([]byte(file), &pn)
	if err != nil {
		fmt.Println(err)
	}
	pn.Init()
	fmt.Println(prettyPrint(pn))
}
