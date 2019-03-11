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

// MinTimeout the minimum timeout for raft
const MinTimeout = 1000 // milliseconds
// MaxTimeout the max timeout for raft
const MaxTimeout = 5000 // milliseconds
// LeaderTimeout to wait for ask response
const LeaderTimeout = 500 // milliseconds
const humanTimeout = 5000

// NodeType enums for PetriNodes communication
type NodeType string

const (
	// Leader is leader node
	Leader    NodeType = "leader"
	// Follower is follower node
	Follower  NodeType = "follower"
	// Candidate is candidate node
	Candidate NodeType = "candidate"
)

type petriNode struct {
	node *noise.Node
	petriNet *petrinet.PetriNet
	nodeType NodeType
	transitionOptions map[string][]*petrinet.Transition
	mux sync.Mutex
	step int
	currentTerm int
	timeoutCount int
	pMsg chan petriMessage
	myVotes map[string]string
	votedFor string
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
		fmt.Println("Expected transition, received something else")
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
	pnNodeIndex := rand.Intn(initial)
	chosenKey := indexToKey[pnNodeIndex]
	options := pn.transitionOptions[chosenKey]
	transitionIndex := rand.Intn(len(options))
	return options[transitionIndex], chosenKey
}

func (pn *petriNode) fireTransition() error {
	transition, peerAddr := pn.selectTransition()
	if transition == nil {
		fmt.Println("NO TRANSITION TO SELECT")
		return nil
	}
	// fmt.Printf("SELECTED TRANSITION: %v\n", transition)
	// fmt.Printf("Of peer: %v\n", peerAddr)
	msgToSend := pn.generateMessage(FireCommand)
	msgToSend.Transitions = []*petrinet.Transition{transition}
	if peerAddr == pn.node.ExternalAddress() {
		pn.petriNet.FireTransitionByID(transition.ID)
		return nil
	}
	// fmt.Printf("TRANSITION IS REMOTE: %v\n", peerAddr)
	return pn.SendMessageByAddress(msgToSend, peerAddr)
}

func (pn *petriNode) SendMessageByAddress(msgToSend petriMessage, peerAddr string) error {
	peer, err := pn.node.Dial(peerAddr)
	if err != nil {
		fmt.Printf("Error dialing: %v\n", peerAddr)
		return err
	}
	return peer.SendMessage(msgToSend)
}

func (pn *petriNode) printPetriNet() {
	fmt.Printf("%v\n", pn.petriNet)
	skademlia.Broadcast(pn.node, pn.generateMessage(PrintCommand))
}

func (pn *petriNode) assembleElection() {
	pn.setNodeType(Candidate)
	pn.myVotes[pn.node.ExternalAddress()] = pn.node.ExternalAddress()
	pn.currentTerm++
	pn.votedFor = pn.node.ExternalAddress()
	timeoutCallback := func () {
		pn.setNodeType(Candidate)
	}
	pn.broadcastWithTimeOut(pn.generateMessage(RequestVoteCommand), func(){}, timeoutCallback)
}

func (pn *petriNode) broadcastWithTimeOut(msg petriMessage, successCallback, timeoutCallback func()) {
	errChan := make(chan []error)
	defer close(errChan)
	go func() {
		err := skademlia.Broadcast(
			pn.node,
			msg)
		errChan <- err
	}()
	select {
	case errList := <- errChan:
		if len(errList) > 0 {
			timeoutCallback()
		} else {
			successCallback()
		}
	case <- time.After(time.Duration(pn.timeoutCount + humanTimeout) * time.Millisecond):
		timeoutCallback()
	}
}

func (pn *petriNode) run() {
	go func() {
		time.Sleep(time.Duration(humanTimeout) * time.Millisecond)
		for  {
			fmt.Printf("LISTENING TO MESSAGES AS %v\n", pn.nodeType)
			fmt.Printf("Will wait for %v millis\n", pn.timeoutCount + humanTimeout)
			if pn.nodeType == Leader {
				pn.processLeader()
			} else if pn.nodeType == Follower {
				select {
				case pMsg := <- pn.pMsg:
					pn.processFollower(pMsg)
				case <- time.After(time.Duration(pn.timeoutCount + humanTimeout) * time.Millisecond):
					// anarchy!!
					fmt.Println("Will do election")
					pn.assembleElection()
				}
			} else if pn.nodeType == Candidate {
				// TODO si hay solo un nodo se queda trabado
				select {
				case pMsg := <- pn.pMsg:
					pn.processCandidate(pMsg)
				case <- time.After(time.Duration(pn.timeoutCount + humanTimeout) * time.Millisecond):
					// vote time is over... lets try again
					pn.setNodeType(Candidate) // to reset vote counts
					pn.assembleElection()
				}
			}
			time.Sleep(time.Duration(humanTimeout) * time.Millisecond)
		}
	}()
}

