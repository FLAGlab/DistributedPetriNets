package communication

import (
	"fmt"
	"math"
	"time"

	"github.com/FLAGlab/DCoPN/petrinet"
	"github.com/perlin-network/noise"
	"github.com/perlin-network/noise/skademlia"
)

type LeaderStep int

const (
	ASK_STEP LeaderStep = 0
	RECEIVING_TRANSITIONS_STEP LeaderStep = 1
	PREPARE_FIRE_STEP LeaderStep = 2
	RECEIVING_MARKS_STEP LeaderStep = 3
	FIRE_STEP LeaderStep = 4
	PRINT_STEP LeaderStep = 5
)

type petriNode struct {
	node *noise.Node
	petriNet *petrinet.PetriNet
	step LeaderStep
	timeoutCount int
	transitionOptions map[string][]*petrinet.Transition
	remoteTransitionOptions map[string]map[int]*petrinet.RemoteTransition
	pMsg chan petriMessage
	chosenTransition *petrinet.Transition
	chosenRemoteTransition *petrinet.RemoteTransition
	chosenTransitionAddress string
	addressMissing map[string]bool
	verifiedRemoteAddrs []string
	marks map[string]map[int]*petrinet.RemoteArc
	contextToAddrs map[string][]string
	addrsToContext map[string]string
}

func InitPetriNode(node *noise.Node, petriNet *petrinet.PetriNet) *petriNode {
	return &petriNode{
		node: node,
		petriNet: petriNet,
		contextToAddrs: make(map[string][]string),
		addrsToContext: make(map[string]string)}
}

func (pn *petriNode) incStep() {
	pn.step = (pn.step + 1) % (PRINT_STEP + 1)
}

func (pn *petriNode) resetStep() {
	pn.step = ASK_STEP
}

func (pn *petriNode) initTransitionOptions() {
	pn.transitionOptions = make(map[string][]*petrinet.Transition)
	pn.remoteTransitionOptions = make(map[string]map[int]*petrinet.RemoteTransition)
	pn.transitionOptions[pn.node.ExternalAddress()], pn.remoteTransitionOptions[pn.node.ExternalAddress()] = pn.petriNet.GetTransitionOptions()
}

func (pn *petriNode) addTransitionOption(key string, options []*petrinet.Transition, remote map[int]*petrinet.RemoteTransition) int {
	pn.transitionOptions[key] = options
	pn.remoteTransitionOptions[key] = remote
	return len(pn.transitionOptions)
}

