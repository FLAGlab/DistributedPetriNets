package communication

import (
	"flag"
	"fmt"
	"math"
	"time"

	"github.com/FLAGlab/DCoPN/petrinet"
)

type LeaderStep int

const (
	ASK_STEP LeaderStep = 0
	RECEIVING_TRANSITIONS_STEP LeaderStep = 1
	CHECK_CONFLICTED_STEP LeaderStep = 2
	RECEIVING_CONFLICTED_MARKS_STEP LeaderStep = 3
	PREPARE_FIRE_STEP LeaderStep = 4
	RECEIVING_MARKS_STEP LeaderStep = 5
	FIRE_STEP LeaderStep = 6
	PRINT_STEP LeaderStep = 7

	UNIVERSAL_PN string = "universal"
)

type PeerNode interface {
	SendMessage(pMsg petriMessage) error
}

type CommunicationNode interface {
	ExternalAddress() string
	Dial(address string) (PeerNode, error)
	CountPeers() int
	Broadcast(pMsg petriMessage) []error
}

type petriNode struct {
	node CommunicationNode //*noise.Node
	petriNet *petrinet.PetriNet
	universalPetriNet *petrinet.PetriNet
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
	priorityToAsk int
	maxPriority int
	lastMsgTo string
	transitionPicker RandomTransitionPicker
	didFire bool
	needsToCheckForConflictedState bool
}

type RandomTransitionPicker func(map[string][]*petrinet.Transition) (*petrinet.Transition, string)

func InitPetriNode(node CommunicationNode, petriNet *petrinet.PetriNet) *petriNode {
	return &petriNode{
		node: node,
		petriNet: petriNet,
		maxPriority: petriNet.GetMaxPriority(),
		contextToAddrs: make(map[string][]string),
		addrsToContext: make(map[string]string),
		transitionPicker: func(options map[string][]*petrinet.Transition) (*petrinet.Transition, string) {
			var indexToKey []string
			for key := range options {
				indexToKey = append(indexToKey, key)
			}
			pnNodeIndex := getRand(len(options))
			chosenKey := indexToKey[pnNodeIndex]
			tOptions := options[chosenKey]
			transitionIndex := getRand(len(tOptions))
			return tOptions[transitionIndex], chosenKey
		}}
}

// SetUniversalPetriNet adds a petri net whose purpose is to keep remote Transitions
// that require connection with all the contexts needed
func (pn *petriNode) SetUniversalPetriNet(upn *petrinet.PetriNet) {
	// transition goes from ctx1 to ctx2 1 on 1, if we have 2 ctx2 and 1 ctx1
	// it should be ctx1 -> t ->ctx2(1) and ctx1 -> t -> ctx2(2)
	pn.universalPetriNet = upn
}

func (pn *petriNode) incStep() {
	if pn.step == RECEIVING_TRANSITIONS_STEP && (!pn.didFire || !pn.needsToCheckForConflictedState) {
		pn.step = PREPARE_FIRE_STEP // skip conflicted steps
	} else {
		pn.step = (pn.step + 1) % (PRINT_STEP + 1)
	}
}

func (pn *petriNode) resetStep() {
	pn.step = ASK_STEP
}

func (pn *petriNode) resetLastMsgTo() {
	pn.lastMsgTo = ""
}

func (pn *petriNode) initTransitionOptions() {
	pn.transitionOptions = make(map[string][]*petrinet.Transition)
	pn.remoteTransitionOptions = make(map[string]map[int]*petrinet.RemoteTransition)
	pn.transitionOptions[pn.node.ExternalAddress()], pn.remoteTransitionOptions[pn.node.ExternalAddress()] = pn.petriNet.GetTransitionOptionsByPriority(pn.priorityToAsk)

	if pn.universalPetriNet != nil {
		pn.transitionOptions[UNIVERSAL_PN], pn.remoteTransitionOptions[UNIVERSAL_PN] = pn.universalPetriNet.GenerateUniversalTransitionsByPriority(pn.contextToAddrs, pn.priorityToAsk)
	}
}