func (pn *petriNode) processCandidate(pMsg petriMessage) {
	fmt.Printf("Processing msg as candidate: %v", pMsg)
	if pMsg.FromType == Leader || pMsg.Term > pn.currentTerm{
		// theres a leader !!
		pn.setNodeType(Follower)
		if pMsg.Command != VoteCommand { // in case it is not a leader
			pn.processFollower(pMsg)
		}
	} else if pMsg.Command == RequestVoteCommand { // someone else wants me to vote
		fmt.Println("someone else wants my vote D:")
		pn.vote(pMsg.Term, pMsg.Address)
	} else { // its a vote
		fmt.Printf("Received %v vote from: %v\n", pMsg.VoteGranted, pMsg.Address)

		pn.myVotes[pMsg.Address] = pMsg.VoteGranted
		total := len(skademlia.Table(pn.node).GetPeers()) + 1 // plus me
		fmt.Printf("Total of votes: %v\n", pn.myVotes)
		fmt.Printf("Total of peers: %v\n", total)
		if len(pn.myVotes) == total { // polls are closed!
			fmt.Println("POLLS ARE CLOSED!!!")
			countMap := make(map[string]int)
			maxVotes := 0
			maxVoteAddress := ""
			for _, voteAddr := range pn.myVotes {
				countMap[voteAddr]++
				if countMap[voteAddr] > maxVotes {
					maxVotes = countMap[voteAddr]
					maxVoteAddress = voteAddr
				}
			}
			fmt.Printf("WINNER: %v, COUNT: %v\n", maxVoteAddress, maxVotes)
			if maxVoteAddress == pn.node.ExternalAddress() { // I won!!
				fmt.Println("LEADER SETTED AS ME !!! >:v")
				pn.setNodeType(Leader)
			} else {
				pn.setNodeType(Follower)
			}
		}
	}
}

func (pn *petriNode) processFollower(pMsg petriMessage) {
	fmt.Printf("Received msg: %v\n", pMsg)
	if pMsg.Term >= pn.currentTerm {
		pn.currentTerm = pMsg.Term

		switch pMsg.Command {
		case TransitionCommand:
			transitionOptions := pn.petriNet.GetTransitionOptions()
			msgToSend := pn.generateMessage(TransitionCommand)
			msgToSend.Transitions = transitionOptions
			pn.SendMessageByAddress(msgToSend, pMsg.Address)
		case FireCommand:
			transitionID := pMsg.Transitions[0].ID
			fmt.Printf("WILL FIRE transition with id: %v\n", transitionID)
			err := pn.petriNet.FireTransitionByID(transitionID)
			if err != nil {
				fmt.Println(err)
			}
		case PrintCommand:
			fmt.Println("CURRENT PETRI NET:")
			fmt.Printf("%v\n", pn.petriNet)
		case RequestVoteCommand:
			fmt.Println("WILL VOTE")
			pn.vote(pMsg.Term, pMsg.Address)
		default:
			fmt.Printf("Unknown command: %v\n", pMsg.Command)
		}
		if pMsg.Command != RequestVoteCommand {
			pn.votedFor = "" // theres a leader, I'll be ready for new elections TODO revisar
		}
	}
}

func (pn *petriNode) vote(term int, address string) {
	ans := pn.generateMessage(VoteCommand)
	fmt.Println("WILL VOTE")
	fmt.Printf("my term: %v, msg term: %v, my last vote for: %v\n", pn.currentTerm, term, pn.votedFor)
	if pn.currentTerm < term && (pn.votedFor == "" || pn.votedFor == address) {
		ans.VoteGranted =  address
		pn.currentTerm = term
		pn.votedFor = address
	} else {
		ans.VoteGranted = pn.votedFor
	}
	fmt.Printf("My vote: %v\n", ans.VoteGranted)
	pn.SendMessageByAddress(ans, address)
}

func (pn *petriNode) processLeader() {
	// print and ask work like heartbeats
	switch pn.step {
	case 0:
		pn.ask()
	case 1:
		select {
		case msg := <- pn.pMsg:
			if (msg.Command == RequestVoteCommand || msg.FromType == Leader) && msg.Term > pn.currentTerm {
				pn.setNodeType(Follower)
			} else {
				pn.getTransition(msg)
			}
		case <-time.After(time.Duration(pn.timeoutCount + humanTimeout) * time.Millisecond):
			pn.resetStep() // ask again
		}
	case 2:
		err := pn.fireTransition()
		if err != nil {
			fmt.Println("Error trying to send fire")
			pn.resetStep()
		} else {
			pn.incStep()
		}
	case 3:
		pn.printPetriNet()
		pn.incStep()
	}
}

func (pn *petriNode) setNodeType(nodeType NodeType) {
	pn.nodeType = nodeType
	currTimeout := LeaderTimeout
	if nodeType != Leader {
		currTimeout = MinTimeout + rand.Intn(MaxTimeout - MinTimeout)
	}
	pn.myVotes = make(map[string]string)
	pn.timeoutCount = currTimeout
}

func (pn *petriNode) init(isLeader bool) {
	if isLeader {
		pn.setNodeType(Leader)
	} else {
		pn.setNodeType(Follower)
	}
	pn.pMsg = make(chan petriMessage)
}

func (pn *petriNode) close() {
	close(pn.pMsg)
}

func (pn *petriNode) generateMessage(command CommandType) petriMessage {
	return petriMessage {
		Command: command,
		Address: pn.node.ExternalAddress(),
		Term: pn.currentTerm,
		FromType: pn.nodeType}
}

func (pn *petriNode) ask() {
	success := func() {
		pn.initTransitionOptions()
		pn.incStep()
	}
	timeoutCallback := func() {
		pn.resetStep()
	}
	pn.broadcastWithTimeOut(pn.generateMessage(TransitionCommand), success, timeoutCallback)
}
