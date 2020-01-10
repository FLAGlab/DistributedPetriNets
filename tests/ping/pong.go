package main

import (
	pn "github.com/FLAGlab/DistributedPetriNets/petrinet"
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ursiform/sleuth"
)

func main() {
	fmt.Println("init Pong net....")
	p := pn.InitPN(0)
	p.AddTransition(1, 1)
	p.AddPlace(1, 0, "pong")
	p.AddInArc(1, 1, 1)
	p.AddRemoteOutArc(1, 1, "ping")
	p.InitService()
	p.Run()
}