func (pn *petriNode) addTransitionOption(key string, options []*petrinet.Transition, remote map[int]*petrinet.RemoteTransition) int {
	pn.transitionOptions[key] = options
	pn.remoteTransitionOptions[key] = remote
	return len(pn.transitionOptions)
}

func (pn *petriNode) updateMaxPriority(pMsg petriMessage) {
	fmt.Printf("WILL UPDATE MAX PRIORITY: %v\n", pMsg)
	if pMsg.imNew {
		pn.priorityToAsk = 0
	}
	if pMsg.AskedPriority > pn.maxPriority {
		pn.maxPriority = pMsg.AskedPriority
	}
	fmt.Printf("MAX PRIORITY AFTER %v\n", pn.maxPriority)
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
	expected := pn.node.CountPeers() + 1 // including me
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
	for key, value := range pn.transitionOptions {
		if len(value) == 0 || value[0].Priority != minPriority {
			delete(pn.transitionOptions, key)
		}
	}
	if len(pn.transitionOptions) == 0 { // there is no transition to pick
		return nil, ""
	}
	return pn.transitionPicker(pn.transitionOptions)
}

func (pn *petriNode) askForMarks(remoteTransition *petrinet.RemoteTransition, isUniversalPN bool, baseMsg petriMessage) (map[string]bool, error) {
	connectedAddrs := make(map[string]bool)
	if remoteTransition == nil {
		return connectedAddrs, nil
	}
	baseMsg.Command = MarksCommand
	var rmtPlaceIDsByAddrs map[string][]int
	if isUniversalPN {
		rmtPlaceIDsByAddrs = remoteTransition.GetAllPlaceIDsByAddrs()
	} else {
		rmtPlaceIDsByAddrs = remoteTransition.GetPlaceIDsByAddrs() // excludes out arcs
	}
	for rmtAddr, places := range rmtPlaceIDsByAddrs {
		// places is []int
		baseMsg2 := baseMsg
		baseMsg2.RemoteArcs = make([]*petrinet.RemoteArc, len(places))
		for i, p := range places {
			baseMsg2.RemoteArcs[i] = &petrinet.RemoteArc{PlaceID: p}
		}
		var err error
		if rmtAddr == pn.node.ExternalAddress() {
			pn.petriNet.CopyPlaceMarksToRemoteArc(baseMsg2.RemoteArcs)
			for _, rmtArc := range baseMsg2.RemoteArcs {
				pn.saveMarks(rmtAddr, rmtArc.PlaceID, rmtArc) // ADDR, PLACE ID, RMT ARC
			}
			pn.verifiedRemoteAddrs = append(pn.verifiedRemoteAddrs, rmtAddr)
		} else {
			err = pn.SendMessageByAddress(baseMsg2, rmtAddr)
			if err == nil {
				connectedAddrs[rmtAddr] = true
			} else if isUniversalPN {
				// on universal pn case, ALL addresses should be available
				return connectedAddrs, fmt.Errorf("Address %v is unreachable but needed for universal remote transition %v", rmtAddr, remoteTransition)
			}
		}
	}
	return connectedAddrs, nil
}

// if transition option is not valid, remove it
func (pn *petriNode) removeTransitionOption(addrs string, transition *petrinet.Transition) {
	// delete transition from option list
	fmt.Println("WILL DELETE TRANSITION OPTION")
	fmt.Printf("Transition options before: %v\n", pn.transitionOptions)
	fmt.Printf("Remote transition options before: %v\n", pn.remoteTransitionOptions)
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
	fmt.Printf("Transition options after: %v\n", pn.transitionOptions)
	fmt.Printf("Remote transition options after: %v\n", pn.remoteTransitionOptions)
	fmt.Println("DONE DELETING TRANSITION OPTION")
}

