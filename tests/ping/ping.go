package ping

import (
	"fmt"
	pn "github.com/FLAGlab/DistributedPetriNets/petrinet"
)

func pingMain() {
	fmt.Println("init Ping net....")
	p := pn.Init(0, "PingPN")
	p.AddTransition(1, 1)
	p.AddPlace(1, 1, "ping")
	p.AddInArc(1, 1, 1)
	p.AddRemoteOutArc(1, 1, "pong")
	p.InitService()
	for {}
}