func (pn *petriNode) getTransition(pMsg petriMessage) {
	if pMsg.Command != TransitionCommand {
		fmt.Printf("Expected transition, received something else: %v HERE\n", pMsg.Command)
		pn.resetStep()
		return
	}
	fmt.Printf("HERE: %v\n", pMsg)
	fmt.Printf("HERE: %v\n", pn.transitionOptions)
	numDone := pn.addTransitionOption(pMsg.Address, pMsg.Transitions, pMsg.RemoteTransitions)
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

func (pn *petriNode) askForMarks(remoteTransition *petrinet.RemoteTransition, baseMsg petriMessage) map[string]bool {
	connectedAddrs := make(map[string]bool)
	if remoteTransition == nil {
		return connectedAddrs
	}
	baseMsg.Command = MarksCommand
	for rmtAddr, places := range remoteTransition.GetPlaceIDsByAddrs() {
		// places is []int
		baseMsg2 := baseMsg
		baseMsg2.RemoteArcs = make([]*petrinet.RemoteArc, len(places))
		for i, p := range places {
			baseMsg2.RemoteArcs[i] = &petrinet.RemoteArc{PlaceID: p}
		}
		var err error
		if rmtAddr == pn.node.ExternalAddress() {
			for _, rmtArc := range baseMsg2.RemoteArcs {
				pn.saveMarks(rmtAddr, rmtArc.PlaceID, rmtArc)
			}
		} else {
			err = pn.SendMessageByAddress(baseMsg2, rmtAddr)
			if err == nil {
				connectedAddrs[rmtAddr] = true
			}
		}
	}
	return connectedAddrs
}

// if transition option is not valid, remove it
func (pn *petriNode) removeTransitionOption(addrs string, transition *petrinet.Transition) {
	// delete transition from option list
	numElem := len(pn.transitionOptions[addrs])
	if numElem - 1 == 0 {
		pn.transitionOptions[addrs] = []*petrinet.Transition{}
	} else {
		delIndex := -1
		for i, v := range pn.transitionOptions[addrs] {
	    if v.ID == transition.ID {
				delIndex = i
				break
	    }
		}
		if delIndex != -1 {
			pn.transitionOptions[addrs][delIndex] = pn.transitionOptions[addrs][numElem - 1]
			pn.transitionOptions[addrs] = pn.transitionOptions[addrs][:numElem - 1]
		}
	}
	// delete remote transition from de transition
	delete(pn.remoteTransitionOptions[addrs], transition.ID)
}

func (pn *petriNode) saveMarks(addr string, placeID int, rmtArc *petrinet.RemoteArc) {
	placeMap, exists := pn.marks[addr]
	if !exists {
		pn.marks[addr] = make(map[int]*petrinet.RemoteArc)
		placeMap = pn.marks[addr]
	}
	placeMap[placeID] = rmtArc
}

func (pn *petriNode) getPlaceMarks(pMsg petriMessage) {
	fmt.Println("RECEIVED A MARK MSG")
	if pMsg.Command != MarksCommand {
		fmt.Printf("Expected marks, received something else: %v HERE\n", pMsg.Command)
		pn.resetStep()
		return
	}
	for _, rmtArc := range pMsg.RemoteArcs {
		pn.saveMarks(pMsg.Address, rmtArc.PlaceID, rmtArc)
	}
	fmt.Printf("CURR ADDRESS: %v\n", pMsg.Address)
	fmt.Printf("ADDRESS MISSING: %v\n", pn.addressMissing)
	fmt.Printf("MARKS: %v\n", pn.marks)
	_, present := pn.addressMissing[pMsg.Address]
	if present {
		fmt.Println("WILL DELETE FROM ADRESS MISSING")
		pn.verifiedRemoteAddrs = append(pn.verifiedRemoteAddrs, pMsg.Address)
		delete(pn.addressMissing, pMsg.Address)
	}
	fmt.Printf("ADDRESS MISSING AFTER DELETION: %v\n", pn.addressMissing)
	if len(pn.addressMissing) == 0 {
		if !pn.validateRemoteTransitionMarks() {
			// transition wasn't ready to fire, remove from options and try again
			pn.removeTransitionOption(pn.chosenTransitionAddress, pn.chosenTransition)
			pn.step = PREPARE_FIRE_STEP
		} else {
			pn.incStep() // FIRE_STEP
		}
	}
}

// after receiving all valid transitions, do this first and wait for askedAddrs to respond
func (pn *petriNode) prepareFire(baseMsg petriMessage) {
	fmt.Println("PRERAREFIRE METH CALLED")
	transition, peerAddr := pn.selectTransition()
	pn.chosenTransition = transition
	pn.chosenTransitionAddress = peerAddr
	if transition == nil {
		fmt.Println("_NO TRANSITION TO SELECT")
		pn.resetStep()
	} else {
		pn.verifiedRemoteAddrs = []string{}
		pn.marks = make(map[string]map[int]*petrinet.RemoteArc)
		rmtTransitionOption, ok := pn.remoteTransitionOptions[peerAddr][transition.ID]
		if ok {
			copy := *rmtTransitionOption // get a copy
			remoteTransition := &copy // pointer to the copy
			remoteTransition.UpdateAddressByContext(pn.contextToAddrs)
			pn.chosenRemoteTransition = remoteTransition
			askedAddrs := pn.askForMarks(remoteTransition, baseMsg)
			fmt.Println(askedAddrs)
			pn.addressMissing = askedAddrs
			pn.incStep() // RECEIVING_MARKS_STEP
			if len(pn.addressMissing) == 0 {
				// skip RECEIVING_MARKS_STEP
				pn.incStep() // FIRE_STEP
			}
		} else {
			// there's nothing remote to fire, skip to FIRE_STEP
			pn.incStep() // RECEIVING_MARKS_STEP
			pn.incStep() // FIRE_STEP
		}
	}
}

// after all askedAddrs responded, do this with the chosen transition
// if timeout should try again, if not valid leader should delete transition and try again PREPARE_FIRE_STEP
func (pn *petriNode) validateRemoteTransitionMarks() bool {
	fmt.Println("ALL MARKS RECEIVED, WILL VALIDATE REMOTE TRANSITION")
	ans := true
	marks := pn.marks
	fmt.Printf("MARKS: %v\n", marks)
	rmtTransition := pn.chosenRemoteTransition
	fmt.Printf("RMT TRANSITION: %v\n", rmtTransition)
	helperFunc := func (arcList []petrinet.RemoteArc, comp func(int, int)bool) {
		for _, currArc := range arcList { //rmtTransition.InArcs {
			place, exists := marks[currArc.Address][currArc.PlaceID]
			if !exists {
				continue
			}
			ans = ans && comp(place.Marks, currArc.Weight)
		}
	}
	helperFunc(rmtTransition.InArcs, func(a, b int) bool { return a >= b})
	helperFunc(rmtTransition.InhibitorArcs, func(a, b int) bool { return a < b})
	return ans
}


func (pn *petriNode) fireRemoteTransition(t *petrinet.RemoteTransition, baseMsg petriMessage) {
	if t != nil {
		fmt.Println("WILL FIRE REMOTE TRANSITION METH")
		helperFunc := func(opType petrinet.OperationType, addrToArcMap map[string][]*petrinet.RemoteArc) {
			fmt.Println("RUNNING HELPER FUNC")
			fmt.Printf("VERIFIED REMOTE ADDRS: %v\n", pn.verifiedRemoteAddrs)
			for _, addr := range pn.verifiedRemoteAddrs {
				baseMsg2 := baseMsg
				baseMsg2.Command = AddToPlacesCommand
				baseMsg2.RemoteArcs = addrToArcMap[addr]
				baseMsg2.OpType = opType
				fmt.Println("WILL SEE IF IT SHOULD FIRE REMOTE")
				if len(baseMsg2.RemoteArcs) > 0 {
					fmt.Println("WILL FIRE REMOTE")
					if addr == pn.node.ExternalAddress() {
						pn.petriNet.AddMarksToPlaces(opType, baseMsg2.RemoteArcs)
						fmt.Println("REMOTE WAS LOCAL, FIRED IMMEDIATLY")
					} else {
						pn.SendMessageByAddress(baseMsg2, addr)
						fmt.Printf("SENT MSG TO ADDRES %v\n", addr)
					}
				}
		  }
		}
		placesToFire := t.GetInArcsByAddrs()
		placesToReceive := t.GetOutArcsByAddrs()
		fmt.Printf("REMOTE IN ARCS TO FIRE: %v\n", placesToFire)
		fmt.Printf("REMOTE OUT ARCS TO FIRE: %v\n", placesToReceive)
		helperFunc(petrinet.SUBSTRACTION, placesToFire)
		helperFunc(petrinet.ADDITION, placesToReceive)
	}
}

func (pn *petriNode) fireTransition(baseMsg petriMessage) error {
	fmt.Println("WILL FIRE TRANSITION METH")
	transition := pn.chosenTransition
	peerAddr := pn.chosenTransitionAddress
	remoteTransition := pn.chosenRemoteTransition
	baseMsg.Command = FireCommand
	baseMsg.Transitions = []*petrinet.Transition{transition}
	var err error
	if peerAddr == pn.node.ExternalAddress() {
		// Transition is local
		fmt.Printf("_WILL FIRE TRANSITION %v\n", transition.ID)
		pn.petriNet.FireTransitionByID(transition.ID)
		err = nil
	} else {
		// Transition is remote
		fmt.Println("WILL SEND MSG TO FIRE TRANSITION")
		err = pn.SendMessageByAddress(baseMsg, peerAddr)
		fmt.Println("DONE  SENDING MSG TO FIRE TRANSITION")
	}
	if err == nil {
		// Transition fired with no problem
		fmt.Println("NO ERROR, WILL FIRE REMOTE TRANSITION")
		pn.fireRemoteTransition(remoteTransition, baseMsg) // Fire remote transition
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

func (pn *petriNode) showPetriNetCurrentState() {
	fmt.Println("Will print petri net")
	fmt.Printf("%v\n", pn.petriNet)
	time.Sleep(time.Duration(humanTimeout) * time.Millisecond)
}

func (pn *petriNode) printPetriNet(baseMsg petriMessage) {
	baseMsg.Command = PrintCommand
	skademlia.Broadcast(pn.node, baseMsg)
	pn.showPetriNetCurrentState()
	pn.incStep()
}

func (pn *petriNode) broadcastWithTimeout(msg petriMessage, successCallback, timeoutCallback func()) {
	errChan := make(chan []error)
	defer close(errChan)
	go func() {
		fmt.Printf("Doing broadcast of %v...\n", msg)
		err := skademlia.Broadcast(pn.node, msg)
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
		baseMsg.Transitions, baseMsg.RemoteTransitions = pn.petriNet.GetTransitionOptions()
		pn.SendMessageByAddress(baseMsg, pMsg.Address)
	case MarksCommand:
		baseMsg.Command = MarksCommand
		pn.petriNet.CopyPlaceMarksToRemoteArc(pMsg.RemoteArcs)
		fmt.Printf("COPY PLACE MARKS TO RMTARC RES: %v\n", pMsg.RemoteArcs)
		baseMsg.RemoteArcs = pMsg.RemoteArcs
		fmt.Printf("WILL SEND MSG: %v\n", baseMsg)
		pn.SendMessageByAddress(baseMsg, pMsg.Address)
	case FireCommand:
		transitionID := pMsg.Transitions[0].ID
		fmt.Printf("_WILL FIRE TRANSITION %v\n", transitionID)
		err := pn.petriNet.FireTransitionByID(transitionID)
		if err != nil {
			fmt.Println(err)
		}
	case PrintCommand:
		pn.showPetriNetCurrentState()
	case AddToPlacesCommand:
		pn.petriNet.AddMarksToPlaces(pMsg.OpType, pMsg.RemoteArcs)
	default:
		fmt.Printf("Unknown command: %v\n", pMsg.Command)
	}
}

func (pn *petriNode) updateCtx(pMsg petriMessage) {
	ctx := pMsg.PetriContext
	addr := pMsg.Address
	oldCtx, exists := pn.addrsToContext[addr]
	if !exists {
		pn.addrsToContext[addr] = ctx
		pn.contextToAddrs[ctx] = append(pn.contextToAddrs[ctx], addr)
	} else if oldCtx != ctx {
		pn.addrsToContext[addr] = ctx
		pn.contextToAddrs[ctx] = removeStringList(ctx, pn.contextToAddrs[ctx])
		pn.contextToAddrs[ctx] = append(pn.contextToAddrs[ctx], addr)
	}
}

func removeStringList(elem string, list []string) []string {
	var ans []string
	for _, curr := range list {
		if elem != curr {
			ans = append(ans, elem)
		}
	}
	return ans
}