func (pn *petriNode) saveMarks(addr string, placeID int, rmtArc *petrinet.RemoteArc) {
	placeMap, exists := pn.marks[addr]
	if !exists {
		pn.marks[addr] = make(map[int]*petrinet.RemoteArc)
		placeMap = pn.marks[addr]
	}
	placeMap[placeID] = rmtArc
}

func (pn *petriNode) saveAllMarksAndUpdateMissing(pMsg petriMessage) {
	if pMsg.Command != MarksCommand {
		fmt.Printf("Expected marks, received something else: %v HERE\n", pMsg.Command)
		pn.resetStep()
		return
	}
	for _, rmtArc := range pMsg.RemoteArcs {
		pn.saveMarks(pMsg.Address, rmtArc.PlaceID, rmtArc)
	}
	_, present := pn.addressMissing[pMsg.Address]
	if present {
		pn.verifiedRemoteAddrs = append(pn.verifiedRemoteAddrs, pMsg.Address)
		delete(pn.addressMissing, pMsg.Address)
	}
}

func (pn *petriNode) getPlaceMarks(pMsg petriMessage) {
	fmt.Println("RECEIVED A MARK MSG")
	pn.saveAllMarksAndUpdateMissing(pMsg)
	fmt.Printf("ADDRESS MISSING AFTER DELETION: %v\n", pn.addressMissing)
	if len(pn.addressMissing) == 0 {
		if !pn.validateRemoteTransitionMarks() {
			// transition wasn't ready to fire, remove from options and try again
			fmt.Println("TRANSITION WASNT READY TO FIRE, WILL GO TO PREPARE_FIRE_STEP")
			pn.removeTransitionOption(pn.chosenTransitionAddress, pn.chosenTransition)
			pn.step = PREPARE_FIRE_STEP
			// save chosenTransition priority
		} else {
			fmt.Println("TRANSITION READY TO FIRE, WILL GO TO FIRE_STEP")
			pn.incStep() // FIRE_STEP
		}
	}
}

func (pn *petriNode) requestTemporalPlacesRollback(baseMsg petriMessage) {
	pn.petriNet.RollBackTemporal()
	baseMsg.Command = RollBackTemporalPlacesCommand
	success := func () {
		fmt.Println("Requested roll back of temporal places")
	}
	timeoutCallback := func() {
		fmt.Println("ERROR: on roll back temporal of all")
	}
	pn.broadcastWithTimeout(baseMsg, success, timeoutCallback)
}

