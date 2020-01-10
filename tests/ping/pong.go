package ping

import (
	pn "github.com/FLAGlab/DistributedPetriNets/petrinet"
)

func pongMain() {
	fmt.Println("init Ping net....")
	p := pn.Init(0, "PongPN")
	p.AddTransition(1, 1)
	p.AddPlace(1, 0, "pong")
	p.AddInArc(1, 1, 1)
	p.AddRemoteOutArc(1, 1, "ping")
	p.InitService()
	for {}
}
