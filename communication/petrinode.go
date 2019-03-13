package communication

import (
	"fmt"
	"math"
	"time"

	"github.com/FLAGlab/DCoPN/petrinet"
	"github.com/perlin-network/noise"
	"github.com/perlin-network/noise/skademlia"
	"github.com/perlin-network/noise/protocol"
)

type petriNode struct {
	node *noise.Node
	petriNet *petrinet.PetriNet
	step int
	timeoutCount int
	transitionOptions map[string][]*petrinet.Transition
	pMsg chan petriMessage
}

func (pn *petriNode) incStep() {
	pn.step = (pn.step + 1) % 4
}

func (pn *petriNode) resetStep() {
	pn.step = 0
}

func (pn *petriNode) initTransitionOptions() {
	pn.transitionOptions = make(map[string][]*petrinet.Transition)
	pn.transitionOptions[pn.node.ExternalAddress()] = pn.petriNet.GetTransitionOptions()
}

func (pn *petriNode) addTransitionOption(key string, options []*petrinet.Transition) int {
	pn.transitionOptions[key] = options
	return len(pn.transitionOptions)
}

func (pn *petriNode) getTransition(pMsg petriMessage) {
	if pMsg.Command != TransitionCommand {
		fmt.Printf("Expected transition, received something else: %v HERE\n", pMsg.Command)
		pn.resetStep()
		return
	}
	numDone := pn.addTransitionOption(pMsg.Address, pMsg.Transitions)
	expected := len(skademlia.Table(pn.node).GetPeers()) + 1 // plus me
	if numDone == expected {
		pn.incStep()
	}
}

func (pn *petriNode) selectTransition() (*petrinet.Transition, string) {
	minPriority := math.MaxInt64
	for _, value := range pn.transitionOptions {
		if len(value) > 0 && value[0].Priority < minPriority{
			minPriority = value[0].Priority
		}
	}
	indexToKey := make(map[int]string)
	initial := 0
	for key, value := range pn.transitionOptions {
		if len(value) == 0 || value[0].Priority != minPriority {
			delete(pn.transitionOptions, key)
		} else {
			indexToKey[initial] = key
			initial++
		}
	}

	if initial == 0 { // there is no transition to pick
		return nil, ""
	}
	pnNodeIndex := getRand(initial)
	chosenKey := indexToKey[pnNodeIndex]
	options := pn.transitionOptions[chosenKey]
	transitionIndex := getRand(len(options))
	return options[transitionIndex], chosenKey
}

func (pn *petriNode) fireTransition(baseMsg petriMessage) error {
	transition, peerAddr := pn.selectTransition()
	if transition == nil {
		fmt.Println("_NO TRANSITION TO SELECT")
		pn.resetStep()
		return nil
	}
	// fmt.Printf("SELECTED TRANSITION: %v\n", transition)
	// fmt.Printf("Of peer: %v\n", peerAddr)
	baseMsg.Command = FireCommand
	baseMsg.Transitions = []*petrinet.Transition{transition}
	if peerAddr == pn.node.ExternalAddress() {
		fmt.Printf("_WILL FIRE TRANSITION %v\n", transition.ID)
		pn.petriNet.FireTransitionByID(transition.ID)
		pn.incStep()
		return nil
	}
	// fmt.Printf("TRANSITION IS REMOTE: %v\n", peerAddr)
	err := pn.SendMessageByAddress(baseMsg, peerAddr)
	if err == nil {
		pn.incStep()
	} else {
		pn.resetStep()
	}
	return err
}

func (pn *petriNode) ask(baseMsg petriMessage) {
	baseMsg.Command = TransitionCommand
	success := func() {
		fmt.Println("Broadcast ask done correctly")
		pn.initTransitionOptions()
		pn.incStep()
	}
	timeoutCallback := func() {
		fmt.Println("Broadcast ask NOT correct...")
		pn.resetStep()
	}
	pn.broadcastWithTimeout(baseMsg, success, timeoutCallback)
}

func (pn *petriNode) SendMessageByAddress(msgToSend petriMessage, peerAddr string) error {
	peer, err := pn.node.Dial(peerAddr)
	if err != nil {
		fmt.Printf("Error dialing: %v\n", peerAddr)
		return err
	}
	return peer.SendMessage(msgToSend)
}

func (pn *petriNode) printPetriNet(baseMsg petriMessage) {
	fmt.Println("Will print petri net")
	fmt.Printf("%v\n", pn.petriNet)
	baseMsg.Command = PrintCommand
	skademlia.Broadcast(pn.node, baseMsg)
	pn.incStep()
}

func (pn *petriNode) broadcastWithTimeout(msg petriMessage, successCallback, timeoutCallback func()) {
	errChan := make(chan []error)
	defer close(errChan)
	go func() {
		fmt.Printf("Doing broadcast of %v...\n", msg)
		err := myBroadcast(pn.node, msg)
		fmt.Println("Broadcast sent!")
		errChan <- err
	}()
	select {
	case <- errChan:
		// if len(errList) > 0 {
		// 	timeoutCallback()
		// } else {
		successCallback()
		// }
	case <- time.After(time.Duration(pn.timeoutCount + humanTimeout + 100000) * time.Millisecond):
		timeoutCallback()
	}
}

func (pn *petriNode) processMessage(pMsg petriMessage, baseMsg petriMessage) {
	switch pMsg.Command {
	case TransitionCommand:
		baseMsg.Command = TransitionCommand
		baseMsg.Transitions = pn.petriNet.GetTransitionOptions()
		pn.SendMessageByAddress(baseMsg, pMsg.Address)
	case FireCommand:
		transitionID := pMsg.Transitions[0].ID
		fmt.Printf("_WILL FIRE TRANSITION %v\n", transitionID)
		err := pn.petriNet.FireTransitionByID(transitionID)
		if err != nil {
			fmt.Println(err)
		}
	case PrintCommand:
		fmt.Println("CURRENT PETRI NET:")
		fmt.Printf("%v\n", pn.petriNet)
	default:
		fmt.Printf("Unknown command: %v\n", pMsg.Command)
	}
}

func myBroadcast(node *noise.Node, message noise.Message) (errs []error) {
	errorChannels := make([]<-chan error, 0)

	for index, peerID := range skademlia.FindClosestPeers(skademlia.Table(node), protocol.NodeID(node).Hash(), skademlia.BucketSize()) {
		peer := protocol.Peer(node, peerID)

		if peer == nil {
			continue
		}
		fmt.Printf("peer %v is not nil: %v:%v\n", index, peer.RemoteIP(), peer.RemotePort())
		errorChannels = append(errorChannels, peer.SendMessageAsync(message))
	}

	for index, ch := range errorChannels {
		fmt.Printf("Done with %v\n", index)
		err := <-ch
		if err != nil {
			errs = append(errs, err)
		}
	}

	return
}
