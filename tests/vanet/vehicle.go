package main

import (
	"fmt"
	pn "github.com/FLAGlab/DistributedPetriNets/petrinet"
	"os"
	"strconv"
)

func main() {
	interf := os.Args[1]
	name := os.Args[2]
	token := os.Args[3]
	fmt.Printf("init Vehivle net.... %v\n", interf)
	p := pn.InitPN(0)
	p.AddTransition(1, 1) //send
	p.AddTransition(2, 1) //commit
	p.AddPlace(1, "car", name)
	p.AddPlace(2, "vsadb", name)
	tokenNum, _ := strconv.Atoi(token)
	/*if err == nil {
		fmt.Printf("worng parameter. Third parameter must be a number\n")
	}*/
	p.Places[2].AddMarks([]pn.Token{pn.Token{tokenNum}})
	p.AddInArc(1, 2, 1)
	p.AddOutArc(2, 2, 1)
	p.AddInArc(2, 1, 1)
	p.AddRemoteOutArc(1, 1, "rsu")
	p.InitService(interf)
}