// after receiving all valid transitions, do this first and wait for askedAddrs to respond
func (pn *petriNode) prepareFire(baseMsg petriMessage) {
	fmt.Println("PRERAREFIRE METH CALLED")
	pn.needsToCheckForConflictedState = false
	pn.didFire = false
	transition, peerAddr := pn.selectTransition()
	if transition == nil {
		fmt.Println("_NO TRANSITION TO SELECT")
		pn.resetStep()
		// will retry with next priority
		if pn.priorityToAsk < pn.maxPriority {
			if pn.priorityToAsk == 0 {
				pn.requestTemporalPlacesRollback(baseMsg)
			}
			pn.priorityToAsk++
		} else {
			pn.priorityToAsk = 0
		}
	} else {
		pn.chosenTransition = transition
		pn.chosenTransitionAddress = peerAddr
		pn.verifiedRemoteAddrs = []string{}
		pn.marks = make(map[string]map[int]*petrinet.RemoteArc)
		rmtTransitionOption, ok := pn.remoteTransitionOptions[peerAddr][transition.ID]
		fmt.Printf("CHOSEN TRANSITION %v\nCHOSEN ADDR %v\n", transition, peerAddr)
		if ok {
			fmt.Println("WILL FIRE REMOTE TRANSITION")
			copy := *rmtTransitionOption // get a copy
			remoteTransition := &copy // pointer to the copy
			fmt.Printf("REMOTE TRANSITION TO FIRE B4 UPDATE BY CTX: %v\n", remoteTransition)
			pn.chosenRemoteTransition = remoteTransition
			if peerAddr != UNIVERSAL_PN {
				remoteTransition.UpdateAddressByContext(pn.contextToAddrs, peerAddr)
			}
			fmt.Printf("CTX TO ADDRS: %v\n", pn.contextToAddrs)
			fmt.Printf("REMOTE TRANSITION TO FIRE AFTER UPDATE BY CTX: %v\n", remoteTransition)
			askedAddrs, err := pn.askForMarks(remoteTransition, peerAddr == UNIVERSAL_PN, baseMsg)
			fmt.Println(askedAddrs)
			if err != nil {
				pn.removeTransitionOption(pn.chosenTransitionAddress, pn.chosenTransition)
				pn.step = PREPARE_FIRE_STEP
			} else {
				pn.addressMissing = askedAddrs
				pn.incStep() // RECEIVING_MARKS_STEP
				if len(pn.addressMissing) == 0 {
					// skip RECEIVING_MARKS_STEP
					pn.incStep() // FIRE_STEP
				}
			}
		} else {
			fmt.Println("THERE IS NO REMOTE TRANSITION TO FIRE")
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
			fmt.Printf("HERE Place %v: %v marks. Arc weight: %v.\n", currArc.PlaceID, place.Marks, currArc.Weight)
			ans = ans && comp(place.Marks, currArc.Weight)
		}
	}
	if rmtTransition != nil {
		helperFunc(rmtTransition.InArcs, func(a, b int) bool { return a >= b})
		helperFunc(rmtTransition.InhibitorArcs, func(a, b int) bool { return a < b})
	}
	return ans
}

func (pn *petriNode) fireRemoteTransition(t *petrinet.RemoteTransition, isUniversalPN bool, baseMsg petriMessage) error {
	if t != nil {
		fmt.Println("WILL FIRE REMOTE TRANSITION METH")
		addrDidSaveHistory := make(map[string]bool)
		helperFunc := func(opType petrinet.OperationType, addrToArcMap map[string][]*petrinet.RemoteArc, verifiedAddrs []string) error {
			fmt.Println("RUNNING HELPER FUNC")
			fmt.Printf("VERIFIED REMOTE ADDRS: %v\n", verifiedAddrs)
			for _, addr := range verifiedAddrs {
				baseMsg2 := baseMsg
				baseMsg2.Command = AddToPlacesCommand
				baseMsg2.RemoteArcs = addrToArcMap[addr]
				baseMsg2.OpType = opType
				fmt.Println("WILL SEE IF IT SHOULD FIRE REMOTE")
				if len(baseMsg2.RemoteArcs) > 0 {
					fmt.Println("WILL FIRE REMOTE")
					if !addrDidSaveHistory[addr] {
						baseMsg2.SaveHistory = true
						addrDidSaveHistory[addr] = true
					}
					if addr == pn.node.ExternalAddress() {
						pn.petriNet.AddMarksToPlaces(opType, baseMsg2.RemoteArcs, baseMsg2.SaveHistory)
						fmt.Println("REMOTE WAS LOCAL, FIRED IMMEDIATLY")
					} else {
						err := pn.SendMessageByAddress(baseMsg2, addr)
						fmt.Printf("SENT MSG TO ADDRES %v, Err: %v\n", addr, err)
						if isUniversalPN && err != nil {
								// For universal pn remote transitions ALL addresses should be connected, else
								// it should roll back
								fmt.Printf("ERROR MSG TO ADDRES %v: %v\n", addr, err)
								return err
						}
					}
				}
		  }
			return nil
		}
		placesToFire := t.GetInArcsByAddrs()
		placesToReceive := t.GetOutArcsByAddrs()
		fmt.Printf("REMOTE IN ARCS TO FIRE: %v\n", placesToFire)
		fmt.Printf("REMOTE OUT ARCS TO FIRE: %v\n", placesToReceive)
		err := helperFunc(petrinet.SUBSTRACTION, placesToFire, pn.verifiedRemoteAddrs)
		if err == nil {
			var verifiedOut []string
			for key := range placesToReceive {
				verifiedOut = append(verifiedOut, key) // for out I dont care if its verified, do all
			}
			err = helperFunc(petrinet.ADDITION, placesToReceive, verifiedOut)
		}
		if err != nil {
			// there was at least one address that was unreachable.
			// should roll back all the ones that did update
			pn.rollBackByAddress(addrDidSaveHistory, baseMsg)
			return fmt.Errorf("Not all adresses were connected for universal petri net, did roll back. %v", err)
		}
	}
	return nil
}

