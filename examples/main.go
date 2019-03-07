package main

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"flag"
	"fmt"
	// "math"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/FLAGlab/DCoPN/petrinet"
	"github.com/perlin-network/noise"
	"github.com/perlin-network/noise/cipher/aead"
	"github.com/perlin-network/noise/handshake/ecdh"
	"github.com/perlin-network/noise/log"
	"github.com/perlin-network/noise/payload"
	"github.com/perlin-network/noise/protocol"
	"github.com/perlin-network/noise/skademlia"
	"github.com/pkg/errors"
)

// CommandType enums for PetriNodes communication
type CommandType string

const (
	// TransitionCommand to query transitions
	TransitionCommand  CommandType = "transitions"
	// FireCommand to activate a fire event on the PetriNet
	FireCommand        CommandType = "fire"
	// PrintCommand to print the current state of the PetriNet
	PrintCommand       CommandType = "print"
)

type petriNode struct {
	node *noise.Node
	petriNet *petrinet.PetriNet
	isLeader bool
	transitionOptions map[string][]petrinet.Transition
	mux sync.Mutex
	step int
	running bool
}

type petriMessage struct {
	Command CommandType
	Address string
	Transitions []petrinet.Transition
}

func (pn *petriNode) incStep() {
	pn.mux.Lock()
	defer pn.mux.Unlock()
	pn.step = (pn.step + 1) % 5
}

func (pn *petriNode) initTransitionOptions() {
	pn.mux.Lock()
	defer pn.mux.Unlock()
	pn.transitionOptions = make(map[string][]petrinet.Transition)
	pn.transitionOptions[pn.node.ExternalAddress()] = pn.petriNet.GetTransitionOptions()
}

func (pn *petriNode) addTransitionOption(key string, options []petrinet.Transition) int {
	pn.mux.Lock()
	defer pn.mux.Unlock()
	pn.transitionOptions[key] = options
	return len(pn.transitionOptions)
}

func (petriMessage) Read(reader payload.Reader) (noise.Message, error) {
	byts, err := reader.ReadBytes()
	if err != nil {
		return nil, errors.Wrap(err, "failed to read msg")
	}
	var m petriMessage
	dec := gob.NewDecoder(bytes.NewReader(byts))
	if err := dec.Decode(&m); err != nil {
		return nil, errors.Wrap(err, "failed to decode msg")
	}

	return m, nil
}

func (m petriMessage) Write() []byte {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(m); err != nil {
		log.Info().Msgf("Got a fucking error: %v", err)
	}
	return payload.NewWriter(nil).WriteBytes(buf.Bytes()).Bytes()
}

func buildPetriNet() *petrinet.PetriNet {
	p := petrinet.Init(1)
	p.AddPlace(1, 1, "")
	p.AddPlace(2, 1, "")
	p.AddPlace(3, 2, "")
	p.AddPlace(4, 1, "")
	p.AddTransition(1,1)
	p.AddTransition(2,0)
	p.AddInArc(1,1,1)
	p.AddInArc(2,2,1)
	p.AddInArc(3,2,1)
	p.AddOutArc(1,2,1)
	p.AddOutArc(1,3,1)
	p.AddOutArc(2,4,1)
  //p.AddInhibitorArc(4,2,1)
	// fmt.Printf("%v\n", p)
	return p
}

func (pn *petriNode) ask() {
	node := pn.node
	skademlia.BroadcastAsync(
		node, petriMessage{Command: TransitionCommand, Address: node.ExternalAddress()})
}

// func (pn *petriNode) wait(opcode noise.Opcode) ([]petrinet.Transition, error) {
// 	peers := skademlia.Table(pn.node).GetPeers()
// 	transitionOptions := pn.petriNet.GetTransitionOptions()
// 	currMin := math.MaxInt64
// 	for _, peer := range peers {
// 		msg := <-peer.Receive(opcode)
// 		pMsg := msg.(petriMessage)
// 		if pMsg.Command != TransitionCommand {
// 			return nil, errors.New("Expected transition, received something else")
// 		}
// 		currTrans := pMsg.Transitions
// 		log.Info().Msgf("Will process transitions " + strconv.Itoa(i))
// 		if len(currTrans) > 0 && currTrans[0].Priority < currMin {
// 			currMin = currTrans[0].Priority
// 			transitionOptions = []petrinet.Transition{currTrans[0]}
// 		} else if (currTrans[0].Priority == currMin) {
// 			transitionOptions = append(transitionOptions, currTrans[0])
// 		}
// 	}
// 	fmt.Printf("Done waiting, got options: %v\n", transitionOptions)
// 	return transitionOptions, nil
// }

