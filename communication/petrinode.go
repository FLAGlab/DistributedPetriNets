package communication

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/FLAGlab/DCoPN/petrinet"
	"github.com/perlin-network/noise"
	"github.com/perlin-network/noise/skademlia"
)

type petriNode struct {
	node *noise.Node
	petriNet *petrinet.PetriNet
	isLeader bool
	transitionOptions map[string][]*petrinet.Transition
	peerCache map[string]*noise.Peer
	mux sync.Mutex
	step int
	pMsg chan petriMessage
}

func (pn *petriNode) incStep() {
	pn.mux.Lock()
	defer pn.mux.Unlock()
	pn.step = (pn.step + 1) % 4
}

func (pn *petriNode) resetStep() {
	pn.mux.Lock()
	defer pn.mux.Unlock()
	pn.step = 0
}

func (pn *petriNode) initTransitionOptions() {
	pn.mux.Lock()
	defer pn.mux.Unlock()
	pn.transitionOptions = make(map[string][]*petrinet.Transition)
	pn.transitionOptions[pn.node.ExternalAddress()] = pn.petriNet.GetTransitionOptions()
	pn.peerCache = make(map[string]*noise.Peer)
}

func (pn *petriNode) addTransitionOption(key string, options []*petrinet.Transition) int {
	pn.mux.Lock()
	defer pn.mux.Unlock()
	pn.transitionOptions[key] = options
	return len(pn.transitionOptions)
}

func (pn *petriNode) getTransition(pMsg petriMessage) {
	fmt.Printf("Received msg %v\n", pMsg)
	if pMsg.Command != TransitionCommand {
		fmt.Println("Expected transition, received something else")
		pn.resetStep()
		return
	}
	fmt.Printf("Received options %v\n", pMsg.Transitions)
	numDone := pn.addTransitionOption(pMsg.Address, pMsg.Transitions)
	expected := len(skademlia.Table(pn.node).GetPeers()) + 1 // plus me
	fmt.Printf("Done with: %v, Expected: %v\n", numDone, expected)
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

	myOptions := pn.transitionOptions[pn.node.ExternalAddress()]
	add := 0
	if len(myOptions) > 0 {
		add = 1
	}
	if initial + add == 0 {
		return nil, ""
	}
	pnNodeIndex := rand.Intn(initial + add)
	chosenKey := pn.node.ExternalAddress()
	options := pn.transitionOptions[chosenKey]
	if pnNodeIndex < initial {
		chosenKey = indexToKey[pnNodeIndex]
		options = pn.transitionOptions[chosenKey]
	}
	transitionIndex := rand.Intn(len(options))
	return options[transitionIndex], chosenKey
}

func (pn *petriNode) fireTransition() error {
	transition, peerAddr := pn.selectTransition()
	if transition == nil {
		fmt.Println("NO TRANSITION TO SELECT")
		return nil
	}
	fmt.Printf("SELECTED TRANSITION: %v\n", transition)
	fmt.Printf("Of peer: %v\n", peerAddr)
	msgToSend := petriMessage{
		Command: FireCommand,
		Address: pn.node.ExternalAddress(),
		Transitions: []*petrinet.Transition{transition}}
	if peerAddr == pn.node.ExternalAddress() {
		pn.petriNet.FireTransitionByID(transition.ID)
		return nil
	}
	fmt.Printf("TRANSITION IS REMOTE: %v\n", pn.peerCache[peerAddr])
	if peerAddr != "" && pn.peerCache[peerAddr] != nil {
		err := pn.peerCache[peerAddr].SendMessage(msgToSend)
		fmt.Printf("Error sending message from cache peer: %v\n", err)
		if err == nil {
				return err // everything ok
		} // else will try to dial
	}
	fmt.Println("WILL DIAL")
	peer, err := pn.node.Dial(peerAddr)
	fmt.Println("DONE DIAL")
	if err != nil {
		fmt.Printf("Error dialing: %v\n", peerAddr)
		return err
	}
	pn.peerCache[peerAddr] = peer
	return pn.peerCache[peerAddr].SendMessage(msgToSend)
}

func (pn *petriNode) printPetriNet() {
	fmt.Printf("%v\n", pn.petriNet)
	skademlia.Broadcast(pn.node, petriMessage{Command: PrintCommand, Address: pn.node.ExternalAddress()})
}

func (pn *petriNode) runLeader() {
	go func() {
		for  {
			switch pn.step {
			case 0:
				pn.ask()
			case 1:
				fmt.Println("WILL GET TRANSITION")
				select {
				case msg := <- pn.pMsg:
					pn.getTransition(msg)
				case <-time.After(5 * time.Second):
					pn.resetStep() // ask again
				}
			case 2:
				fmt.Println("WILL FIRE")
				pn.fireTransition()
				pn.incStep()
			case 3:
				fmt.Println("WILL PRINT")
				pn.printPetriNet()
				pn.incStep()
			default:
				time.Sleep(5 * time.Second)
			}
		}
	}()
}

func (pn *petriNode) initLeader() {
	pn.mux.Lock()
	defer pn.mux.Unlock()
	pn.isLeader = true
	pn.pMsg = make(chan petriMessage)
}

func (pn *petriNode) closeLeader() {
	pn.mux.Lock()
	defer pn.mux.Unlock()
	pn.isLeader = false
	close(pn.pMsg)
}


func (pn *petriNode) ask() {
	node := pn.node
	fmt.Println("Will Broadcast")
	errChan := make(chan []error)
	defer close(errChan)
	go func() {
		err := skademlia.Broadcast(node, petriMessage{Command: TransitionCommand, Address: node.ExternalAddress()})
		errChan <- err
	}()
	select {
	case err := <- errChan:
		fmt.Printf("Broadcast error: %v\n", err)
		fmt.Println("Done Broadcast")
		pn.initTransitionOptions()
		pn.incStep()
	case <-time.After(5 * time.Second):
		fmt.Println("Not even if I shout :'v ...")
		pn.resetStep()
	}

}