func (pn *petriNode) rollBackByAddress(addrMap map[string]bool, baseMsg petriMessage) {
	baseMsg.Command = RollBackCommand
	for peerAddr := range addrMap {
		pn.SendMessageByAddress(baseMsg, peerAddr)
	}
}

func (pn *petriNode) fireTransition(baseMsg petriMessage) error {
	if !pn.validateRemoteTransitionMarks() {
		// transition wasn't ready to fire, remove from options and try again
		pn.removeTransitionOption(pn.chosenTransitionAddress, pn.chosenTransition)
		pn.step = PREPARE_FIRE_STEP
		// save chosenTransition priority
		fmt.Println("Didnt fire because wasnt ready")
		return nil
	}
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
	} else if peerAddr != UNIVERSAL_PN {
		// Transition is from other peer
		fmt.Println("WILL SEND MSG TO FIRE TRANSITION")
		err = pn.SendMessageByAddress(baseMsg, peerAddr)
		fmt.Println("DONE  SENDING MSG TO FIRE TRANSITION")
	}
	if err == nil {
		// Transition fired with no problem
		fmt.Println("NO ERROR, WILL FIRE REMOTE TRANSITION")
		err = pn.fireRemoteTransition(remoteTransition, peerAddr == UNIVERSAL_PN, baseMsg) // Fire remote transition
		fmt.Println(err)
		if err != nil {
			pn.removeTransitionOption(pn.chosenTransitionAddress, pn.chosenTransition)
			pn.step = PREPARE_FIRE_STEP
		} else {
			pn.incStep()
			pn.priorityToAsk = 0 // reset for next iteration
		}
	} else {
		pn.resetStep()
	}
	return err
}

func (pn *petriNode) reInitPetriNode(baseMsg petriMessage) {
	pn.contextToAddrs = make(map[string][]string)
	pn.addrsToContext = make(map[string]string)
	pn.marks = make(map[string]map[int]*petrinet.RemoteArc)
	pn.updateCtx(baseMsg)
}

func (pn *petriNode) ask(baseMsg petriMessage) {
	pn.reInitPetriNode(baseMsg)
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
	if peerAddr != pn.lastMsgTo {
		msgToSend.imNew = true
		pn.lastMsgTo = peerAddr
	}
	fmt.Printf("WILL SEND: %v\n", msgToSend)
	return peer.SendMessage(msgToSend)
}

func (pn *petriNode) showPetriNetCurrentState() {
	fmt.Println("Will print petri net")
	fmt.Printf("%v\n", pn.petriNet)
	if flag.Lookup("test.v") == nil {
		time.Sleep(time.Duration(humanTimeout) * time.Millisecond)
  }
}

func (pn *petriNode) printPetriNet(baseMsg petriMessage) {
	baseMsg.Command = PrintCommand
	pn.node.Broadcast(baseMsg)
	pn.showPetriNetCurrentState()
	pn.incStep()
}