/** ENTRY POINT **/
func setup(pn *petriNode) {
	opcodeChat := noise.RegisterMessage(noise.NextAvailableOpcode(), (*petriMessage)(nil))
	node := pn.node
	node.OnPeerInit(func(node *noise.Node, peer *noise.Peer) error {
		// init se llama cuando se conecta un nodo o se le hace dial
		peer.OnConnError(func(node *noise.Node, peer *noise.Peer, err error) error {
			log.Info().Msgf("Got an error: %v", err)
			return nil
		})

		peer.OnDisconnect(func(node *noise.Node, peer *noise.Peer) error {
			log.Info().Msgf("Peer %v has disconnected.", peer.RemoteIP().String()+":"+strconv.Itoa(int(peer.RemotePort())))
			return nil
		})

		if pn.isLeader {
			time.Sleep(10 * time.Second)
		}
		if !pn.running {
			pn.running = true
			// acá solo se comunica con el peer que se acaba de inicializar
			go func() {
				for i:=0; i<100; i++ {
					fmt.Println("WILL PROCESS")
					if pn.isLeader {
						switch pn.step {
						case 0:
							fmt.Println("WILL ASK")
							pn.ask()
							pn.initTransitionOptions()
							pn.incStep()
						case 1:
							fmt.Println("WILL WAIT")
							msg := <-peer.Receive(opcodeChat)
							pMsg := msg.(petriMessage)
							fmt.Printf("Received msg %v\n", pMsg)
							if pMsg.Command != TransitionCommand {
								fmt.Println("Expected transition, received something else")
								pn.step = 0
							}
							fmt.Printf("Received options %v\n", pMsg.Transitions)
							numDone := pn.addTransitionOption(pMsg.Address, pMsg.Transitions)
							expected := len(skademlia.Table(node).GetPeers())
							fmt.Printf("Done with: %v Expected: %v\n", numDone, expected)
							if numDone == expected {
								pn.incStep()
							}
						case 3:
							fmt.Println("WILL FIRE")
							pn.incStep()
						case 4:
							fmt.Println("WILL PRINT")
							pn.incStep()
						}

						// fmt.Println("WILL WAIT")
						// transitions, err := pn.wait(opcodeChat)
						// if err != nil {
						// 		log.Info().Msgf("error on transitions: %v", err)
						// 		continue
						// }
						// log.Info().Msgf("transitions: %v", transitions)
						// //selectAndFire(transitions,node)
						// log.Info().Msgf("a petri net: %v", pn)
						//msg := <-peer.Receive(opcodeChat)
						//log.Info().Msgf("[%s]: %s, %s", protocol.PeerID(peer), msg.(petriMessage).Command,msg.(petriMessage).Address)
					} else {
						fmt.Println("WAITING")
						msg := <-peer.Receive(opcodeChat)
						pMsg := msg.(petriMessage)
						fmt.Printf("RECEIVED: %v\n", pMsg)
						fmt.Println("Will dial...")
						remotePeer, err := node.Dial(pMsg.Address) // hace que se llame otra vez init
						if err != nil {
							panic(err)
						}
						fmt.Printf("Dial ok, %v\n", remotePeer)
						switch pMsg.Command {
						case TransitionCommand:
							fmt.Println("Type transition")
							transitionOptions := pn.petriNet.GetTransitionOptions()
							msgToSend := petriMessage{
								Command: TransitionCommand,
								Address: node.ExternalAddress(),
								Transitions: transitionOptions}
							fmt.Printf("will send %v\n", msgToSend)
							remotePeer.SendMessageAsync(msgToSend)
							fmt.Printf("sent transitions: %v\n", transitionOptions)
						case FireCommand:
							fmt.Println("fire")
						case PrintCommand:
							fmt.Println("print")
						default:
							fmt.Println("Unknown command")
						}
					}
					time.Sleep(5 * time.Second)
					/*log.Info().Msgf("[%s]: %s, %s", protocol.PeerID(peer), msg.(petriMessage).Command,msg.(petriMessage).Address)
					if msg.(petriMessage).Transtion < 5 {
						peer, err := node.Dial(msg.(petriMessage).Address)
						if err != nil {
							panic(err)
						}
						peer.SendMessageAsync(petriMessage{Command: "jajajja", Address: node.ExternalAddress(), Transtion: msg.(petriMessage).Transtion+1 })
					}*/

				}
			}()
		}
		return nil
	})
}

func main() {
	//gob.Register(skademlia.ID{})
	hostFlag := flag.String("h", "127.0.0.1", "host to listen for peers on")
	portFlag := flag.Uint("p", 3000, "port to listen for peers on")
	leaderFlag := flag.Bool("l", false, "is leader node")
	flag.Parse()

	params := noise.DefaultParams()
	//params.NAT = nat.NewPMP()
	params.Keys = skademlia.RandomKeys()
	params.Host = *hostFlag
	params.Port = uint16(*portFlag)

	node, err := noise.NewNode(params)
	if err != nil {
		panic(err)
	}
	defer node.Kill()

	pnNode := &petriNode{node: node, petriNet: buildPetriNet(), isLeader: *leaderFlag}

	p := protocol.New()
	p.Register(ecdh.New())
	p.Register(aead.New())
	p.Register(skademlia.New())
	p.Enforce(node)
	setup(pnNode)
	go node.Listen()

	log.Info().Msgf("Listening for peers on port %d.", node.ExternalPort())

	if len(flag.Args()) > 0 {
		for _, address := range flag.Args() {
			peer, err := node.Dial(address)
			if err != nil {
				panic(err)
			}

			skademlia.WaitUntilAuthenticated(peer)
		}

		peers := skademlia.FindNode(node, protocol.NodeID(node).(skademlia.ID), skademlia.BucketSize(), 8)
		log.Info().Msgf("Bootstrapped with peers: %+v", peers)
	}
	reader := bufio.NewReader(os.Stdin)
	//if *leaderFlag {

	//}
	for {
		txt, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		log.Info().Msgf("read %v.", txt)

	}
/*

	for {
		txt, err := reader.ReadString('\n')

		if err != nil {
			panic(err)
		}
		x := 0
		txt = strings.TrimSpace(txt)
		for _, peerID := range skademlia.FindClosestPeers(skademlia.Table(node), protocol.NodeID(node).Hash(), skademlia.BucketSize()) {
				peer := protocol.Peer(node, peerID)

				if peer == nil {
					continue
*///				}

	//			peer.SendMessageAsync(petriMessage{Command: txt, Address: node.ExternalAddress(), Transtion: x})
	//				//peer.SendMessageAsync(petriMessage{Command: txt, Transtion: x})
	//			x = x+1
	//	}
		//skademlia.BroadcastAsync(node, petriMessage{text: })
	//}
}
