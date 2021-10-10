package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/FLAGlab/DistributedPetriNets/petrinet"
)

func main() {
	// interf := os.Args[1]
	// name := os.Args[2]
	// fmt.Printf("init Ping net.... %v\n", interf)
	// p := pn.InitPN(0)
	// p.AddTransition(1, 1)
	// p.AddPlace(1, "pong", name)
	// p.AddInArc(1, 1, 1)
	// p.AddRemoteOutArc(1, 1, "ping")
	// p.Places[1].AddMarks([]pn.Token{pn.Token{2}})
	// p.InitService(interf)
	pn := petrinet.PetriNet{}
	file, _ := os.ReadFile("pong.json")
	err := json.Unmarshal([]byte(file), &pn)
	if err != nil {
		fmt.Println(err)
	}
	pn.Init()
}