func (pn *petriNode) broadcastWithTimeout(msg petriMessage, successCallback, timeoutCallback func()) {
	errChan := make(chan []error)
	defer close(errChan)
	go func() {
		fmt.Printf("Doing broadcast of %v...\n", msg)
		// err := skademlia.Broadcast(pn.node, msg)
		err := pn.node.Broadcast(msg)
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
		baseMsg.Transitions, baseMsg.RemoteTransitions = pn.petriNet.GetTransitionOptionsByPriority(pMsg.AskedPriority)
		baseMsg.AskedPriority = pn.petriNet.GetMaxPriority()
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
		pn.petriNet.AddMarksToPlaces(pMsg.OpType, pMsg.RemoteArcs, pMsg.SaveHistory)
	case RollBackTemporalPlacesCommand:
		err := pn.petriNet.RollBackTemporal()
		if err != nil {
			fmt.Printf("Tried to roll back more, got: %v\n", err)
		}
	case RollBackCommand:
		err := pn.petriNet.RollBack()
		if err != nil {
			fmt.Printf("Tried to roll back more, got: %v\n", err)
		}
	default:
		fmt.Printf("Unknown command: %v\n", pMsg.Command)
	}
}

func (pn *petriNode) updateCtx(pMsg petriMessage) {
	fmt.Printf("_WILL UPDATE CTX WITH: %v\n", pMsg)
	fmt.Printf("_ctx b4: %v\n", pn.contextToAddrs)
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
	fmt.Printf("_ctx after: %v\n", pn.contextToAddrs)
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

func (pn *petriNode) updateNeedsToCheckForConflictedState(pMsg petriMessage) {
	pn.needsToCheckForConflictedState = pn.needsToCheckForConflictedState || pMsg.imNew
	pn.didFire = pn.didFire || (pMsg.imNew && pMsg.iveBeenFired)
}

func (pn *petriNode) checkConflictedStep(baseMsg petriMessage) {
		var placesToAskByAddress map[string][]int
		var connectedAddrs map[string]bool
		placesToAskByAddress = pn.getPossibleConflictPlacesByAddress()
		for addr, places := range placesToAskByAddress {
			msgCopy := baseMsg
			msgCopy.RemoteArcs = make([]*petrinet.RemoteArc, len(places))
			for i, p := range places {
				msgCopy.RemoteArcs[i] = &petrinet.RemoteArc{PlaceID: p}
			}
			var err error
			if addr == pn.node.ExternalAddress() {
				pn.petriNet.CopyPlaceMarksToRemoteArc(msgCopy.RemoteArcs)
				for _, rmtArc := range msgCopy.RemoteArcs {
					pn.saveMarks(addr, rmtArc.PlaceID, rmtArc) // ADDR, PLACE ID, RMT ARC
				}
				pn.verifiedRemoteAddrs = append(pn.verifiedRemoteAddrs, addr)
			} else {
				err = pn.SendMessageByAddress(msgCopy, addr)
				if err == nil {
					connectedAddrs[addr] = true
				}
			}
		}
		pn.addressMissing = connectedAddrs
}

func (pn *petriNode) getPlaceConflictedMarks(pMsg petriMessage, baseMsg petriMessage) {
	 pn.saveAllMarksAndUpdateMissing(pMsg)
	 if len(pn.addressMissing) == 0 {
		 // check if conflict.
		 conflictedAddrs := pn.getConflictedAddrs()
		 if len(conflictedAddrs) > 0 {
			 pn.rollBackByAddress(conflictedAddrs, baseMsg)
			 pn.step = CHECK_CONFLICTED_STEP
		 } else {
			 pn.incStep()
			 pn.resetStep() // everything ok, should start from ask
			 pn.didFire = false
			 pn.needsToCheckForConflictedState = false
		 }
	 }
}

func (pn *petriNode) getConflictedAddrs() map[string]bool {
	// TODO: complete.
	return nil
}

func (pn *petriNode) getPossibleConflictPlacesByAddress() map[string][]int {
	// TODO: Complete.
	return nil
}
